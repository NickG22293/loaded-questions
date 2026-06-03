package daily

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"candid/internal/httputil"
)

// StreamEvents opens an SSE connection for a group and streams events until
// the client disconnects. Events: answer_progress, guessing_unlocked, chat_message.
func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	if _, err := h.store.GetGroup(r.Context(), groupID); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "group not found", http.StatusNotFound)
			return
		}
		httputil.ServerError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 32)
	h.registerSSEClient(groupID, ch)
	defer h.unregisterSSEClient(groupID, ch)

	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			w.Write(msg)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (h *Handler) registerSSEClient(groupID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sseClients[groupID] = append(h.sseClients[groupID], ch)
}

func (h *Handler) unregisterSSEClient(groupID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.sseClients[groupID]
	for i, c := range clients {
		if c == ch {
			h.sseClients[groupID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.sseClients[groupID]) == 0 {
		delete(h.sseClients, groupID)
	}
}
