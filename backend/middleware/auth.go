package middleware

import (
	"context"
	"net/http"

	"loaded-questions/store"
)

type contextKey string

const (
	PlayerIDKey contextKey = "playerID"
	LobbyIDKey  contextKey = "lobbyID"
)

// Auth injects the authenticated player's IDs into the request context via cookie.
func Auth(s store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("player_token")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			lobbyID, playerID, err := s.GetPlayerByToken(cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), PlayerIDKey, playerID)
			ctx = context.WithValue(ctx, LobbyIDKey, lobbyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetPlayerID(ctx context.Context) string {
	v, _ := ctx.Value(PlayerIDKey).(string)
	return v
}

func GetLobbyID(ctx context.Context) string {
	v, _ := ctx.Value(LobbyIDKey).(string)
	return v
}
