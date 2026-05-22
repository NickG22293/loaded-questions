package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"loaded-questions/middleware"
	"loaded-questions/models"
)

// ── test helpers ──────────────────────────────────────────────────────────

func withChiID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withPlayerID(r *http.Request, playerID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middleware.PlayerIDKey, playerID))
}

func twoPlayerLobby(creatorID string) *models.Lobby {
	return &models.Lobby{
		ID:          "ABC123",
		CreatorID:   creatorID,
		TargetScore: 10,
		Players: []*models.Player{
			{ID: creatorID, Name: "Alice", IsCreator: true},
			{ID: "p2", Name: "Bob"},
		},
	}
}

// ── CreateLobby ───────────────────────────────────────────────────────────

func TestCreateLobby_HappyPath(t *testing.T) {
	s := &mockStore{
		createLobbyFn:    func(*models.Lobby) error { return nil },
		setPlayerTokenFn: func(_, _, _ string) error { return nil },
	}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Alice","targetScore":5,"answerTimerSecs":30}`)
	req := httptest.NewRequest(http.MethodPost, "/api/lobbies", body)
	rr := httptest.NewRecorder()
	h.CreateLobby(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["lobbyId"])
	assert.Len(t, resp["lobbyId"], 6, "lobby ID should be 6 chars")
	assert.NotEmpty(t, resp["playerId"])

	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "player_token", cookies[0].Name)
	assert.True(t, cookies[0].HttpOnly)
}

func TestCreateLobby_DefaultsApplied(t *testing.T) {
	var captured *models.Lobby
	s := &mockStore{
		createLobbyFn:    func(l *models.Lobby) error { captured = l; return nil },
		setPlayerTokenFn: func(_, _, _ string) error { return nil },
	}
	h := New(s)

	// omit targetScore and answerTimerSecs — should get defaults
	body := strings.NewReader(`{"playerName":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/lobbies", body)
	rr := httptest.NewRecorder()
	h.CreateLobby(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.NotNil(t, captured)
	assert.Equal(t, 10, captured.TargetScore)
	assert.Equal(t, 60, captured.AnswerTimerSecs)
}

func TestCreateLobby_CreatorFlagSet(t *testing.T) {
	var captured *models.Lobby
	s := &mockStore{
		createLobbyFn:    func(l *models.Lobby) error { captured = l; return nil },
		setPlayerTokenFn: func(_, _, _ string) error { return nil },
	}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/lobbies", body)
	rr := httptest.NewRecorder()
	h.CreateLobby(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Len(t, captured.Players, 1)
	assert.True(t, captured.Players[0].IsCreator)
	assert.Equal(t, captured.CreatorID, captured.Players[0].ID)
}

func TestCreateLobby_MissingName(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"playerName":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/lobbies", body)
	rr := httptest.NewRecorder()
	h.CreateLobby(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateLobby_BadJSON(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/lobbies", body)
	rr := httptest.NewRecorder()
	h.CreateLobby(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── JoinLobby ─────────────────────────────────────────────────────────────

func TestJoinLobby_HappyPath(t *testing.T) {
	lobby := &models.Lobby{
		ID:      "ABC123",
		Players: []*models.Player{{ID: "p1", Name: "Alice", IsCreator: true}},
	}
	s := &mockStore{
		getLobbyFn:       func(string) (*models.Lobby, error) { return lobby, nil },
		updateLobbyFn:    func(*models.Lobby) error { return nil },
		setPlayerTokenFn: func(_, _, _ string) error { return nil },
		broadcastToLobbyFn: noop,
	}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Bob"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/join", body), "ABC123")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["playerId"])

	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "player_token", cookies[0].Name)
}

func TestJoinLobby_PlayerAppendedToLobby(t *testing.T) {
	lobby := &models.Lobby{
		ID:      "ABC123",
		Players: []*models.Player{{ID: "p1", Name: "Alice", IsCreator: true}},
	}
	var updated *models.Lobby
	s := &mockStore{
		getLobbyFn:         func(string) (*models.Lobby, error) { return lobby, nil },
		updateLobbyFn:      func(l *models.Lobby) error { updated = l; return nil },
		setPlayerTokenFn:   func(_, _, _ string) error { return nil },
		broadcastToLobbyFn: noop,
	}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Bob"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/join", body), "ABC123")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, updated)
	assert.Len(t, updated.Players, 2)
	assert.Equal(t, "Bob", updated.Players[1].Name)
}

func TestJoinLobby_LobbyNotFound(t *testing.T) {
	s := &mockStore{
		getLobbyFn: func(string) (*models.Lobby, error) { return nil, fmt.Errorf("not found") },
	}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Bob"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/NOPE/join", body), "NOPE")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestJoinLobby_GameAlreadyStarted(t *testing.T) {
	lobby := &models.Lobby{ID: "ABC123", GameStarted: true}
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Bob"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/join", body), "ABC123")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestJoinLobby_LobbyFull(t *testing.T) {
	players := make([]*models.Player, 10)
	for i := range players {
		players[i] = &models.Player{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Player%d", i)}
	}
	lobby := &models.Lobby{ID: "ABC123", Players: players}
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Extra"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/join", body), "ABC123")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestJoinLobby_DuplicateName(t *testing.T) {
	lobby := &models.Lobby{
		ID:      "ABC123",
		Players: []*models.Player{{ID: "p1", Name: "Alice"}},
	}
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	body := strings.NewReader(`{"playerName":"Alice"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/join", body), "ABC123")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestJoinLobby_MissingName(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"playerName":""}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/join", body), "ABC123")
	rr := httptest.NewRecorder()
	h.JoinLobby(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── GetLobby ──────────────────────────────────────────────────────────────

func TestGetLobby_Found(t *testing.T) {
	lobby := &models.Lobby{ID: "ABC123", TargetScore: 10}
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodGet, "/api/lobbies/ABC123", nil), "ABC123")
	rr := httptest.NewRecorder()
	h.GetLobby(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var got models.Lobby
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	assert.Equal(t, "ABC123", got.ID)
}

func TestGetLobby_NotFound(t *testing.T) {
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodGet, "/api/lobbies/NOPE", nil), "NOPE")
	rr := httptest.NewRecorder()
	h.GetLobby(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── StartGame ─────────────────────────────────────────────────────────────

func TestStartGame_HappyPath(t *testing.T) {
	const creatorID = "creator1"
	lobby := twoPlayerLobby(creatorID)
	var broadcasted []byte
	s := &mockStore{
		getLobbyFn:         func(string) (*models.Lobby, error) { return lobby, nil },
		updateLobbyFn:      func(*models.Lobby) error { return nil },
		broadcastToLobbyFn: func(_ string, event []byte) { broadcasted = event },
	}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/start", nil), "ABC123")
	req = withPlayerID(req, creatorID)
	rr := httptest.NewRecorder()
	h.StartGame(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.True(t, lobby.GameStarted)
	assert.Contains(t, string(broadcasted), "game_started")
}

func TestStartGame_NonCreatorForbidden(t *testing.T) {
	lobby := twoPlayerLobby("creator1")
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/start", nil), "ABC123")
	req = withPlayerID(req, "someoneelse")
	rr := httptest.NewRecorder()
	h.StartGame(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestStartGame_AlreadyStarted(t *testing.T) {
	lobby := twoPlayerLobby("creator1")
	lobby.GameStarted = true
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/start", nil), "ABC123")
	req = withPlayerID(req, "creator1")
	rr := httptest.NewRecorder()
	h.StartGame(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestStartGame_NotEnoughPlayers(t *testing.T) {
	lobby := &models.Lobby{
		ID:        "ABC123",
		CreatorID: "creator1",
		Players:   []*models.Player{{ID: "creator1", Name: "Alice", IsCreator: true}},
	}
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/start", nil), "ABC123")
	req = withPlayerID(req, "creator1")
	rr := httptest.NewRecorder()
	h.StartGame(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestStartGame_LobbyNotFound(t *testing.T) {
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/NOPE/start", nil), "NOPE")
	req = withPlayerID(req, "creator1")
	rr := httptest.NewRecorder()
	h.StartGame(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
