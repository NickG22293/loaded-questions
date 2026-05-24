package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"loaded-questions/models"
	"loaded-questions/store"
)

const lobbyIDChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	store store.Store
}

func New(s store.Store) *Handler {
	return &Handler{store: s}
}

func generateLobbyID() string {
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lobbyIDChars))))
		b[i] = lobbyIDChars[n.Int64()]
	}
	return string(b)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// sseEvent formats data as an SSE message with a named event.
func sseEvent(eventName string, data any) []byte {
	d, _ := json.Marshal(data)
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, d))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// getGameForLobby resolves the active game for a lobby in two store lookups.
// Returns a descriptive error if the lobby doesn't exist or has no game yet.
func (h *Handler) getGameForLobby(lobbyID string) (*models.Game, error) {
	lobby, err := h.store.GetLobby(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("lobby not found")
	}
	if !lobby.GameStarted || lobby.GameID == "" {
		return nil, fmt.Errorf("game not started")
	}
	return h.store.GetGame(lobby.GameID)
}
