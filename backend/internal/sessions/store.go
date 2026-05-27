package sessions

// Store defines all persistence operations for sessions.
type Store interface {
	CreateLobby(lobby *Lobby) error
	GetLobby(id string) (*Lobby, error)
	UpdateLobby(lobby *Lobby) error
	DeleteLobby(id string) error

	CreateGame(game *Game) error
	GetGame(id string) (*Game, error)
	UpdateGame(game *Game) error

	SetPlayerToken(token, lobbyID, playerID string) error
	GetPlayerByToken(token string) (lobbyID, playerID string, err error)

	RegisterSSEClient(lobbyID string, ch chan []byte)
	UnregisterSSEClient(lobbyID string, ch chan []byte)
	BroadcastToLobby(lobbyID string, event []byte)
}
