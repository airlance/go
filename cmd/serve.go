package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"git.emercury.dev/emercury/senderscore/api/internal/data"
	handler "git.emercury.dev/emercury/senderscore/api/internal/http"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure"
	"git.emercury.dev/emercury/senderscore/api/internal/usecase"
	"git.emercury.dev/emercury/senderscore/api/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd, args)
	},
}

func serve(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	cfg := config.Init(ctx)

	// Database
	db, err := infrastructure.NewDatabase(cfg.DB.DSN)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	sqlDB, _ := db.DB()
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logrus.WithError(err).Error("Error closing database")
		}
	}()

	// Data
	groupRepo := data.NewGroupRepository(db)
	ipRepo := data.NewIPRepository(db)
	historyRepo := data.NewHistoryRepository(db)
	scoreStatRepo := data.NewScoreStatRepository(db)

	// Use Cases
	groupUC := usecase.NewGroupUseCase(groupRepo, ipRepo, historyRepo, scoreStatRepo)
	ipUC := usecase.NewIPUseCase(groupRepo, ipRepo, historyRepo)

	// Handlers
	groupHandler := handler.NewGroupHandler(groupUC, ipUC)

	// Middleware
	validTokens := cfg.Auth.GetTokens()
	authMiddleware := infrastructure.AuthMiddleware(validTokens)

	if len(validTokens) == 0 {
		logrus.Warn("No API tokens configured. All protected endpoints will be inaccessible.")
	} else {
		logrus.Infof("API authentication enabled with %d token(s)", len(validTokens))
	}

	// Router
	router := gin.Default()
	handler.RegisterRoutes(router, groupHandler, authMiddleware)

	// Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logrus.Infof("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	logrus.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}
