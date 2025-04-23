package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
)

var TokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

type Token struct {
	db      *sqlx.DB
	jwtAuth *jwtauth.JWTAuth
}

func NewTokenManager(db *sqlx.DB) *Token {
	return &Token{
		db:      db,
		jwtAuth: TokenAuth,
	}
}

func (a *Token) TokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := jwtauth.TokenFromHeader(r)
		token, err := jwtauth.VerifyToken(a.jwtAuth, t)

		if err != nil || token == nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Token) addToBlacklist(ctx context.Context, tokenID, token string, expiresAt time.Time) error {
	_, err := a.db.ExecContext(ctx, `
		INSERT INTO token_blacklist (token_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_id) DO UPDATE
		SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at
	`, tokenID, token, expiresAt)
	return err
}

func ClearTokenFromClient(w http.ResponseWriter) {
	w.Header().Del("Authorization")

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// BlacklistMiddleware checks PostgreSQL for revoked tokens
func (a *Token) BlacklistMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tokenString := jwtauth.TokenFromHeader(r)
		if tokenString == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if token is blacklisted
		var exists bool
		err := a.db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM token_blacklist 
				WHERE token = $1 AND expires_at > NOW()
			)
		`, tokenString).Scan(&exists)

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if exists {
			http.Error(w, "Token revoked", http.StatusUnauthorized)
			return
		}

		// Cleanup expired tokens (could be moved to a background job)
		_, _ = a.db.ExecContext(ctx, "DELETE FROM token_blacklist WHERE expires_at <= NOW()")

		next.ServeHTTP(w, r)
	})
}
