package main

import (
	"context"
	"errors"
	"fmt"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"

	_ "scaffold-api/docs"
	"scaffold-api/internal/config"
	"scaffold-api/internal/db"
	"scaffold-api/internal/db/query"
	httpapi "scaffold-api/internal/http"
	"scaffold-api/internal/logger"
	"scaffold-api/internal/service"
)

//go:generate swag init --parseInternal -g main.go -o docs

// @title Scaffold API Users API
// @version 1.0
// @description Online Swagger documentation for the users CRUD service.
// @BasePath /
// @schemes http https
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	appLogger, err := logger.New(cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	store := db.NewStore(pool, query.New(pool))
	userService := service.NewUserService(store)
	handler := httpapi.NewHandler(cfg, appLogger, userService)

	server := &stdhttp.Server{
		Addr:         cfg.Address(),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		appLogger.Info("http server started", "addr", cfg.Address())
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, stdhttp.ErrServerClosed) {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		appLogger.Info("shutdown signal received")
	case serveErr := <-errCh:
		if serveErr != nil {
			return fmt.Errorf("http server crashed: %w", serveErr)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}
