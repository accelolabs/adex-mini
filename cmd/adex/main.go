package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/accelolabs/adex-mini/internal/auction"
	"github.com/accelolabs/adex-mini/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.MustLoad()

	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("parse LOG_LEVEL: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	startupCtx, cancelStartup := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	if err := pool.Ping(startupCtx); err != nil {
		cancelStartup()
		return fmt.Errorf("connect to postgres: %w", err)
	}
	cancelStartup()

	repository := auction.NewPostgres(pool)
	service := auction.NewService(repository, auction.NewMockDSPClient(), cfg.AuctionTimeout)
	handler := auction.NewHandler(service, repository, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", "address", cfg.HTTPAddr)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutting down HTTP server")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverError; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("stop HTTP server: %w", err)
	}
	return nil
}
