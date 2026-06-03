package daily

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"candid/internal/auth"
	"candid/internal/httputil"
)

// GetToday returns today's question and the user's answer if already submitted.
func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	q, err := h.store.GetTodayQuestion(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "no question scheduled for today", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	u := auth.GetUser(r.Context())
	answer, err := h.store.GetUserAnswer(r.Context(), u.ID, q.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		httputil.ServerError(w, r, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"question": q,
		"answer":   answer, // nil if not yet answered
	})
}

// SubmitAnswer records the user's answer to today's question and broadcasts
// progress to all their groups via SSE.
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AnswerText string `json:"answer_text"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil || body.AnswerText == "" {
		http.Error(w, "answer_text required", http.StatusBadRequest)
		return
	}

	q, err := h.store.GetTodayQuestion(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "no question scheduled for today", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	u := auth.GetUser(r.Context())
	if _, err := h.store.GetUserAnswer(r.Context(), u.ID, q.ID); err == nil {
		http.Error(w, "already answered today", http.StatusConflict)
		return
	}

	a := &Answer{
		ID:         uuid.New().String(),
		QuestionID: q.ID,
		UserID:     u.ID,
		AnswerText: body.AnswerText,
	}
	if err := h.store.SubmitAnswer(r.Context(), a); err != nil {
		httputil.ServerError(w, r, err)
		return
	}

	// Broadcast progress to every group the user belongs to.
	go h.broadcastAnswerProgress(r.Context(), u.ID, q.ID)

	httputil.WriteJSON(w, http.StatusCreated, a)
}

// broadcastAnswerProgress fetches progress for each of the user's groups and
// broadcasts answer_progress (and guessing_unlocked if all answered).
func (h *Handler) broadcastAnswerProgress(ctx context.Context, userID, questionID string) {
	groups, err := h.store.GetUserGroups(ctx, userID)
	if err != nil {
		return
	}
	for _, g := range groups {
		progress, err := h.store.GetGroupAnswerProgress(ctx, g.ID, questionID)
		if err != nil {
			continue
		}
		h.broadcastToGroup(g.ID, httputil.SSEEvent("answer_progress", progress))
		if guessingUnlocked(progress) {
			h.broadcastToGroup(g.ID, httputil.SSEEvent("guessing_unlocked", struct{}{}))
		}
	}
}
