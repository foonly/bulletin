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
	ID        uuid.UUID `json:"id"`
	CircleID  uuid.UUID `json:"circle_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "chat", "join", "leave"
	CreatedAt time.Time `json:"created_at"`
}

type Client struct {
	hub      *Hub
	circleID uuid.UUID
	userID   uuid.UUID
	username string
	conn     *websocket.Conn
	send     chan Message
}

type Hub struct {
	DB         *pgxpool.Pool
	circles    map[uuid.UUID]map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub(db *pgxpool.Pool) *Hub {
	return &Hub{
		DB:         db,
		circles:    make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.circles[client.circleID] == nil {
				h.circles[client.circleID] = make(map[*Client]bool)
			}
			h.circles[client.circleID][client] = true

			// Get all currently online users in this circle
			var onlineIDs []uuid.UUID
			seen := make(map[uuid.UUID]bool)
			for c := range h.circles[client.circleID] {
				if !seen[c.userID] {
					onlineIDs = append(onlineIDs, c.userID)
					seen[c.userID] = true
				}
			}
			h.mu.Unlock()

			// Send the list of online users to the new client
			userList, _ := json.Marshal(map[string]interface{}{
				"type": "presence",
				"ids":  onlineIDs,
			})
			client.conn.WriteMessage(websocket.TextMessage, userList)

			// Broadcast join
			h.broadcast <- Message{
				CircleID:  client.circleID,
				UserID:    client.userID,
				Username:  client.username,
				Type:      "join",
				CreatedAt: time.Now(),
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.circles[client.circleID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.circles, client.circleID)
					}

					// Broadcast leave
					h.broadcast <- Message{
						CircleID:  client.circleID,
						UserID:    client.userID,
						Username:  client.username,
						Type:      "leave",
						CreatedAt: time.Now(),
					}
				}
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			if clients, ok := h.circles[message.CircleID]; ok {
				for client := range clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	circleIDStr := chi.URLParam(r, "circleID")
	circleID, err := uuid.Parse(circleIDStr)
	if err != nil {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v\n", err)
		return
	}

	var username string
	_ = h.DB.QueryRow(r.Context(), "SELECT username FROM users WHERE id = $1", userID).Scan(&username)

	client := &Client{hub: h, circleID: circleID, userID: userID, username: username, conn: conn, send: make(chan Message, 256)}
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
			Content string `json:"content"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Save to DB
		var chatMsg Message
		err = c.hub.DB.QueryRow(context.Background(),
			`INSERT INTO chat_messages (circle_id, user_id, content)
			 VALUES ($1, $2, $3)
			 RETURNING id, created_at`,
			c.circleID, c.userID, msg.Content).Scan(&chatMsg.ID, &chatMsg.CreatedAt)

		if err != nil {
			log.Printf("DB error: %v\n", err)
			continue
		}

		chatMsg.CircleID = c.circleID
		chatMsg.UserID = c.userID
		chatMsg.Username = c.username
		chatMsg.Content = msg.Content
		chatMsg.Type = "chat"

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

	var messages []Message
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
