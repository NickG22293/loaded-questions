package auth

import "context"

// User represents an authenticated user extracted from a verified JWT.
type User struct {
	ID          string // "sub" claim — provider-issued UUID
	Email       string
	DisplayName string
	AvatarURL   string
}

// Provider verifies tokens and returns the authenticated user.
// Implement this interface to swap auth backends (Supabase, Keycloak, etc.).
type Provider interface {
	VerifyToken(ctx context.Context, token string) (*User, error)
}

type contextKey struct{}

// WithUser stores a verified user in the request context.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, contextKey{}, u)
}

// GetUser returns the authenticated user from context, or nil if unauthenticated.
func GetUser(ctx context.Context) *User {
	u, _ := ctx.Value(contextKey{}).(*User)
	return u
}
