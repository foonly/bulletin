package chat

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CleanupWorker struct {
	DB *pgxpool.Pool
}

func NewCleanupWorker(db *pgxpool.Pool) *CleanupWorker {
	return &CleanupWorker{DB: db}
}

func (w *CleanupWorker) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	log.Printf("Chat cleanup worker started (Interval: %v)\n", interval)

	go func() {
		for range ticker.C {
			w.Cleanup()
		}
	}()
}

func (w *CleanupWorker) Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This query deletes messages that meet BOTH criteria:
	// 1. They are older than the circle's retention_days
	// 2. They are NOT among the latest retention_count messages
	query := `
		DELETE FROM chat_messages
		WHERE id IN (
			SELECT cm.id
			FROM chat_messages cm
			JOIN circles c ON cm.circle_id = c.id
			WHERE cm.created_at < NOW() - (c.chat_retention_days || ' days')::interval
			AND cm.id NOT IN (
				SELECT id
				FROM (
					SELECT id, ROW_NUMBER() OVER (PARTITION BY circle_id ORDER BY created_at DESC) as rank
					FROM chat_messages
				) ranks
				WHERE rank <= c.chat_retention_count
			)
		)
	`

	tag, err := w.DB.Exec(ctx, query)
	if err != nil {
		log.Printf("Chat cleanup error: %v\n", err)
		return
	}

	if tag.RowsAffected() > 0 {
		log.Printf("Chat cleanup complete: purged %d old messages\n", tag.RowsAffected())
	}
}
