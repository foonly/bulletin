package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SessionMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			var userID uuid.UUID
			var expiresAt time.Time
			err = db.QueryRow(r.Context(), "SELECT user_id, expires_at FROM sessions WHERE token = $1", cookie.Value).Scan(&userID, &expiresAt)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if time.Now().After(expiresAt) {
				_, _ = db.Exec(r.Context(), "DELETE FROM sessions WHERE token = $1", cookie.Value)
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
