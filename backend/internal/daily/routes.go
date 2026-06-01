package daily

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"loaded-questions/internal/auth"
)

// Routes mounts all daily routes under auth.Required + ensureUser.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(auth.Required(h.authProvider))
	r.Use(h.ensureUser)

	r.Get("/today", h.GetToday)
	r.Post("/answers", h.SubmitAnswer)

	r.Get("/groups", h.ListGroups)
	r.Post("/groups", h.CreateGroup)
	r.Post("/groups/join", h.JoinGroup)

	r.Route("/groups/{groupId}", func(r chi.Router) {
		r.Get("/", h.GetGroup)
		r.Get("/events", h.StreamEvents)
		r.Get("/guesses", h.GetGuessTargets)
		r.Post("/guesses", h.SubmitGuesses)
		r.Get("/results", h.GetResults)
		r.Get("/leaderboard", h.GetLeaderboard)
		r.Get("/chat", h.GetChat)
		r.Post("/chat", h.SendMessage)
	})

	return r
}
