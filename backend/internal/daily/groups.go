package daily

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"loaded-questions/internal/auth"
	"loaded-questions/internal/httputil"
)

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	u := auth.GetUser(r.Context())
	groups, err := h.store.GetUserGroups(r.Context(), u.ID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if groups == nil {
		groups = []Group{}
	}
	httputil.WriteJSON(w, http.StatusOK, groups)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	u := auth.GetUser(r.Context())
	g := &Group{
		ID:         uuid.New().String(),
		Name:       body.Name,
		InviteCode: generateInviteCode(),
		CreatedBy:  u.ID,
	}
	if err := h.store.CreateGroup(r.Context(), g); err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, g)
}

func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		InviteCode string `json:"invite_code"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil || body.InviteCode == "" {
		http.Error(w, "invite_code required", http.StatusBadRequest)
		return
	}

	g, err := h.store.GetGroupByInviteCode(r.Context(), body.InviteCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "group not found", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	u := auth.GetUser(r.Context())
	if err := h.store.JoinGroup(r.Context(), g.ID, u.ID); err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, g)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	g, err := h.store.GetGroup(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "group not found", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	members, err := h.store.GetGroupMembers(r.Context(), groupID)
	if err != nil {
		httputil.ServerError(w, r, err)
		return
	}
	if members == nil {
		members = []GroupMember{}
	}

	u := auth.GetUser(r.Context())
	var progress AnswerProgress
	var userAnswered, userGuessed bool

	q, qErr := h.store.GetTodayQuestion(r.Context())
	if qErr == nil {
		progress, _ = h.store.GetGroupAnswerProgress(r.Context(), groupID, q.ID)
		_, aErr := h.store.GetUserAnswer(r.Context(), u.ID, q.ID)
		userAnswered = aErr == nil
		userGuessed, _ = h.store.HasUserSubmittedGuesses(r.Context(), u.ID, groupID, q.ID)
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"group":             g,
		"members":           members,
		"progress":          progress,
		"guessing_unlocked": guessingUnlocked(progress),
		"user_answered":     userAnswered,
		"user_guessed":      userGuessed,
	})
}
