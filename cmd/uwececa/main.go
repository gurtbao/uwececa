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

	"uwece.ca/app/config"
	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/shutdown"
	"uwece.ca/app/site"
)

func main() {
	if os.Getenv("SLOG_DEBUG") == "1" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("error running app", "error", err)
		os.Exit(1)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, os.Interrupt)
	<-s

	shutdown.Shutdown(10 * time.Second)
}

func run(ctx context.Context) error {
	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("config load error: %w", err)
	}

	if cfg.Core.Development {
		slog.Info("started in development mode")
	}

	db, err := db.New(cfg.DB.Location)
	if err != nil {
		return fmt.Errorf("error opening db: %w", err)
	}

	if err := db.RunMigrations(models.Migrations); err != nil {
		return fmt.Errorf("error running db migrations: %w", err)
	}

	mainsite := site.New(cfg, db)

	startServer(mainsite.Routes(), cfg.Core.Addr)

	return nil
}

func startServer(h http.Handler, addr string) {
	s := &http.Server{
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second, //nolint:mnd // fine
		ReadHeaderTimeout: 2 * time.Second,  //nolint:mnd // fine
		Handler:           h,
		Addr:              addr,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start http server", "error", err)
			os.Exit(1)
		}
	}()

	shutdown.AddFunc(func() {
		sCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(sCtx); err != nil {
			slog.Warn("failed to stop http server", "error", err)
		}

		slog.Debug("stopped http server")
	})

	slog.Info("started http server", "addr", addr)
}
