package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For development
	},
}

type Message struct {
	ID        uuid.UUID   `json:"id,omitempty"`
	CircleID  uuid.UUID   `json:"circle_id,omitempty"`
	UserID    uuid.UUID   `json:"user_id,omitempty"`
	Username  string      `json:"username,omitempty"`
	Content   string      `json:"content,omitempty"`
	Type      string      `json:"type"` // "chat", "join", "leave", "presence", "unread_update"
	OnlineIDs []uuid.UUID `json:"online_ids,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
}

type Client struct {
	hub      *Hub
	userID   uuid.UUID
	username string
	conn     *websocket.Conn
	send     chan Message
}

type Hub struct {
	DB          *pgxpool.Pool
	clients     map[*Client]bool
	userClients map[uuid.UUID]map[*Client]bool
	broadcast   chan Message
	register    chan *Client
	unregister  chan *Client
	mu          sync.Mutex
}

func NewHub(db *pgxpool.Pool) *Hub {
	return &Hub{
		DB:          db,
		clients:     make(map[*Client]bool),
		userClients: make(map[uuid.UUID]map[*Client]bool),
		broadcast:   make(chan Message, 1024),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if h.userClients[client.userID] == nil {
				h.userClients[client.userID] = make(map[*Client]bool)
			}
			h.userClients[client.userID][client] = true

			// Get all currently online users globally
			var onlineIDs []uuid.UUID
			for uid := range h.userClients {
				onlineIDs = append(onlineIDs, uid)
			}
			h.mu.Unlock()

			// Send the list of online users to the new client
			select {
			case client.send <- Message{
				Type:      "presence",
				OnlineIDs: onlineIDs,
			}:
			default:
			}

			// Broadcast join globally
			go h.Broadcast(Message{
				UserID:    client.userID,
				Username:  client.username,
				Type:      "join",
				CreatedAt: time.Now(),
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if clients, ok := h.userClients[client.userID]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.userClients, client.userID)
					}
				}
				close(client.send)

				// Broadcast leave globally
				msg := Message{
					UserID:    client.userID,
					Username:  client.username,
					Type:      "leave",
					CreatedAt: time.Now(),
				}
				h.mu.Unlock()
				go h.Broadcast(msg)
			} else {
				h.mu.Unlock()
			}
		case message := <-h.broadcast:
			h.mu.Lock()
			var targetClients []*Client

			// Logic for determining targets:
			// 1. If Type is presence/join/leave, it's global.
			// 2. If CircleID is set, it goes to all circle members.
			// 3. If UserID is set (and no CircleID), it's targeted to a specific user.

			if message.Type == "presence" || message.Type == "join" || message.Type == "leave" {
				for c := range h.clients {
					targetClients = append(targetClients, c)
				}
			} else if message.CircleID != uuid.Nil {
				// Circle-wide broadcast
				// We need to release the lock while querying the DB to avoid deadlocks
				// or just blocking the hub for too long.
				h.mu.Unlock()
				var uids []uuid.UUID
				rows, err := h.DB.Query(context.Background(), "SELECT user_id FROM circle_members WHERE circle_id = $1", message.CircleID)
				if err == nil {
					for rows.Next() {
						var uid uuid.UUID
						if err := rows.Scan(&uid); err == nil {
							uids = append(uids, uid)
						}
					}
					rows.Close()
				}
				h.mu.Lock()
				for _, uid := range uids {
					if clients, ok := h.userClients[uid]; ok {
						for c := range clients {
							targetClients = append(targetClients, c)
						}
					}
				}
			} else if message.UserID != uuid.Nil {
				// Targeted to a specific user
				if clients, ok := h.userClients[message.UserID]; ok {
					for c := range clients {
						targetClients = append(targetClients, c)
					}
				}
			}

			for _, client := range targetClients {
				select {
				case client.send <- message:
				default:
					// If send buffer is full, we could close and unregister,
					// but for now we just skip to avoid blocking the hub.
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		// If not in context, check query param (for desktop/cross-origin clients)
		token := r.URL.Query().Get("token")
		if token != "" {
			var userID uuid.UUID
			var expiresAt time.Time
			var mfaPending bool
			err := h.DB.QueryRow(r.Context(), "SELECT user_id, expires_at, mfa_pending FROM sessions WHERE token = $1", token).Scan(&userID, &expiresAt, &mfaPending)
			if err == nil && !mfaPending && !time.Now().After(expiresAt) {
				userIDVal = userID
			}
		}
	}

	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := userIDVal.(uuid.UUID)

	log.Printf("User %s connecting to global WS\n", userID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v\n", err)
		return
	}

	var username string
	_ = h.DB.QueryRow(r.Context(), "SELECT username FROM users WHERE id = $1", userID).Scan(&username)

	client := &Client{hub: h, userID: userID, username: username, conn: conn, send: make(chan Message, 256)}
	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			CircleID uuid.UUID `json:"circle_id"`
			Content  string    `json:"content"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.CircleID == uuid.Nil {
			continue
		}

		// Save to DB
		var chatMsg Message
		err = c.hub.DB.QueryRow(context.Background(),
			`INSERT INTO chat_messages (circle_id, user_id, content)
			 VALUES ($1, $2, $3)
			 RETURNING id, created_at`,
			msg.CircleID, c.userID, msg.Content).Scan(&chatMsg.ID, &chatMsg.CreatedAt)

		if err != nil {
			log.Printf("DB error: %v\n", err)
			continue
		}

		chatMsg.CircleID = msg.CircleID
		chatMsg.UserID = c.userID
		chatMsg.Username = c.username
		chatMsg.Content = msg.Content
		chatMsg.Type = "chat"

		// Update read marker for the sender
		_, _ = c.hub.DB.Exec(context.Background(),
			`INSERT INTO read_markers (user_id, entity_id, last_read_at)
			 VALUES ($1, $2, NOW())
			 ON CONFLICT (user_id, entity_id) DO UPDATE SET last_read_at = NOW()`,
			c.userID, msg.CircleID)

		c.hub.broadcast <- chatMsg
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteJSON(message)
		}
	}
}

func (h *Hub) GetHistory(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT m.id, m.circle_id, m.user_id, u.username, m.content, m.created_at
		 FROM chat_messages m
		 JOIN users u ON m.user_id = u.id
		 WHERE m.circle_id = $1
		 ORDER BY m.created_at ASC`, circleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.CircleID, &m.UserID, &m.Username, &m.Content, &m.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, m)
	}

	json.NewEncoder(w).Encode(messages)
}
