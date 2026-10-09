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

	"github.com/example/shop-cabinet/internal/auth"
	"github.com/example/shop-cabinet/internal/config"
	"github.com/example/shop-cabinet/internal/handler"
	"github.com/example/shop-cabinet/internal/repository"
	"github.com/example/shop-cabinet/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := repository.Connect(ctx, cfg.MongoURI)
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error("disconnect from mongo", "error", err)
		}
	}()

	db := client.Database(cfg.MongoDatabase)
	if err := repository.EnsureIndexes(ctx, db); err != nil {
		return err
	}

	orderService := service.NewOrderService(repository.NewOrderRepository(db))
	userService := service.NewUserService(repository.NewUserRepository(db))

	router := handler.NewRouter(handler.Deps{
		Orders:   orderService,
		Users:    userService,
		Sessions: auth.NewSessions([]byte(cfg.SessionSecret), cfg.SessionTTL),
		WebDir:   cfg.WebDir,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server started", "addr", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
