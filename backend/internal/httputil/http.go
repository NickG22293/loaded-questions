package httputil

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ServerError logs the error with request context and returns 500 to the client.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("handler error", "method", r.Method, "path", r.URL.Path, "error", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

func SSEEvent(eventName string, data any) []byte {
	d, _ := json.Marshal(data)
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, d))
}
