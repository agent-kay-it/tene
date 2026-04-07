// Tene Cloud API Server entrypoint.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		FreeRPM:            100,
		ProRPM:             1000,
	}

	e := api.NewServer(cfg)

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
