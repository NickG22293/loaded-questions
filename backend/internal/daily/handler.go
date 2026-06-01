package daily

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"
	"sync"
	"time"

	"loaded-questions/internal/auth"
	"loaded-questions/internal/httputil"
)

const inviteCodeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Handler holds shared state for all daily HTTP handlers.
type Handler struct {
	store        DailyStore
	authProvider auth.Provider

	mu         sync.RWMutex
	sseClients map[string][]chan []byte // groupID → open SSE channels
}

// New creates a Handler and starts the background 6pm guessing-unlock timer.
func New(ctx context.Context, store DailyStore, ap auth.Provider) *Handler {
	h := &Handler{
		store:        store,
		authProvider: ap,
		sseClients:   make(map[string][]chan []byte),
	}
	go h.runDailyTimer(ctx)
	return h
}

// ensureUser upserts the authenticated user into the users table so every
// daily handler can assume a DB row exists.
func (h *Handler) ensureUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := auth.GetUser(r.Context())
		if u == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := h.store.UpsertUser(r.Context(), &User{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
		}); err != nil {
			httputil.ServerError(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// broadcastToGroup sends an SSE event to all active clients for a group.
func (h *Handler) broadcastToGroup(groupID string, event []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.sseClients[groupID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// runDailyTimer fires broadcastGuessingUnlockedAll every day at 18:00 local time.
func (h *Handler) runDailyTimer(ctx context.Context) {
	for {
		now := time.Now()
		sixPM := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.Local)
		if !now.Before(sixPM) {
			sixPM = sixPM.Add(24 * time.Hour)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(sixPM)):
			h.broadcastGuessingUnlockedAll()
		}
	}
}

// broadcastGuessingUnlockedAll notifies every connected SSE client that
// guessing has opened. The frontend handles duplicate events gracefully.
func (h *Handler) broadcastGuessingUnlockedAll() {
	event := httputil.SSEEvent("guessing_unlocked", struct{}{})
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, clients := range h.sseClients {
		for _, ch := range clients {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// guessingUnlocked returns true when either all group members have answered
// or the server clock has passed 18:00 local time.
func guessingUnlocked(p AnswerProgress) bool {
	if p.Total > 0 && p.Answered >= p.Total {
		return true
	}
	now := time.Now()
	sixPM := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.Local)
	return !now.Before(sixPM)
}

func generateInviteCode() string {
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		b[i] = inviteCodeChars[n.Int64()]
	}
	return string(b)
}
