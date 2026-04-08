// Tene Cloud API Server entrypoint.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/tomo-kay/tene/internal/api"
)

func main() {
	cfg := api.Config{
		Port:               envOr("PORT", "8080"),
		JWTSecret:          envRequired("JWT_SECRET"),
		GitHubClientID:     envOr("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: envOr("GITHUB_CLIENT_SECRET", ""),
		CallbackBase:       envOr("CALLBACK_BASE", "http://127.0.0.1:8080"),
		DashboardURL:       envOr("DASHBOARD_URL", "https://app.tene.sh"),
		DatabaseURL:        envOr("DATABASE_URL", ""),
		FreeRPM:            100,
		ProRPM:             1000,
	}

	// Run database migrations if DATABASE_URL is set
	if cfg.DatabaseURL != "" {
		runMigrations(cfg.DatabaseURL)
	}

	e, cleanup, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}
	defer cleanup()

	// Graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server exited cleanly")
}

// runMigrations applies pending database migrations from the migrations/ directory.
func runMigrations(databaseURL string) {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		log.Fatalf("migration init: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("migration close source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("migration close db: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration up: %v", err)
	}
	log.Println("database migrations applied")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}
