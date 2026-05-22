package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"loaded-questions/models"
)

// ── helpers ───────────────────────────────────────────────────────────────

func askingGame(gameID, lobbyID, askerID string) *models.Game {
	return &models.Game{
		ID:      gameID,
		LobbyID: lobbyID,
		Players: []*models.Player{{ID: askerID, Name: "Alice"}},
		CurrentRound: &models.Round{
			RoundNumber: 1,
			AskerID:     askerID,
			Phase:       models.PhaseAsking,
		},
	}
}

func answeringGame(gameID, lobbyID, askerID string, existingAnswers []*models.Answer) *models.Game {
	return &models.Game{
		ID:      gameID,
		LobbyID: lobbyID,
		CurrentRound: &models.Round{
			RoundNumber: 1,
			AskerID:     askerID,
			Phase:       models.PhaseAnswering,
			Answers:     existingAnswers,
		},
	}
}

// ── SubmitQuestion ────────────────────────────────────────────────────────

func TestSubmitQuestion_HappyPath(t *testing.T) {
	const (
		gameID   = "g1"
		lobbyID  = "ABC123"
		askerID  = "p1"
	)
	game := askingGame(gameID, lobbyID, askerID)
	s := &mockStore{
		getGameFn:          func(string) (*models.Game, error) { return game, nil },
		updateGameFn:       func(*models.Game) error { return nil },
		broadcastToLobbyFn: noop,
	}
	h := New(s)

	body := strings.NewReader(`{"question":"What is your superpower?"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/question", body), gameID)
	req = withPlayerID(req, askerID)
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "What is your superpower?", game.CurrentRound.Question)
	assert.Equal(t, models.PhaseAnswering, game.CurrentRound.Phase)
}

func TestSubmitQuestion_EmptyQuestion(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"question":""}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/question", body), "g1")
	req = withPlayerID(req, "p1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubmitQuestion_GameNotFound(t *testing.T) {
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	body := strings.NewReader(`{"question":"What is your superpower?"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/question", body), "g1")
	req = withPlayerID(req, "p1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSubmitQuestion_NotAsker(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"question":"What is your superpower?"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/question", body), "g1")
	req = withPlayerID(req, "nottheasker")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestSubmitQuestion_NoCurrentRound(t *testing.T) {
	game := &models.Game{ID: "g1", LobbyID: "ABC123", CurrentRound: nil}
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"question":"Anything?"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/question", body), "g1")
	req = withPlayerID(req, "p1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// ── SubmitAnswer ──────────────────────────────────────────────────────────

func TestSubmitAnswer_HappyPath(t *testing.T) {
	game := answeringGame("g1", "ABC123", "asker1", nil)
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/answer", body), "g1")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestSubmitAnswer_GameNotFound(t *testing.T) {
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/answer", body), "g1")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSubmitAnswer_WrongPhase(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1") // phase = ASKING, not ANSWERING
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/answer", body), "g1")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestSubmitAnswer_AlreadyAnswered(t *testing.T) {
	existing := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "first answer"}}
	game := answeringGame("g1", "ABC123", "asker1", existing)
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"answer":"second attempt"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/answer", body), "g1")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestSubmitAnswer_EmptyAnswer(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"answer":""}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/answer", body), "g1")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── AssignAnswer ──────────────────────────────────────────────────────────

func TestAssignAnswer_HappyPath(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	game.CurrentRound.Phase = models.PhaseAssigning
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/assign", body), "g1")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestAssignAnswer_NotAsker(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/assign", body), "g1")
	req = withPlayerID(req, "notasker")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAssignAnswer_MissingFields(t *testing.T) {
	h := New(&mockStore{})

	body := strings.NewReader(`{"answerId":"a1"}`) // missing playerId
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/assign", body), "g1")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAssignAnswer_GameNotFound(t *testing.T) {
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/assign", body), "g1")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── LockAssignments ───────────────────────────────────────────────────────

func TestLockAssignments_HappyPath(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	game.CurrentRound.Phase = models.PhaseAssigning
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/lock", nil), "g1")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestLockAssignments_NotAsker(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/lock", nil), "g1")
	req = withPlayerID(req, "notasker")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestLockAssignments_GameNotFound(t *testing.T) {
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/lock", nil), "g1")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── NextRound ─────────────────────────────────────────────────────────────

func TestNextRound_HappyPath(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return game, nil }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/next", nil), "g1")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestNextRound_GameNotFound(t *testing.T) {
	s := &mockStore{getGameFn: func(string) (*models.Game, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/games/g1/next", nil), "g1")
	req = withPlayerID(req, "p1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
