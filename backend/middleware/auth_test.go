package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"loaded-questions/store"
)

// tokenStore is a minimal Store stub that only implements GetPlayerByToken.
// All other methods delegate to the embedded nil interface and panic if called.
type tokenStore struct {
	store.Store
	fn func(string) (string, string, error)
}

func (s *tokenStore) GetPlayerByToken(token string) (string, string, error) {
	return s.fn(token)
}

func TestAuth_NoCookie(t *testing.T) {
	s := &tokenStore{fn: func(string) (string, string, error) {
		t.Fatal("GetPlayerByToken should not be called without a cookie")
		return "", "", nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	Auth(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	s := &tokenStore{fn: func(string) (string, string, error) {
		return "", "", fmt.Errorf("token not found")
	}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "player_token", Value: "badtoken"})
	rr := httptest.NewRecorder()

	Auth(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_ValidToken_CallsNext(t *testing.T) {
	s := &tokenStore{fn: func(token string) (string, string, error) {
		assert.Equal(t, "validtoken", token)
		return "lobby1", "player1", nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "player_token", Value: "validtoken"})
	rr := httptest.NewRecorder()

	var capturedCtx context.Context
	Auth(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "player1", GetPlayerID(capturedCtx))
	assert.Equal(t, "lobby1", GetLobbyID(capturedCtx))
}

func TestGetPlayerID_EmptyWhenNotSet(t *testing.T) {
	assert.Empty(t, GetPlayerID(context.Background()))
}

func TestGetLobbyID_EmptyWhenNotSet(t *testing.T) {
	assert.Empty(t, GetLobbyID(context.Background()))
}
