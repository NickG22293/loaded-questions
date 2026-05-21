package store

import "loaded-questions/models"

// Store defines all persistence operations. Swap implementations (memory, Redis, Postgres)
// by satisfying this interface and injecting the concrete type at startup.
type Store interface {
	// Lobby operations
	CreateLobby(lobby *models.Lobby) error
	GetLobby(id string) (*models.Lobby, error)
	UpdateLobby(lobby *models.Lobby) error
	DeleteLobby(id string) error

	// Game operations
	CreateGame(game *models.Game) error
	GetGame(id string) (*models.Game, error)
	UpdateGame(game *models.Game) error

	// Token/session operations
	SetPlayerToken(token, lobbyID, playerID string) error
	GetPlayerByToken(token string) (lobbyID, playerID string, err error)

	// SSE broadcast operations
	RegisterSSEClient(lobbyID string, ch chan []byte)
	UnregisterSSEClient(lobbyID string, ch chan []byte)
	BroadcastToLobby(lobbyID string, event []byte)
}
