package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"loaded-questions/handlers"
	authmiddleware "loaded-questions/middleware"
	"loaded-questions/store"
)

func main() {
	s := store.NewMemoryStore()
	h := handlers.New(s)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		// Public endpoints — no auth cookie required.
		r.Post("/lobbies", h.CreateLobby)
		r.Post("/lobbies/{id}/join", h.JoinLobby)
		r.Get("/lobbies/{id}", h.GetLobby)
		r.Get("/lobbies/{id}/events", h.StreamEvents)
		r.Get("/lobbies/{id}/game", h.GetGame)

		// Authenticated endpoints — all game actions use the lobby ID so
		// clients don't need to track a separate game ID.
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.Auth(s))
			r.Post("/lobbies/{id}/start", h.StartGame)
			r.Post("/lobbies/{id}/question", h.SubmitQuestion)
			r.Post("/lobbies/{id}/answer", h.SubmitAnswer)
			r.Post("/lobbies/{id}/assign", h.AssignAnswer)
			r.Post("/lobbies/{id}/lock", h.LockAssignments)
			r.Post("/lobbies/{id}/next", h.NextRound)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // SSE streams need no write timeout
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Player-Token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
