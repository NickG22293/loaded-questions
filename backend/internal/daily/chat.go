package daily

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"loaded-questions/internal/auth"
	"loaded-questions/internal/httputil"
)

// GetChat returns all chat messages for today's question thread in this group.
func (h *Handler) GetChat(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "chat unlocks when guessing opens", http.StatusForbidden)
		return
	}

	msgs, err := h.store.GetChatMessages(r.Context(), groupID, q.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if msgs == nil {
		msgs = []ChatMessage{}
	}
	httputil.WriteJSON(w, http.StatusOK, msgs)
}

// SendMessage posts a chat message and broadcasts it to SSE clients.
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	var body struct {
		Message string `json:"message"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil || body.Message == "" {
		http.Error(w, "message required", http.StatusBadRequest)
		return
	}

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
		http.Error(w, "chat unlocks when guessing opens", http.StatusForbidden)
		return
	}

	u := auth.GetUser(r.Context())
	m := &ChatMessage{
		ID:         uuid.New().String(),
		GroupID:    groupID,
		QuestionID: q.ID,
		UserID:     u.ID,
		Message:    body.Message,
	}
	if err := h.store.SendChatMessage(r.Context(), m); err != nil {
		httputil.ServerError(w, r, err)
		return
	}

	// Populate display fields for the SSE broadcast and response.
	m.DisplayName = u.DisplayName
	m.AvatarURL = u.AvatarURL

	h.broadcastToGroup(groupID, httputil.SSEEvent("chat_message", m))
	httputil.WriteJSON(w, http.StatusCreated, m)
}
