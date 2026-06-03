// Package supabase implements auth.Provider using Supabase's JWKS endpoint.
//
// Swap story: to move off Supabase, implement auth.Provider backed by any
// OIDC-compliant JWKS URL (Keycloak, Dex, Auth0, etc.) and change the two
// env vars. No handler or middleware code needs to change.
package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"candid/internal/auth"
)

// Provider verifies Supabase JWTs using the project's JWKS endpoint.
type Provider struct {
	issuer  string
	keyfunc jwt.Keyfunc
}

// supabaseClaims maps the JWT claims Supabase includes in its tokens.
type supabaseClaims struct {
	jwt.RegisteredClaims
	Email        string       `json:"email"`
	UserMetadata userMetadata `json:"user_metadata"`
	Role         string       `json:"role"`
}

type userMetadata struct {
	FullName  string `json:"full_name"`
	Name      string `json:"name"` // fallback used by some OAuth providers
	AvatarURL string `json:"avatar_url"`
}

// NewProvider creates a Provider that fetches keys from jwksURL and validates
// the "iss" claim against issuer.
//
// jwksURL:  https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json
// issuer:   https://<project-ref>.supabase.co/auth/v1
func NewProvider(jwksURL, issuer string) (*Provider, error) {
	if jwksURL == "" || issuer == "" {
		return nil, fmt.Errorf("supabase: jwksURL and issuer are required")
	}

	kf, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("supabase: failed to initialise JWKS keyfunc: %w", err)
	}

	return &Provider{issuer: issuer, keyfunc: kf.Keyfunc}, nil
}

// VerifyToken validates a JWT and returns the authenticated user.
func (p *Provider) VerifyToken(_ context.Context, rawToken string) (*auth.User, error) {
	var claims supabaseClaims
	token, err := jwt.ParseWithClaims(rawToken, &claims, p.keyfunc,
		jwt.WithIssuedAt(),
		jwt.WithIssuer(p.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("supabase: token validation failed: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("supabase: invalid token")
	}
	if claims.Role != "authenticated" {
		return nil, fmt.Errorf("supabase: token role is not authenticated")
	}

	displayName := claims.UserMetadata.FullName
	if displayName == "" {
		displayName = claims.UserMetadata.Name
	}

	return &auth.User{
		ID:          claims.Subject,
		Email:       claims.Email,
		DisplayName: displayName,
		AvatarURL:   claims.UserMetadata.AvatarURL,
	}, nil
}

// NewProviderFromJWTSecret creates a Provider that verifies tokens using a
// shared HMAC-SHA256 secret instead of JWKS. Use this only when a JWKS
// endpoint is unavailable (e.g. local dev with Supabase CLI).
func NewProviderFromJWTSecret(secret, issuer string) (*Provider, error) {
	if secret == "" || issuer == "" {
		return nil, fmt.Errorf("supabase: secret and issuer are required")
	}

	secretBytes := []byte(secret)
	kf := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretBytes, nil
	}

	return &Provider{issuer: issuer, keyfunc: kf}, nil
}

// healthCheck does a quick GET on the JWKS URL and returns an error if it
// is unreachable. Useful for startup validation.
func healthCheck(jwksURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}
	var v json.RawMessage
	return json.NewDecoder(resp.Body).Decode(&v)
}
