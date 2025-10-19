package main

import (
	"location-service/internal/config"
	"location-service/internal/database"
	"location-service/internal/events"
	"location-service/internal/handlers"
	"location-service/internal/middleware"
	"location-service/internal/publisher"
	"location-service/internal/repository"
	"location-service/internal/services"
	ws "location-service/internal/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate database
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize RabbitMQ
	rabbitMQ, err := events.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitMQ.Close()

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize repositories
	locationRepo := repository.NewLocationRepository(db)

	// Initialize event publisher
	eventPublisher := publisher.NewEventPublisher(rabbitMQ)

	// Initialize services with WebSocket hub
	locationService := services.NewLocationService(locationRepo, eventPublisher, hub)

	// Start event consumers in background
	go events.StartLocationTrackingListener(rabbitMQ)

	// Initialize handlers
	locationHandler := handlers.NewLocationHandler(locationService)
	wsHandler := handlers.NewWebSocketHandler(hub)

	// Setup Gin router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "location-service",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Location routes
		locations := api.Group("/locations")
		{
			locations.POST("/track", locationHandler.UpdateLocation)
			locations.GET("/order/:order_id", locationHandler.GetOrderLocation)
			locations.GET("/order/:order_id/history", locationHandler.GetLocationHistory)
			locations.GET("/provider/:order_id", locationHandler.GetProviderLocation)
		}

		// WebSocket routes
		api.GET("/ws/locations/:order_id", wsHandler.HandleLocationWebSocket)
	}

	log.Printf("🚀 Location Service running on port %s", cfg.Port)
	log.Println("📚 Available endpoints:")
	log.Println("  POST /api/v1/locations/track              - Update location")
	log.Println("  GET  /api/v1/locations/order/:order_id   - Get current location")
	log.Println("  GET  /api/v1/locations/order/:order_id/history - Get location history")
	log.Println("  GET  /api/v1/locations/provider/:order_id - Get provider location")
	log.Println("  GET  /api/v1/ws/locations/:order_id      - WebSocket for real-time location updates")
	log.Println("  GET  /health                             - Health check")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
