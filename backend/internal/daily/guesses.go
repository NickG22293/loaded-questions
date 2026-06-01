package daily

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"loaded-questions/internal/auth"
	"loaded-questions/internal/httputil"
)

// GetGuessTargets returns the anonymous answer pool for the guessing phase.
// Returns 403 if guessing is not yet unlocked for this group.
func (h *Handler) GetGuessTargets(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	q, err := h.store.GetTodayQuestion(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "no question today", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	progress, err := h.store.GetGroupAnswerProgress(r.Context(), groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if !guessingUnlocked(progress) {
		http.Error(w, "guessing not yet unlocked", http.StatusForbidden)
		return
	}

	answers, err := h.store.GetGroupAnswers(r.Context(), groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}

	// Strip user attribution — answers are anonymous during guessing.
	targets := make([]GuessTarget, len(answers))
	for i, a := range answers {
		targets[i] = GuessTarget{AnswerID: a.ID, AnswerText: a.AnswerText}
	}

	members, err := h.store.GetGroupMembers(r.Context(), groupID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"answers": targets,
		"members": members,
	})
}

// SubmitGuesses locks in the user's guess assignments and immediately returns results.
func (h *Handler) SubmitGuesses(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	q, err := h.store.GetTodayQuestion(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "no question today", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	progress, err := h.store.GetGroupAnswerProgress(r.Context(), groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if !guessingUnlocked(progress) {
		http.Error(w, "guessing not yet unlocked", http.StatusForbidden)
		return
	}

	u := auth.GetUser(r.Context())
	alreadySubmitted, err := h.store.HasUserSubmittedGuesses(r.Context(), u.ID, groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if alreadySubmitted {
		http.Error(w, "guesses already submitted", http.StatusConflict)
		return
	}

	var body []struct {
		AnswerID      string `json:"answer_id"`
		GuessedUserID string `json:"guessed_user_id"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil || len(body) == 0 {
		http.Error(w, "guesses required", http.StatusBadRequest)
		return
	}

	guesses := make([]Guess, len(body))
	for i, b := range body {
		guesses[i] = Guess{
			QuestionID:    q.ID,
			GroupID:       groupID,
			GuesserID:     u.ID,
			AnswerID:      b.AnswerID,
			GuessedUserID: b.GuessedUserID,
		}
	}

	if err := h.store.SubmitGuesses(r.Context(), guesses); err != nil {
		httputil.ServerError(w, r, err)
		return
	}

	// Return results immediately after submission.
	results, err := h.store.GetGroupResults(r.Context(), u.ID, groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, results)
}

// GetResults returns the user's guess results. Returns 403 if they haven't submitted yet.
func (h *Handler) GetResults(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	q, err := h.store.GetTodayQuestion(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "no question today", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	u := auth.GetUser(r.Context())
	submitted, err := h.store.HasUserSubmittedGuesses(r.Context(), u.ID, groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if !submitted {
		http.Error(w, "guesses not yet submitted", http.StatusForbidden)
		return
	}

	results, err := h.store.GetGroupResults(r.Context(), u.ID, groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, results)
}

// GetLeaderboard returns cumulative scores for all group members across all days.
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	entries, err := h.store.GetLeaderboard(r.Context(), groupID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if entries == nil {
		entries = []LeaderboardEntry{}
	}
	httputil.WriteJSON(w, http.StatusOK, entries)
}
