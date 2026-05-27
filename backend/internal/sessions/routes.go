package sessions

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes returns a router with all session endpoints mounted.
// Intended to be mounted at /api/sessions in main.go.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/lobbies", h.CreateLobby)
	r.Post("/lobbies/{id}/join", h.JoinLobby)
	r.Get("/lobbies/{id}", h.GetLobby)
	r.Get("/lobbies/{id}/events", h.StreamEvents)
	r.Get("/lobbies/{id}/game", h.GetGame)

	r.Group(func(r chi.Router) {
		r.Use(Auth(h.store))
		r.Post("/lobbies/{id}/start", h.StartGame)
		r.Post("/lobbies/{id}/question", h.SubmitQuestion)
		r.Post("/lobbies/{id}/answer", h.SubmitAnswer)
		r.Post("/lobbies/{id}/assign", h.AssignAnswer)
		r.Post("/lobbies/{id}/lock", h.LockAssignments)
		r.Post("/lobbies/{id}/next", h.NextRound)
	})

	return r
}
