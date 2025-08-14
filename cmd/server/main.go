package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chatmix-backend/internal/config"
	"chatmix-backend/internal/handler"
	"chatmix-backend/internal/repository"
	"chatmix-backend/internal/router"
	"chatmix-backend/internal/service"
	"chatmix-backend/pkg/utils"

	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	configPath := getConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("====> Database URI: ", cfg.Database.URI)

	// Initialize logger
	logger := utils.NewLogger(cfg)
	logger.Info("Starting ChatMix Backend Server")

	// Initialize database
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			logger.WithError(err).Error("Failed to close database connection")
		}
	}()

	logger.Info("Connected to MongoDB successfully")

	// Initialize services
	userService := service.NewUserService(db.UserRepo, cfg, logger)
	authService := service.NewAuthService(db.UserRepo, db.RefreshTokenRepo, db.SessionRepo, db.CaptchaRepo, cfg, logger)
	chatService := service.NewChatService(cfg, logger)

	// Initialize handlers
	httpHandler := handler.NewHTTPHandler(userService, logger)
	authHandler := handler.NewUserHandler(authService, userService, logger)
	chatHandler := handler.NewChatHandler(chatService, authService)

	// Initialize router
	appRouter := router.NewRouter(cfg, logger, httpHandler, authHandler, authService, chatHandler)
	routes := appRouter.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      routes,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.WithFields(logrus.Fields{
			"addr":          server.Addr,
			"read_timeout":  cfg.Server.ReadTimeout,
			"write_timeout": cfg.Server.WriteTimeout,
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	logger.Info("Available routes:")
	for _, route := range appRouter.ListRoutes() {
		logger.WithField("route", route).Info("Route registered")
	}

	logger.WithFields(logrus.Fields{
		"version":       "1.0.0",
		"server_addr":   server.Addr,
		"database_name": cfg.Database.Name,
		"log_level":     cfg.Logging.Level,
	}).Info("ChatMix Backend Server started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}

	possiblePaths := []string{
		"configs/config.yaml",
		"../../configs/config.yaml",
		"/etc/chatmix/config.yaml",
		"config.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "configs/config.yaml"
}
