package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"loaded-questions/middleware"
)

type submitQuestionRequest struct {
	Question string `json:"question"`
}

// SubmitQuestion handles Phase 1 → transitions to Phase 2.
func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	playerID := middleware.GetPlayerID(r.Context())

	var req submitQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, "question required", http.StatusBadRequest)
		return
	}

	game, err := h.store.GetGame(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.AskerID != playerID {
		http.Error(w, "not your turn to ask", http.StatusForbidden)
		return
	}

	game.CurrentRound.Question = req.Question
	game.CurrentRound.Phase = "ANSWERING"

	if err := h.store.UpdateGame(game); err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(game.LobbyID, sseEvent("phase_changed", game.CurrentRound))
	w.WriteHeader(http.StatusNoContent)
}

type submitAnswerRequest struct {
	Answer string `json:"answer"`
}

// SubmitAnswer handles a player's answer during Phase 2.
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	playerID := middleware.GetPlayerID(r.Context())

	var req submitAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Answer == "" {
		http.Error(w, "answer required", http.StatusBadRequest)
		return
	}

	game, err := h.store.GetGame(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.Phase != "ANSWERING" {
		http.Error(w, "not in answering phase", http.StatusConflict)
		return
	}

	// Check player hasn't already answered.
	for _, a := range game.CurrentRound.Answers {
		if a.PlayerID == playerID {
			http.Error(w, "already answered", http.StatusConflict)
			return
		}
	}

	answer := &struct {
		ID       string `json:"id"`
		PlayerID string `json:"playerId"`
		Text     string `json:"text"`
	}{
		ID:       generateToken()[:8],
		PlayerID: playerID,
		Text:     req.Answer,
	}
	_ = answer // will be appended in game logic implementation

	// Placeholder: full logic implemented in game phase iteration.
	w.WriteHeader(http.StatusNoContent)
}

type assignAnswerRequest struct {
	AnswerID string `json:"answerId"`
	PlayerID string `json:"playerId"`
}

// AssignAnswer handles the Asker assigning an answer to a player during Phase 3.
func (h *Handler) AssignAnswer(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	askerID := middleware.GetPlayerID(r.Context())

	var req assignAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AnswerID == "" || req.PlayerID == "" {
		http.Error(w, "answerId and playerId required", http.StatusBadRequest)
		return
	}

	game, err := h.store.GetGame(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.AskerID != askerID {
		http.Error(w, "not your turn to assign", http.StatusForbidden)
		return
	}

	// Placeholder: full logic implemented in game phase iteration.
	_ = game
	w.WriteHeader(http.StatusNoContent)
}

// LockAssignments handles the Asker locking in all assignments, ending Phase 3.
func (h *Handler) LockAssignments(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	askerID := middleware.GetPlayerID(r.Context())

	game, err := h.store.GetGame(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	if game.CurrentRound == nil || game.CurrentRound.AskerID != askerID {
		http.Error(w, "not your turn", http.StatusForbidden)
		return
	}

	// Placeholder: full logic implemented in game phase iteration.
	_ = game
	w.WriteHeader(http.StatusNoContent)
}

// NextRound advances the game to the next round after Phase 4.
func (h *Handler) NextRound(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	_ = middleware.GetPlayerID(r.Context())

	_, err := h.store.GetGame(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	// Placeholder: full logic implemented in game phase iteration.
	w.WriteHeader(http.StatusNoContent)
}
