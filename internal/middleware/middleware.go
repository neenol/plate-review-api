package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	clerkUserIDKey contextKey = "clerk_user_id"
	requestIDKey   contextKey = "request_id"
)

// RequestID attaches a unique request ID to the context and response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs each request with duration and status using slog.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		reqID, _ := r.Context().Value(requestIDKey).(string)
		slog.InfoContext(r.Context(), "request",
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// AuthVerifier is the subset of the Clerk client needed by middleware.
type AuthVerifier interface {
	VerifyToken(token string) (string, error)
}

// RequireAuth verifies the Clerk JWT and injects the Clerk user ID into the
// context.  Returns 401 if the token is missing or invalid.
func RequireAuth(auth AuthVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, `{"error":{"code":"unauthorized","message":"missing authorization header"}}`, http.StatusUnauthorized)
				return
			}
			clerkUserID, err := auth.VerifyToken(token)
			if err != nil {
				slog.WarnContext(r.Context(), "auth failed", "error", err)
				http.Error(w, `{"error":{"code":"unauthorized","message":"invalid token"}}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), clerkUserIDKey, clerkUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth extracts the Clerk user ID if a Bearer token is present and valid,
// but does not fail requests without one.
func OptionalAuth(auth AuthVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token != "" {
				if clerkUserID, err := auth.VerifyToken(token); err == nil {
					ctx := context.WithValue(r.Context(), clerkUserIDKey, clerkUserID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClerkUserID retrieves the authenticated Clerk user ID from the context.
// Returns ("", false) when the user is not authenticated.
func ClerkUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(clerkUserIDKey).(string)
	return id, ok && id != ""
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
