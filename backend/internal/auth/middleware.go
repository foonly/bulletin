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
			var token string
			cookie, err := r.Cookie("session_token")
			if err == nil {
				token = cookie.Value
			} else {
				// Fallback 1: Authorization header
				authHeader := r.Header.Get("Authorization")
				if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					token = authHeader[7:]
				} else {
					// Fallback 2: Query parameter (for WebSockets/Assets)
					token = r.URL.Query().Get("token")
				}
			}

			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			var userID uuid.UUID
			var expiresAt time.Time
			var mfaPending bool
			err = db.QueryRow(r.Context(), "SELECT user_id, expires_at, mfa_pending FROM sessions WHERE token = $1", token).Scan(&userID, &expiresAt, &mfaPending)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if time.Now().After(expiresAt) || mfaPending {
				if time.Now().After(expiresAt) {
					_, _ = db.Exec(r.Context(), "DELETE FROM sessions WHERE token = $1", token)
				}
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
