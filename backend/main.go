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
	"github.com/joho/godotenv"
	"candid/internal/auth/supabase"
	"candid/internal/daily"
	"candid/internal/sessions"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(corsMiddleware)

	// Sessions (always on — no auth required)
	s := sessions.NewMemoryStore()
	r.Mount("/api/sessions", sessions.New(s).Routes())

	// Daily (requires both auth env vars and a database URL)
	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	jwtIssuer := os.Getenv("SUPABASE_JWT_ISSUER")
	databaseURL := os.Getenv("DATABASE_URL")

	switch {
	case databaseURL == "":
		log.Println("warn: DATABASE_URL not set — daily routes disabled")
	case jwksURL == "" || jwtIssuer == "":
		log.Println("warn: SUPABASE_JWKS_URL or SUPABASE_JWT_ISSUER not set — daily routes disabled")
	default:
		authProvider, err := supabase.NewProvider(jwksURL, jwtIssuer)
		if err != nil {
			log.Fatalf("failed to initialise auth provider: %v", err)
		}
		dailyStore, err := daily.NewPostgresStore(ctx, databaseURL)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		r.Mount("/api/daily", daily.New(ctx, dailyStore, authProvider).Routes())
		log.Println("daily routes mounted (Supabase JWKS + Postgres)")
	}

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

	cancel() // stop daily timer goroutine

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
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
