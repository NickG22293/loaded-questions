package sessions

import (
	"fmt"
	"sync"
)

type playerToken struct {
	lobbyID  string
	playerID string
}

type MemoryStore struct {
	mu      sync.RWMutex
	lobbies map[string]*Lobby
	games   map[string]*Game
	tokens  map[string]playerToken
	clients map[string][]chan []byte
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		lobbies: make(map[string]*Lobby),
		games:   make(map[string]*Game),
		tokens:  make(map[string]playerToken),
		clients: make(map[string][]chan []byte),
	}
}

func (s *MemoryStore) CreateLobby(lobby *Lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lobbies[lobby.ID] = lobby
	return nil
}

func (s *MemoryStore) GetLobby(id string) (*Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lobby, ok := s.lobbies[id]
	if !ok {
		return nil, fmt.Errorf("lobby %s not found", id)
	}
	return lobby, nil
}

func (s *MemoryStore) UpdateLobby(lobby *Lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lobbies[lobby.ID]; !ok {
		return fmt.Errorf("lobby %s not found", lobby.ID)
	}
	s.lobbies[lobby.ID] = lobby
	return nil
}

func (s *MemoryStore) DeleteLobby(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lobbies, id)
	return nil
}

func (s *MemoryStore) CreateGame(game *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[game.ID] = game
	return nil
}

func (s *MemoryStore) GetGame(id string) (*Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, ok := s.games[id]
	if !ok {
		return nil, fmt.Errorf("game %s not found", id)
	}
	return game, nil
}

func (s *MemoryStore) UpdateGame(game *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[game.ID]; !ok {
		return fmt.Errorf("game %s not found", game.ID)
	}
	s.games[game.ID] = game
	return nil
}

func (s *MemoryStore) SetPlayerToken(token, lobbyID, playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = playerToken{lobbyID: lobbyID, playerID: playerID}
	return nil
}

func (s *MemoryStore) GetPlayerByToken(token string) (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pt, ok := s.tokens[token]
	if !ok {
		return "", "", fmt.Errorf("token not found")
	}
	return pt.lobbyID, pt.playerID, nil
}

func (s *MemoryStore) RegisterSSEClient(lobbyID string, ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[lobbyID] = append(s.clients[lobbyID], ch)
}

func (s *MemoryStore) UnregisterSSEClient(lobbyID string, ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clients := s.clients[lobbyID]
	for i, c := range clients {
		if c == ch {
			s.clients[lobbyID] = append(clients[:i], clients[i+1:]...)
			return
		}
	}
}

func (s *MemoryStore) BroadcastToLobby(lobbyID string, event []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.clients[lobbyID] {
		select {
		case ch <- event:
		default:
		}
	}
}
