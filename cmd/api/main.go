package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"user-management/internal/config"
	"user-management/internal/connector"
	"user-management/internal/handler"
	userhandler "user-management/internal/handler/user"
	userrepo "user-management/internal/repository/user"
	userusecase "user-management/internal/usecase/user"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, disconnect, err := connector.Mongo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		slog.Error("mongo connect failed", "err", err)
		os.Exit(1)
	}

	repo := userrepo.NewMongo(db)
	uc := userusecase.New(repo, cfg.JWTSecret)
	h := userhandler.NewHandler(uc)

	e := handler.NewRouter(h, cfg.JWTSecret)

	// background task: log user count every 10s
	ctx, cancel := context.WithCancel(context.Background())
	go func(repo userrepo.Repository) {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, err := repo.Count(ctx)
				if err != nil {
					slog.Error("count failed", "err", err)
					continue
				}
				slog.Info("user count", "total", count)
			}
		}
	}(repo)

	// start server
	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel() // stop background goroutine

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	_ = disconnect(shutdownCtx)
	slog.Info("server stopped gracefully")
}
