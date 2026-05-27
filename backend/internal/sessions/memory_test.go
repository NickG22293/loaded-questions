package sessions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────

func testLobby(id string) *Lobby {
	return &Lobby{
		ID:              id,
		Players:         []*Player{{ID: "p1", Name: "Alice", IsCreator: true}},
		CreatorID:       "p1",
		TargetScore:     10,
		AnswerTimerSecs: 60,
	}
}

func testGame(id, lobbyID string) *Game {
	return &Game{
		ID:      id,
		LobbyID: lobbyID,
		Players: []*Player{{ID: "p1", Name: "Alice"}},
		CurrentRound: &Round{
			RoundNumber: 1,
			AskerID:     "p1",
			Phase:       PhaseAsking,
		},
		TargetScore: 10,
	}
}

// ── lobby ─────────────────────────────────────────────────────────────────

func TestMemoryStore_LobbyLifecycle(t *testing.T) {
	s := NewMemoryStore()
	lobby := testLobby("ABC123")

	require.NoError(t, s.CreateLobby(lobby))

	got, err := s.GetLobby("ABC123")
	require.NoError(t, err)
	assert.Equal(t, "ABC123", got.ID)

	lobby.GameStarted = true
	require.NoError(t, s.UpdateLobby(lobby))
	got, err = s.GetLobby("ABC123")
	require.NoError(t, err)
	assert.True(t, got.GameStarted)

	require.NoError(t, s.DeleteLobby("ABC123"))
	_, err = s.GetLobby("ABC123")
	assert.Error(t, err)
}

func TestMemoryStore_GetLobby_NotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetLobby("NOPE00")
	assert.Error(t, err)
}

func TestMemoryStore_UpdateLobby_NotFound(t *testing.T) {
	s := NewMemoryStore()
	err := s.UpdateLobby(&Lobby{ID: "NOPE00"})
	assert.Error(t, err)
}

// ── game ──────────────────────────────────────────────────────────────────

func TestMemoryStore_GameLifecycle(t *testing.T) {
	s := NewMemoryStore()
	game := testGame("g1", "ABC123")

	require.NoError(t, s.CreateGame(game))

	got, err := s.GetGame("g1")
	require.NoError(t, err)
	assert.Equal(t, "g1", got.ID)

	game.WinnerID = "p1"
	require.NoError(t, s.UpdateGame(game))
	got, err = s.GetGame("g1")
	require.NoError(t, err)
	assert.Equal(t, "p1", got.WinnerID)
}

func TestMemoryStore_GetGame_NotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetGame("nope")
	assert.Error(t, err)
}

func TestMemoryStore_UpdateGame_NotFound(t *testing.T) {
	s := NewMemoryStore()
	err := s.UpdateGame(&Game{ID: "nope"})
	assert.Error(t, err)
}

// ── tokens ────────────────────────────────────────────────────────────────

func TestMemoryStore_PlayerToken(t *testing.T) {
	s := NewMemoryStore()

	require.NoError(t, s.SetPlayerToken("tok1", "lobby1", "player1"))

	lobbyID, playerID, err := s.GetPlayerByToken("tok1")
	require.NoError(t, err)
	assert.Equal(t, "lobby1", lobbyID)
	assert.Equal(t, "player1", playerID)
}

func TestMemoryStore_GetPlayerByToken_NotFound(t *testing.T) {
	s := NewMemoryStore()
	_, _, err := s.GetPlayerByToken("notarealtoken")
	assert.Error(t, err)
}

// ── SSE ───────────────────────────────────────────────────────────────────

func TestMemoryStore_SSEBroadcast(t *testing.T) {
	s := NewMemoryStore()
	ch := make(chan []byte, 4)
	s.RegisterSSEClient("lobby1", ch)

	msg := []byte("event: test\ndata: {}\n\n")
	s.BroadcastToLobby("lobby1", msg)

	got := <-ch
	assert.Equal(t, msg, got)
}

func TestMemoryStore_SSEUnregister(t *testing.T) {
	s := NewMemoryStore()
	ch := make(chan []byte, 4)
	s.RegisterSSEClient("lobby1", ch)
	s.UnregisterSSEClient("lobby1", ch)

	s.BroadcastToLobby("lobby1", []byte("should not arrive"))
	assert.Empty(t, ch)
}

func TestMemoryStore_SSEBroadcastDropsWhenFull(t *testing.T) {
	s := NewMemoryStore()
	ch := make(chan []byte, 1)
	ch <- []byte("already here")
	s.RegisterSSEClient("lobby1", ch)

	s.BroadcastToLobby("lobby1", []byte("dropped"))

	assert.Len(t, ch, 1)
	assert.Equal(t, []byte("already here"), <-ch)
}

func TestMemoryStore_SSEBroadcastToUnknownLobby(t *testing.T) {
	s := NewMemoryStore()
	assert.NotPanics(t, func() {
		s.BroadcastToLobby("ghost", []byte("event: x\ndata: {}\n\n"))
	})
}
