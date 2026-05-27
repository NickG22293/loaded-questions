package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockProvider implements Provider for tests.
type mockProvider struct {
	fn func(ctx context.Context, token string) (*User, error)
}

func (m *mockProvider) VerifyToken(ctx context.Context, token string) (*User, error) {
	return m.fn(ctx, token)
}

var validUser = &User{ID: "u1", Email: "alice@example.com", DisplayName: "Alice"}

func okProvider() *mockProvider {
	return &mockProvider{fn: func(_ context.Context, _ string) (*User, error) {
		return validUser, nil
	}}
}

func errProvider() *mockProvider {
	return &mockProvider{fn: func(_ context.Context, _ string) (*User, error) {
		return nil, fmt.Errorf("invalid token")
	}}
}

// ── Required ──────────────────────────────────────────────────────────────

func TestRequired_NoHeader_401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	Required(okProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequired_InvalidToken_401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	rr := httptest.NewRecorder()
	Required(errProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequired_ValidToken_InjectsUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer goodtoken")
	rr := httptest.NewRecorder()

	var got *User
	Required(okProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, validUser, got)
}

func TestRequired_MalformedHeader_401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc123") // not "Bearer"
	rr := httptest.NewRecorder()
	Required(okProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// ── Optional ──────────────────────────────────────────────────────────────

func TestOptional_NoHeader_PassesThrough(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	var got *User
	Optional(okProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, got)
}

func TestOptional_InvalidToken_PassesThrough(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	rr := httptest.NewRecorder()

	var got *User
	Optional(errProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, got)
}

func TestOptional_ValidToken_InjectsUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer goodtoken")
	rr := httptest.NewRecorder()

	var got *User
	Optional(okProvider())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, validUser, got)
}

// ── GetUser ───────────────────────────────────────────────────────────────

func TestGetUser_NilWhenNotSet(t *testing.T) {
	assert.Nil(t, GetUser(context.Background()))
}
