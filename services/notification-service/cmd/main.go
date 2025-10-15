package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/config"
	"notification-service/internal/consumers"
	"notification-service/internal/handlers"
	"notification-service/internal/middleware"
	"notification-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, using system environment variables")
	}

	// Initialize logger
	logger := logrus.New()
	if os.Getenv("LOG_FORMAT") == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		if level, err := logrus.ParseLevel(logLevel); err == nil {
			logger.SetLevel(level)
		}
	}

	logger.Info("🚀 Starting Notification Service...")

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Initialize services
	emailService, err := services.NewEmailService(cfg)
	if err != nil {
		logger.Fatalf("❌ Failed to initialize email service: %v", err)
	}

	// Initialize notification service
	notificationService := services.NewNotificationService(emailService, logger)

	// Initialize event consumer
	eventConsumer, err := consumers.NewEventConsumer(cfg, notificationService, logger)
	if err != nil {
		logger.Fatalf("❌ Failed to initialize event consumer: %v", err)
	}

	// Start event consumer
	if err := eventConsumer.Start(); err != nil {
		logger.Fatalf("❌ Failed to start event consumer: %v", err)
	}
	logger.Info("✅ Event consumer started successfully")

	// Initialize Gin router
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck(logger))

	// Notification endpoints
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/notifications/send", notificationHandler.SendNotification)
		api.GET("/notifications/history", notificationHandler.GetNotificationHistory)
		api.GET("/notifications/templates", notificationHandler.GetTemplates)
		api.POST("/notifications/templates", notificationHandler.CreateTemplate)
	}

	// Start HTTP server
	port := cfg.Port
	if port == "" {
		port = "5004"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("🌐 HTTP server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("❌ Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down Notification Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop event consumer
	if err := eventConsumer.Stop(); err != nil {
		logger.Errorf("❌ Failed to stop event consumer: %v", err)
	}

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("❌ Failed to shutdown HTTP server: %v", err)
	}

	logger.Info("✅ Notification Service stopped gracefully")
}
