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
	"loaded-questions/internal/auth/supabase"
	"loaded-questions/internal/sessions"
)

func main() {
	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	jwtIssuer := os.Getenv("SUPABASE_JWT_ISSUER")
	if jwksURL == "" || jwtIssuer == "" {
		log.Println("warn: SUPABASE_JWKS_URL or SUPABASE_JWT_ISSUER not set — starting in sessions-only mode (no JWT auth)")
	} else {
		authProvider, err := supabase.NewProvider(jwksURL, jwtIssuer)
		if err != nil {
			log.Fatalf("failed to initialise auth provider: %v", err)
		}
		log.Println("auth provider initialised (Supabase JWKS)")
		_ = authProvider // used when daily routes are wired: r.Mount("/api/daily", daily.Routes(dailyHandler, authProvider))
	}

	s := sessions.NewMemoryStore()
	h := sessions.New(s)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(corsMiddleware)

	r.Mount("/api/sessions", h.Routes())

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
