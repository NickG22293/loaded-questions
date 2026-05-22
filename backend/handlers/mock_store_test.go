package handlers

import (
	"loaded-questions/models"
	"loaded-questions/store"
)

// mockStore implements store.Store for handler tests. Set only the *Fn fields
// your test exercises; unimplemented fields delegate to the embedded nil
// interface and panic, making missing setup immediately obvious.
type mockStore struct {
	store.Store
	createLobbyFn       func(*models.Lobby) error
	getLobbyFn          func(string) (*models.Lobby, error)
	updateLobbyFn       func(*models.Lobby) error
	deleteLobbyFn       func(string) error
	createGameFn        func(*models.Game) error
	getGameFn           func(string) (*models.Game, error)
	updateGameFn        func(*models.Game) error
	setPlayerTokenFn    func(tok, lobbyID, playerID string) error
	getPlayerByTokenFn  func(string) (string, string, error)
	registerSSEFn       func(lobbyID string, ch chan []byte)
	unregisterSSEFn     func(lobbyID string, ch chan []byte)
	broadcastToLobbyFn  func(lobbyID string, event []byte)
}

func (m *mockStore) CreateLobby(l *models.Lobby) error                   { return m.createLobbyFn(l) }
func (m *mockStore) GetLobby(id string) (*models.Lobby, error)           { return m.getLobbyFn(id) }
func (m *mockStore) UpdateLobby(l *models.Lobby) error                   { return m.updateLobbyFn(l) }
func (m *mockStore) DeleteLobby(id string) error                         { return m.deleteLobbyFn(id) }
func (m *mockStore) CreateGame(g *models.Game) error                     { return m.createGameFn(g) }
func (m *mockStore) GetGame(id string) (*models.Game, error)             { return m.getGameFn(id) }
func (m *mockStore) UpdateGame(g *models.Game) error                     { return m.updateGameFn(g) }
func (m *mockStore) SetPlayerToken(tok, lid, pid string) error           { return m.setPlayerTokenFn(tok, lid, pid) }
func (m *mockStore) GetPlayerByToken(tok string) (string, string, error) { return m.getPlayerByTokenFn(tok) }
func (m *mockStore) RegisterSSEClient(id string, ch chan []byte)          { m.registerSSEFn(id, ch) }
func (m *mockStore) UnregisterSSEClient(id string, ch chan []byte)        { m.unregisterSSEFn(id, ch) }
func (m *mockStore) BroadcastToLobby(id string, event []byte)            { m.broadcastToLobbyFn(id, event) }

// noop is a convenience BroadcastToLobby implementation for tests that don't
// care about SSE side-effects.
func noop(_ string, _ []byte) {}
