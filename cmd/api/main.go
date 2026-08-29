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
	"github.com/labstack/echo/v4"

	httpAdapter "user-management/internal/adapter/http"
	"user-management/internal/adapter/mongo"
	"user-management/internal/config"
	"user-management/internal/domain"
	"user-management/internal/usecase"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, disconnect, err := mongo.Connect(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		slog.Error("mongo connect failed", "err", err)
		os.Exit(1)
	}

	repo := mongo.NewUserRepository(db)
	uc := usecase.NewUserUsecase(repo, cfg.JWTSecret)
	h := httpAdapter.NewHandler(uc)

	e := echo.New()
	e.Use(httpAdapter.RequestLogger())

	e.POST("/register", h.Register)
	e.POST("/login", h.Login)

	authGroup := e.Group("", httpAdapter.JWTMiddleware(cfg.JWTSecret))
	authGroup.GET("/users/:id", h.GetByID)
	authGroup.GET("/users", h.List)
	authGroup.PUT("/users/:id", h.Update)
	authGroup.DELETE("/users/:id", h.Delete)

	// background task: log user count every 10s
	ctx, cancel := context.WithCancel(context.Background())
	go func(repo domain.UserRepository) {
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
