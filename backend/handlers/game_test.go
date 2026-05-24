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

// ── test helpers ──────────────────────────────────────────────────────────

// mockLobbyGame wires getLobbyFn + getGameFn together so getGameForLobby works.
func mockLobbyGame(lobbyID string, game *models.Game) *mockStore {
	lobby := &models.Lobby{
		ID:          lobbyID,
		GameStarted: true,
		GameID:      game.ID,
	}
	return &mockStore{
		getLobbyFn: func(string) (*models.Lobby, error) { return lobby, nil },
		getGameFn:  func(string) (*models.Game, error) { return game, nil },
	}
}

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

func answeringGame(gameID, lobbyID, askerID string, existing []*models.Answer) *models.Game {
	return &models.Game{
		ID:      gameID,
		LobbyID: lobbyID,
		Players: []*models.Player{
			{ID: askerID, Name: "Asker"},
			{ID: "p2", Name: "Bob"},
		},
		CurrentRound: &models.Round{
			RoundNumber: 1,
			AskerID:     askerID,
			Phase:       models.PhaseAnswering,
			Answers:     existing,
		},
	}
}

// assigningGame returns a game in Phase 3 with the given answers.
// Players: asker + "p2" (Bob) + "p3" (Carol).
func assigningGame(gameID, lobbyID, askerID string, answers []*models.Answer) *models.Game {
	return &models.Game{
		ID:          gameID,
		LobbyID:     lobbyID,
		TargetScore: 10,
		Players: []*models.Player{
			{ID: askerID, Name: "Asker", Score: 0},
			{ID: "p2", Name: "Bob", Score: 0},
			{ID: "p3", Name: "Carol", Score: 0},
		},
		CurrentRound: &models.Round{
			RoundNumber: 1,
			AskerID:     askerID,
			Question:    "Test question",
			Phase:       models.PhaseAssigning,
			Answers:     answers,
		},
		Rounds: []*models.Round{},
	}
}

// scoringGame returns a game in Phase 4 with the given answers (AssignedTo set).
func scoringGame(gameID, lobbyID, askerID string, answers []*models.Answer) *models.Game {
	return &models.Game{
		ID:          gameID,
		LobbyID:     lobbyID,
		TargetScore: 10,
		Players: []*models.Player{
			{ID: askerID, Name: "Asker", Score: 0},
			{ID: "p2", Name: "Bob", Score: 0},
		},
		CurrentRound: &models.Round{
			RoundNumber: 1,
			AskerID:     askerID,
			Phase:       models.PhaseScoring,
			Answers:     answers,
		},
		Rounds: []*models.Round{},
	}
}

// lobbyNotStarted returns a mock that reports the lobby has no game yet.
func lobbyNotStarted() *mockStore {
	return &mockStore{
		getLobbyFn: func(string) (*models.Lobby, error) {
			return &models.Lobby{ID: "ABC123", GameStarted: false}, nil
		},
	}
}

// ── GetGame ───────────────────────────────────────────────────────────────

func TestGetGame_Found(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodGet, "/api/lobbies/ABC123/game", nil), "ABC123")
	rr := httptest.NewRecorder()
	h.GetGame(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetGame_LobbyNotStarted(t *testing.T) {
	h := New(lobbyNotStarted())

	req := withChiID(httptest.NewRequest(http.MethodGet, "/api/lobbies/ABC123/game", nil), "ABC123")
	rr := httptest.NewRecorder()
	h.GetGame(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetGame_LobbyNotFound(t *testing.T) {
	s := &mockStore{getLobbyFn: func(string) (*models.Lobby, error) { return nil, fmt.Errorf("not found") }}
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodGet, "/api/lobbies/NOPE/game", nil), "NOPE")
	rr := httptest.NewRecorder()
	h.GetGame(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── SubmitQuestion ────────────────────────────────────────────────────────

func TestSubmitQuestion_HappyPath(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	body := strings.NewReader(`{"question":"What is your superpower?"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/question", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "What is your superpower?", game.CurrentRound.Question)
	assert.Equal(t, models.PhaseAnswering, game.CurrentRound.Phase)
}

func TestSubmitQuestion_BroadcastsPhaseChanged(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	var broadcasted string
	s.broadcastToLobbyFn = func(_ string, event []byte) { broadcasted = string(event) }
	h := New(s)

	body := strings.NewReader(`{"question":"A question"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/question", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Contains(t, broadcasted, "phase_changed")
	assert.Contains(t, broadcasted, "ANSWERING")
}

func TestSubmitQuestion_EmptyQuestion(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"question":""}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/question", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubmitQuestion_LobbyNotStarted(t *testing.T) {
	h := New(lobbyNotStarted())
	body := strings.NewReader(`{"question":"A question"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/question", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSubmitQuestion_NotAsker(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"question":"A question"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/question", body), "ABC123")
	req = withPlayerID(req, "nottheasker")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestSubmitQuestion_NoCurrentRound(t *testing.T) {
	game := &models.Game{ID: "g1", LobbyID: "ABC123", CurrentRound: nil}
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"question":"Anything?"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/question", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.SubmitQuestion(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// ── SubmitAnswer ──────────────────────────────────────────────────────────

func TestSubmitAnswer_HappyPath(t *testing.T) {
	game := answeringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestSubmitAnswer_LobbyNotStarted(t *testing.T) {
	h := New(lobbyNotStarted())
	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSubmitAnswer_WrongPhase(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1") // phase ASKING, not ANSWERING
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestSubmitAnswer_ReplacesExistingAnswer(t *testing.T) {
	existing := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "first answer"}}
	game := answeringGame("g1", "ABC123", "asker1", existing)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	body := strings.NewReader(`{"answer":"updated answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "updated answer", game.CurrentRound.Answers[0].Text)
	assert.Len(t, game.CurrentRound.Answers, 1, "resubmission should replace, not append")
}

func TestSubmitAnswer_EmptyAnswer(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"answer":""}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubmitAnswer_AskerForbidden(t *testing.T) {
	game := answeringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answer":"sneaky"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestSubmitAnswer_AppendsNewAnswer(t *testing.T) {
	game := answeringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Len(t, game.CurrentRound.Answers, 1)
	assert.Equal(t, "p2", game.CurrentRound.Answers[0].PlayerID)
	assert.Equal(t, "My answer", game.CurrentRound.Answers[0].Text)
}

func TestSubmitAnswer_LastAnswerTransitionsPhase(t *testing.T) {
	// 2 players (asker + p2); p2 is the only eligible answerer.
	game := answeringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	var broadcastCount int
	s.broadcastToLobbyFn = func(_ string, _ []byte) { broadcastCount++ }
	h := New(s)

	body := strings.NewReader(`{"answer":"final answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, models.PhaseAssigning, game.CurrentRound.Phase)
	// answer_count + phase_changed = 2 broadcasts
	assert.Equal(t, 2, broadcastCount)
}

func TestSubmitAnswer_BroadcastsAnswerCount(t *testing.T) {
	game := answeringGame("g1", "ABC123", "asker1", nil)
	// Add a third player so the last-answer transition doesn't fire after p2 submits.
	game.Players = append(game.Players, &models.Player{ID: "p3", Name: "Carol"})
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	var broadcastPayload string
	s.broadcastToLobbyFn = func(_ string, event []byte) { broadcastPayload = string(event) }
	h := New(s)

	body := strings.NewReader(`{"answer":"My answer"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/answer", body), "ABC123")
	req = withPlayerID(req, "p2")
	rr := httptest.NewRecorder()
	h.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Contains(t, broadcastPayload, "answer_count")
	assert.Contains(t, broadcastPayload, `"submitted":1`)
	assert.Contains(t, broadcastPayload, `"total":2`)
	assert.Contains(t, broadcastPayload, `"p2"`)
}

// ── AssignAnswer ──────────────────────────────────────────────────────────

func TestAssignAnswer_HappyPath(t *testing.T) {
	answers := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "some answer"}}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "p2", game.CurrentRound.Answers[0].AssignedTo)
}

func TestAssignAnswer_NotAsker(t *testing.T) {
	game := assigningGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "notasker")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAssignAnswer_WrongPhase(t *testing.T) {
	game := askingGame("g1", "ABC123", "asker1")
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestAssignAnswer_AnswerNotFound(t *testing.T) {
	game := assigningGame("g1", "ABC123", "asker1", nil) // no answers
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answerId":"nonexistent","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAssignAnswer_PlayerNotInGame(t *testing.T) {
	answers := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "answer"}}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"unknown"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAssignAnswer_CannotAssignToAsker(t *testing.T) {
	answers := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "answer"}}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"asker1"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAssignAnswer_MissingFields(t *testing.T) {
	h := New(&mockStore{})
	body := strings.NewReader(`{"answerId":"a1"}`) // missing playerId
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAssignAnswer_LobbyNotStarted(t *testing.T) {
	h := New(lobbyNotStarted())
	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAssignAnswer_BroadcastsAssignmentUpdated(t *testing.T) {
	answers := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "answer"}}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	var payload string
	s.broadcastToLobbyFn = func(_ string, event []byte) { payload = string(event) }
	h := New(s)

	body := strings.NewReader(`{"answerId":"a1","playerId":"p2"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Contains(t, payload, "assignment_updated")
	assert.Contains(t, payload, `"p2"`)
}

func TestAssignAnswer_Reassignment(t *testing.T) {
	answers := []*models.Answer{{ID: "a1", PlayerID: "p2", Text: "answer", AssignedTo: "p2"}}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	// Reassign to p3
	body := strings.NewReader(`{"answerId":"a1","playerId":"p3"}`)
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/assign", body), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.AssignAnswer(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "p3", game.CurrentRound.Answers[0].AssignedTo)
}

// ── LockAssignments ───────────────────────────────────────────────────────

func TestLockAssignments_HappyPath(t *testing.T) {
	answers := []*models.Answer{
		{ID: "a1", PlayerID: "p2", Text: "answer1", AssignedTo: "p2"}, // correct
		{ID: "a2", PlayerID: "p3", Text: "answer2", AssignedTo: "p2"}, // wrong
	}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, models.PhaseScoring, game.CurrentRound.Phase)
}

func TestLockAssignments_ScoresCorrectly(t *testing.T) {
	answers := []*models.Answer{
		{ID: "a1", PlayerID: "p2", Text: "ans1", AssignedTo: "p2"}, // correct +1
		{ID: "a2", PlayerID: "p3", Text: "ans2", AssignedTo: "p2"}, // wrong
		{ID: "a3", PlayerID: "p2", Text: "ans3", AssignedTo: "p3"}, // wrong
	}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	var askerScore int
	for _, p := range game.Players {
		if p.ID == "asker1" {
			askerScore = p.Score
		}
	}
	assert.Equal(t, 1, askerScore)
}

func TestLockAssignments_UnassignedAnswerBlocked(t *testing.T) {
	answers := []*models.Answer{
		{ID: "a1", PlayerID: "p2", Text: "answer", AssignedTo: ""}, // not assigned
	}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestLockAssignments_WrongPhase(t *testing.T) {
	game := answeringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestLockAssignments_NotAsker(t *testing.T) {
	game := assigningGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "notasker")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestLockAssignments_LobbyNotStarted(t *testing.T) {
	h := New(lobbyNotStarted())
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestLockAssignments_BroadcastsPhaseChanged(t *testing.T) {
	answers := []*models.Answer{
		{ID: "a1", PlayerID: "p2", Text: "ans", AssignedTo: "p2"},
	}
	game := assigningGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	var payload string
	s.broadcastToLobbyFn = func(_ string, event []byte) { payload = string(event) }
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/lock", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.LockAssignments(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Contains(t, payload, "phase_changed")
	assert.Contains(t, payload, "SCORING")
}

// ── NextRound ─────────────────────────────────────────────────────────────

func TestNextRound_AdvancesToNextRound(t *testing.T) {
	answers := []*models.Answer{{ID: "a1", PlayerID: "p2", AssignedTo: "p2"}}
	game := scoringGame("g1", "ABC123", "asker1", answers)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/next", nil), "ABC123")
	req = withPlayerID(req, "p2") // non-asker triggers — any player can
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, models.PhaseAsking, game.CurrentRound.Phase)
	assert.Equal(t, 2, game.CurrentRound.RoundNumber)
}

func TestNextRound_RotatesAsker(t *testing.T) {
	game := scoringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/next", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	// players[0] = asker1, players[1] = p2 → next asker = p2
	assert.Equal(t, "p2", game.CurrentRound.AskerID)
}

func TestNextRound_ArchivesCurrentRound(t *testing.T) {
	game := scoringGame("g1", "ABC123", "asker1", nil)
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/next", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Len(t, game.Rounds, 1)
}

func TestNextRound_GameOver(t *testing.T) {
	game := scoringGame("g1", "ABC123", "asker1", nil)
	game.Players[0].Score = 10 // asker1 has hit the target
	game.TargetScore = 10
	s := mockLobbyGame("ABC123", game)
	s.updateGameFn = func(*models.Game) error { return nil }
	s.broadcastToLobbyFn = noop
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/next", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "asker1", game.WinnerID)
	assert.Equal(t, models.PhaseGameOver, game.CurrentRound.Phase)
}

func TestNextRound_WrongPhase(t *testing.T) {
	game := assigningGame("g1", "ABC123", "asker1", nil) // ASSIGNING, not SCORING
	s := mockLobbyGame("ABC123", game)
	h := New(s)

	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/next", nil), "ABC123")
	req = withPlayerID(req, "asker1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestNextRound_LobbyNotStarted(t *testing.T) {
	h := New(lobbyNotStarted())
	req := withChiID(httptest.NewRequest(http.MethodPost, "/api/lobbies/ABC123/next", nil), "ABC123")
	req = withPlayerID(req, "p1")
	rr := httptest.NewRecorder()
	h.NextRound(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
