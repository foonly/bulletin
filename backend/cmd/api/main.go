package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/bulletin/backend/internal/auth"
	"github.com/username/bulletin/backend/internal/chat"
	"github.com/username/bulletin/backend/internal/posts"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://bulletin:bulletin_password@localhost:5432/bulletin?sslmode=disable"
	}

	var pool *pgxpool.Pool
	var err error

	// Retry connecting to DB
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(context.Background(), dbURL)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				break
			}
		}
		log.Printf("Waiting for database... (%d/10)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Unable to connect to database after retries: %v\n", err)
	}
	defer pool.Close()

	// Run migrations
	err = runMigrations(pool)
	if err != nil {
		log.Printf("Migration warning: %v\n", err)
	}

	// Bootstrap initial data if empty
	bootstrap(pool)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(auth.SessionMiddleware(pool))

	authHandler := auth.NewHandler(pool)
	postHandler := posts.NewHandler(pool)
	chatHub := chat.NewHub(pool)
	go chatHub.Run()

	// Start background chat cleanup worker (runs every hour)
	cleanupWorker := chat.NewCleanupWorker(pool)
	cleanupWorker.Start(1 * time.Hour)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth)
			r.Get("/auth/me", authHandler.Me)
			r.Put("/auth/me", authHandler.UpdateMe)
			r.Post("/circles", postHandler.CreateCircle)

			r.Get("/circles", postHandler.ListCircles)
			r.Route("/circles/{circleID}", func(r chi.Router) {
				r.Use(postHandler.MembershipMiddleware)
				r.Put("/", postHandler.UpdateCircle)
				r.Get("/threads", postHandler.ListThreads)
				r.Get("/threads/{postID}", postHandler.GetThread)
				r.Put("/threads/{postID}", postHandler.UpdatePost)
				r.Delete("/threads/{postID}", postHandler.DeletePost)
				r.Post("/read/{entityID}", postHandler.UpdateReadMarker)
				r.Get("/posts", postHandler.ListPosts)
				r.Post("/posts", postHandler.CreatePost)
				r.Get("/members", postHandler.ListMembers)
				r.Put("/members/{userID}", postHandler.UpdateMember)
				r.Delete("/members/{userID}", postHandler.DeleteMember)
				r.Get("/tags", postHandler.ListTags)
				r.Post("/tags", postHandler.CreateTag)
				r.Post("/tags/{tagID}/pin", postHandler.PinTag)
				r.Get("/invites", postHandler.ListInvites)
				r.Post("/invites", postHandler.CreateInvite)
				r.Delete("/invites/{inviteID}", postHandler.DeleteInvite)
				r.Get("/chat/ws", chatHub.HandleWS)
				r.Get("/chat/history", chatHub.GetHistory)
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func bootstrap(pool *pgxpool.Pool) {
	var count int
	err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM circles").Scan(&count)
	if err != nil || count > 0 {
		return
	}

	fmt.Println("Bootstrapping initial circle and invite...")
	// Create a "System" user if needed, or just a circle.
	// We'll create a circle and a one-time use invite code 'welcome'.
	var circleID uuid.UUID
	err = pool.QueryRow(context.Background(), "INSERT INTO circles (name, description) VALUES ($1, $2) RETURNING id", "Default Circle", "The first circle").Scan(&circleID)
	if err != nil {
		log.Printf("Bootstrap error: %v\n", err)
		return
	}

	_, err = pool.Exec(context.Background(), "INSERT INTO invites (code, circle_id, role_to_grant, max_uses) VALUES ($1, $2, $3, $4)", "welcome", circleID, "admin", 1)
	if err != nil {
		log.Printf("Bootstrap error: %v\n", err)
		return
	}
	fmt.Println("Bootstrap complete. Use invite code 'welcome' to register.")
}

func runMigrations(pool *pgxpool.Pool) error {
	// Simple migration runner that looks for .up.sql files in alphabetical order
	files, err := os.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() && (len(f.Name()) > 7 && f.Name()[len(f.Name())-7:] == ".up.sql") {
			log.Printf("Running migration: %s\n", f.Name())
			content, err := os.ReadFile("migrations/" + f.Name())
			if err != nil {
				return err
			}
			_, err = pool.Exec(context.Background(), string(content))
			if err != nil {
				log.Printf("Error in migration %s: %v\n", f.Name(), err)
			}
		}
	}
	return nil
}
