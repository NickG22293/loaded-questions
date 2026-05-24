package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"loaded-questions/middleware"
	"loaded-questions/models"
)

// GetGame returns the active game for a lobby. Used by clients to fetch
// initial state when they navigate to the game page.
func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	game, err := h.getGameForLobby(lobbyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, game)
}

type submitQuestionRequest struct {
	Question string `json:"question"`
}

// SubmitQuestion handles Phase 1 → transitions to Phase 2 (ANSWERING).
func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	playerID := middleware.GetPlayerID(r.Context())

	var req submitQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, "question required", http.StatusBadRequest)
		return
	}

	game, err := h.getGameForLobby(lobbyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.AskerID != playerID {
		http.Error(w, "not your turn to ask", http.StatusForbidden)
		return
	}

	now := time.Now()
	game.CurrentRound.Question = req.Question
	game.CurrentRound.Phase = models.PhaseAnswering
	game.CurrentRound.PhaseStartedAt = &now

	if err := h.store.UpdateGame(game); err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(lobbyID, sseEvent("phase_changed", game))
	w.WriteHeader(http.StatusNoContent)
}

type submitAnswerRequest struct {
	Answer string `json:"answer"`
}

type answerCountEvent struct {
	Submitted          int      `json:"submitted"`
	Total              int      `json:"total"`
	SubmittedPlayerIDs []string `json:"submittedPlayerIds"`
}

// SubmitAnswer handles a player's answer during Phase 2. Resubmission replaces
// the existing answer. When all eligible players have answered the round
// transitions to ASSIGNING automatically.
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	playerID := middleware.GetPlayerID(r.Context())

	var req submitAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Answer == "" {
		http.Error(w, "answer required", http.StatusBadRequest)
		return
	}

	game, err := h.getGameForLobby(lobbyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.Phase != models.PhaseAnswering {
		http.Error(w, "not in answering phase", http.StatusConflict)
		return
	}
	if game.CurrentRound.AskerID == playerID {
		http.Error(w, "asker cannot answer", http.StatusForbidden)
		return
	}

	replaced := false
	for _, a := range game.CurrentRound.Answers {
		if a.PlayerID == playerID {
			a.Text = req.Answer
			replaced = true
			break
		}
	}
	if !replaced {
		game.CurrentRound.Answers = append(game.CurrentRound.Answers, &models.Answer{
			ID:       generateToken()[:8],
			PlayerID: playerID,
			Text:     req.Answer,
		})
	}

	total := len(game.Players) - 1
	submitted := len(game.CurrentRound.Answers)

	submittedIDs := make([]string, 0, submitted)
	for _, a := range game.CurrentRound.Answers {
		submittedIDs = append(submittedIDs, a.PlayerID)
	}

	allAnswered := total > 0 && submitted >= total
	if allAnswered {
		game.CurrentRound.Phase = models.PhaseAssigning
	}

	if err := h.store.UpdateGame(game); err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(lobbyID, sseEvent("answer_count", answerCountEvent{
		Submitted:          submitted,
		Total:              total,
		SubmittedPlayerIDs: submittedIDs,
	}))
	if allAnswered {
		h.store.BroadcastToLobby(lobbyID, sseEvent("phase_changed", game))
	}

	w.WriteHeader(http.StatusNoContent)
}

type assignAnswerRequest struct {
	AnswerID string `json:"answerId"`
	PlayerID string `json:"playerId"`
}

// AssignAnswer handles the Asker assigning an answer to a player during Phase 3.
// The Asker can reassign at any time before locking in.
func (h *Handler) AssignAnswer(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	askerID := middleware.GetPlayerID(r.Context())

	var req assignAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AnswerID == "" || req.PlayerID == "" {
		http.Error(w, "answerId and playerId required", http.StatusBadRequest)
		return
	}

	game, err := h.getGameForLobby(lobbyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.AskerID != askerID {
		http.Error(w, "not your turn to assign", http.StatusForbidden)
		return
	}
	if game.CurrentRound.Phase != models.PhaseAssigning {
		http.Error(w, "not in assigning phase", http.StatusConflict)
		return
	}

	// Validate the answer exists.
	var target *models.Answer
	for _, a := range game.CurrentRound.Answers {
		if a.ID == req.AnswerID {
			target = a
			break
		}
	}
	if target == nil {
		http.Error(w, "answer not found", http.StatusNotFound)
		return
	}

	// Validate the assigned player exists in the game and is not the Asker.
	if req.PlayerID == askerID {
		http.Error(w, "cannot assign to the asker", http.StatusBadRequest)
		return
	}
	playerExists := false
	for _, p := range game.Players {
		if p.ID == req.PlayerID {
			playerExists = true
			break
		}
	}
	if !playerExists {
		http.Error(w, "player not found", http.StatusBadRequest)
		return
	}

	target.AssignedTo = req.PlayerID

	if err := h.store.UpdateGame(game); err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(lobbyID, sseEvent("assignment_updated", game.CurrentRound))
	w.WriteHeader(http.StatusNoContent)
}

// LockAssignments handles the Asker locking in all assignments, ending Phase 3.
// Scores the round and transitions to SCORING.
func (h *Handler) LockAssignments(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	askerID := middleware.GetPlayerID(r.Context())

	game, err := h.getGameForLobby(lobbyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.AskerID != askerID {
		http.Error(w, "not your turn", http.StatusForbidden)
		return
	}
	if game.CurrentRound.Phase != models.PhaseAssigning {
		http.Error(w, "not in assigning phase", http.StatusConflict)
		return
	}

	// All answers must be assigned before locking in.
	for _, a := range game.CurrentRound.Answers {
		if a.AssignedTo == "" {
			http.Error(w, "all answers must be assigned before locking in", http.StatusBadRequest)
			return
		}
	}

	// Score: 1 point per correctly identified answer.
	score := 0
	for _, a := range game.CurrentRound.Answers {
		if a.AssignedTo == a.PlayerID {
			score++
		}
	}
	for _, p := range game.Players {
		if p.ID == askerID {
			p.Score += score
			break
		}
	}

	game.CurrentRound.Phase = models.PhaseScoring

	if err := h.store.UpdateGame(game); err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(lobbyID, sseEvent("phase_changed", game))
	w.WriteHeader(http.StatusNoContent)
}

// NextRound advances the game after Phase 4. Any authenticated player may call
// this; the first request wins. Checks for a winner and either transitions to
// GAME_OVER or starts the next round.
func (h *Handler) NextRound(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")

	game, err := h.getGameForLobby(lobbyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.Phase != models.PhaseScoring {
		http.Error(w, "not in scoring phase", http.StatusConflict)
		return
	}

	// Check for a winner.
	for _, p := range game.Players {
		if p.Score >= game.TargetScore {
			game.WinnerID = p.ID
			game.CurrentRound.Phase = models.PhaseGameOver
			if err := h.store.UpdateGame(game); err != nil {
				http.Error(w, "failed to update game", http.StatusInternalServerError)
				return
			}
			h.store.BroadcastToLobby(lobbyID, sseEvent("phase_changed", game))
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Archive current round and start the next one.
	game.Rounds = append(game.Rounds, game.CurrentRound)

	// Rotate the Asker.
	currentIdx := 0
	for i, p := range game.Players {
		if p.ID == game.CurrentRound.AskerID {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(game.Players)
	nextAskerID := game.Players[nextIdx].ID

	game.CurrentRound = &models.Round{
		RoundNumber: game.CurrentRound.RoundNumber + 1,
		AskerID:     nextAskerID,
		Phase:       models.PhaseAsking,
	}

	if err := h.store.UpdateGame(game); err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(lobbyID, sseEvent("phase_changed", game))
	w.WriteHeader(http.StatusNoContent)
}
