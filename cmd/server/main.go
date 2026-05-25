package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/nglambertjr/plate-review-api/internal/config"
	"github.com/nglambertjr/plate-review-api/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Load .env if present; silently ignored in production where env vars
	// come from the platform (Fly.io secrets, etc.).
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx := context.Background()
	handler, err := server.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	addr := ":" + cfg.Port
	slog.Info("listening", "addr", addr)
	return http.ListenAndServe(addr, handler)
}
