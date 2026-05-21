package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// StreamEvents opens an SSE connection for a lobby and streams events until the client disconnects.
func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")

	if _, err := h.store.GetLobby(lobbyID); err != nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
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
	h.store.RegisterSSEClient(lobbyID, ch)
	defer h.store.UnregisterSSEClient(lobbyID, ch)

	// Send a connected event so the client knows the stream is live.
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
			// Keepalive comment to prevent proxy timeouts.
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
