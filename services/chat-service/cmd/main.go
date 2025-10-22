package main

import (
	"chat-service/internal/config"
	"chat-service/internal/database"
	"chat-service/internal/events"
	"chat-service/internal/handlers"
	"chat-service/internal/repository"
	"chat-service/internal/services"
	"chat-service/internal/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Initialize RabbitMQ
	rabbitMQ, err := events.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	log.Println("✅ RabbitMQ connected successfully")

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize repositories
	chatRepo := repository.NewChatRepository(db)

	// Initialize services
	chatService := services.NewChatService(chatRepo, rabbitMQ, hub)

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(chatService)
	wsHandler := handlers.NewWebSocketHandler(hub, chatService)

	// Setup router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "chat-service",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Chat routes
		chats := api.Group("/chats")
		{
			chats.POST("/messages", chatHandler.SendMessage)
			chats.GET("/order/:order_id", chatHandler.GetChatHistory)
			chats.GET("/order/:order_id/unread", chatHandler.GetUnreadCount)
			chats.PATCH("/order/:order_id/read", chatHandler.MarkAsRead)
		}

		// WebSocket route
		api.GET("/ws/chat/:order_id", wsHandler.HandleWebSocket)
	}

	log.Printf("🚀 Chat Service running on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}

