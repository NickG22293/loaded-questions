package auth

import (
	"net/http"
	"strings"
)

// Required rejects requests without a valid JWT with 401.
// Use for routes that must have an authenticated user.
func Required(p Provider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			user, err := p.VerifyToken(r.Context(), token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
		})
	}
}

// Optional injects the authenticated user into context if a valid JWT is present.
// Requests without a token (or with an invalid one) pass through unauthenticated.
// Use for routes that behave differently for logged-in users but don't require login.
func Optional(p Provider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, ok := bearerToken(r); ok {
				if user, err := p.VerifyToken(r.Context(), token); err == nil {
					r = r.WithContext(WithUser(r.Context(), user))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(h, "Bearer ")
	if token == "" {
		return "", false
	}
	return token, true
}
