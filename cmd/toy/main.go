package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/QuirkyQuestor/toy-service/internal/api"
	"github.com/QuirkyQuestor/toy-service/internal/store"
)

// Inits the app
// inits logger on load
func init() {
	// Configure slog to use JSON handler and output to stdout.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := store.DBConnect(stop)
	if err != nil {
		slog.Error("Database is not available; refusing to start", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	serverSettings := api.NewServer(db)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           serverSettings.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("Starting listening on", "Addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server start has failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("Shutdown signal received")
	serverSettings.ShutdownStarted.Store(true)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server has failed to shutdown before the timeout", "error", err.Error())
	}
	slog.Info("Server gracefully stopped")
}
