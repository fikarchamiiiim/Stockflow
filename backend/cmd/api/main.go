package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"stockflow/backend/internal/config"
	"stockflow/backend/internal/database"
	"stockflow/backend/internal/httpx"
	"stockflow/backend/internal/orders"
	"stockflow/backend/internal/products"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: newRouter(db),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", cfg.Port)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func newRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			httpx.WriteError(c, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database unavailable", nil)
			return
		}
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			httpx.WriteError(c, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database unavailable", nil)
			return
		}
		httpx.WriteJSON(c, http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	products.NewHandler(products.NewService(db)).Register(api)
	orders.NewHandler(orders.NewService(db)).Register(api)

	return r
}
