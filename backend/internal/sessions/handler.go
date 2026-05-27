package sessions

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const lobbyIDChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Handler holds shared dependencies for all session HTTP handlers.
type Handler struct {
	store Store
}

func New(s Store) *Handler {
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

func (h *Handler) getGameForLobby(lobbyID string) (*Game, error) {
	lobby, err := h.store.GetLobby(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("lobby not found")
	}
	if !lobby.GameStarted || lobby.GameID == "" {
		return nil, fmt.Errorf("game not started")
	}
	return h.store.GetGame(lobby.GameID)
}
