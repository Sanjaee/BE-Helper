package main

import (
	"log"
	"order-service/internal/config"
	"order-service/internal/database"
	"order-service/internal/events"
	"order-service/internal/handlers"
	"order-service/internal/middleware"
	"order-service/internal/publisher"
	"order-service/internal/repository"
	"order-service/internal/services"

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

	// Initialize repositories
	orderRepo := repository.NewOrderRepository(db)
	broadcastRepo := repository.NewBroadcastRepository(db)

	// Initialize event publisher
	eventPublisher := publisher.NewEventPublisher(rabbitMQ)

	// Initialize services
	orderService := services.NewOrderService(orderRepo, broadcastRepo, eventPublisher)

	// Start event consumers in background
	go events.StartOrderBroadcastListener(rabbitMQ)
	go events.StartOrderAcceptedListener(rabbitMQ)

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(orderService)

	// Setup Gin router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestLogger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "order-service",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Order routes
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("/pending", orderHandler.GetPendingOrders)
			orders.PATCH("/:id/accept", orderHandler.AcceptOrder)
			orders.PATCH("/:id/on-the-way", orderHandler.UpdateToOnTheWay)
			orders.PATCH("/:id/arrived", orderHandler.UpdateToArrived)
			orders.GET("/client/:client_id", orderHandler.GetClientOrders)
			orders.GET("/provider/:provider_id", orderHandler.GetProviderOrders)
		}
	}

	log.Printf("🚀 Order Service running on port %s", cfg.Port)
	log.Println("📚 Available endpoints:")
	log.Println("  POST /api/v1/orders                    - Create new order")
	log.Println("  GET  /api/v1/orders/:id               - Get order by ID")
	log.Println("  GET  /api/v1/orders/pending           - Get pending orders")
	log.Println("  PATCH /api/v1/orders/:id/accept       - Accept order")
	log.Println("  PATCH /api/v1/orders/:id/on-the-way   - Update to on the way")
	log.Println("  PATCH /api/v1/orders/:id/arrived      - Update to arrived")
	log.Println("  GET  /api/v1/orders/client/:client_id - Get client orders")
	log.Println("  GET  /api/v1/orders/provider/:provider_id - Get provider orders")
	log.Println("  GET  /health                          - Health check")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
