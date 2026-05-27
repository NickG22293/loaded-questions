package sessions

import (
	"context"
	"net/http"
)

type contextKey string

const (
	PlayerIDKey contextKey = "playerID"
	LobbyIDKey  contextKey = "lobbyID"
)

// Auth injects the authenticated player's IDs into the request context.
// It reads the token from the X-Player-Token header first, then falls back
// to the player_token cookie. The header is preferred so that multiple
// browser tabs can each carry their own per-tab token via sessionStorage.
func Auth(s Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Player-Token")
			if token == "" {
				cookie, err := r.Cookie("player_token")
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				token = cookie.Value
			}
			lobbyID, playerID, err := s.GetPlayerByToken(token)
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
