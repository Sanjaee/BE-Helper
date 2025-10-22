package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	UserServiceURL         = getEnv("USER_SERVICE_URL", "http://localhost:5001")
	OrderServiceURL        = getEnv("ORDER_SERVICE_URL", "http://localhost:5002")
	LocationServiceURL     = getEnv("LOCATION_SERVICE_URL", "http://localhost:5003")
	NotificationServiceURL = getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:5004")
	ChatServiceURL         = getEnv("CHAT_SERVICE_URL", "http://localhost:5005")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Log the service URLs being used
	log.Printf("🔗 User Service URL: %s", UserServiceURL)
	log.Printf("🔗 Order Service URL: %s", OrderServiceURL)
	log.Printf("🔗 Location Service URL: %s", LocationServiceURL)
	log.Printf("🔗 Notification Service URL: %s", NotificationServiceURL)
	log.Printf("🔗 Chat Service URL: %s", ChatServiceURL)

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

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	// User Service Routes
	userRoutes := r.Group("/api/v1")
	{
		// Health check for user service
		userRoutes.GET("/user/health", proxyToUserService("GET", "/health"))

		// Authentication routes
		authRoutes := userRoutes.Group("/auth")
		{
			authRoutes.POST("/register", proxyToUserService("POST", "/api/v1/auth/register"))
			authRoutes.POST("/login", proxyToUserService("POST", "/api/v1/auth/login"))
			authRoutes.POST("/verify-otp", proxyToUserService("POST", "/api/v1/auth/verify-otp"))
			authRoutes.POST("/resend-otp", proxyToUserService("POST", "/api/v1/auth/resend-otp"))
			authRoutes.POST("/refresh-token", proxyToUserService("POST", "/api/v1/auth/refresh-token"))
			authRoutes.POST("/google-oauth", proxyToUserService("POST", "/api/v1/auth/google-oauth"))
			authRoutes.POST("/request-reset-password", proxyToUserService("POST", "/api/v1/auth/request-reset-password"))
			authRoutes.POST("/verify-otp-reset-password", proxyToUserService("POST", "/api/v1/auth/verify-otp-reset-password"))
			authRoutes.POST("/verify-reset-password", proxyToUserService("POST", "/api/v1/auth/verify-reset-password"))
			authRoutes.POST("/check-user-status", proxyToUserService("POST", "/api/v1/auth/check-user-status"))
		}

		// Protected user routes
		userProtectedRoutes := userRoutes.Group("/user")
		{
			userProtectedRoutes.GET("/profile", proxyToUserService("GET", "/api/v1/user/profile"))
			userProtectedRoutes.PUT("/profile", proxyToUserService("PUT", "/api/v1/user/profile"))
		}

		// Rating routes
		ratings := userRoutes.Group("/ratings")
		{
			ratings.POST("", proxyToUserService("POST", "/api/v1/ratings"))
			ratings.GET("/my-ratings", proxyToUserService("GET", "/api/v1/ratings/my-ratings"))
			ratings.GET("/order/:order_id", proxyToUserService("GET", "/api/v1/ratings/order/:order_id"))
			ratings.GET("/order/:order_id/check", proxyToUserService("GET", "/api/v1/ratings/order/:order_id/check"))
			ratings.GET("/provider/:provider_id", proxyToUserService("GET", "/api/v1/ratings/provider/:provider_id"))
			ratings.GET("/provider/:provider_id/stats", proxyToUserService("GET", "/api/v1/ratings/provider/:provider_id/stats"))
		}
	}

	// Order Service Routes
	orderRoutes := r.Group("/api/v1")
	{
		// Health check for order service
		orderRoutes.GET("/order/health", proxyToOrderService("GET", "/health"))

		// Order routes
		orders := orderRoutes.Group("/orders")
		{
			orders.POST("", proxyToOrderService("POST", "/api/v1/orders"))
			orders.GET("/:id", proxyToOrderService("GET", "/api/v1/orders/:id"))
			orders.GET("/pending", proxyToOrderService("GET", "/api/v1/orders/pending"))
			orders.PATCH("/:id/accept", proxyToOrderService("PATCH", "/api/v1/orders/:id/accept"))
			orders.PATCH("/:id/on-the-way", proxyToOrderService("PATCH", "/api/v1/orders/:id/on-the-way"))
			orders.PATCH("/:id/arrived", proxyToOrderService("PATCH", "/api/v1/orders/:id/arrived"))
			orders.PATCH("/:id/start-job", proxyToOrderService("PATCH", "/api/v1/orders/:id/start-job"))
			orders.PATCH("/:id/complete-job", proxyToOrderService("PATCH", "/api/v1/orders/:id/complete-job"))
			orders.PATCH("/:id/cancel", proxyToOrderService("PATCH", "/api/v1/orders/:id/cancel"))
			orders.GET("/client/:client_id", proxyToOrderService("GET", "/api/v1/orders/client/:client_id"))
			orders.GET("/provider/:provider_id", proxyToOrderService("GET", "/api/v1/orders/provider/:provider_id"))
		}
	}

	// Location Service Routes
	locationRoutes := r.Group("/api/v1")
	{
		// Health check for location service
		locationRoutes.GET("/location/health", proxyToLocationService("GET", "/health"))

		// Location routes
		locations := locationRoutes.Group("/locations")
		{
			locations.POST("/track", proxyToLocationService("POST", "/api/v1/locations/track"))
			locations.GET("/order/:order_id", proxyToLocationService("GET", "/api/v1/locations/order/:order_id"))
			locations.GET("/order/:order_id/history", proxyToLocationService("GET", "/api/v1/locations/order/:order_id/history"))
			locations.GET("/provider/:order_id", proxyToLocationService("GET", "/api/v1/locations/provider/:order_id"))
		}
	}

	// Chat Service Routes
	chatRoutes := r.Group("/api/v1")
	{
		// Health check for chat service
		chatRoutes.GET("/chat/health", proxyToChatService("GET", "/health"))

		// Chat routes
		chats := chatRoutes.Group("/chats")
		{
			chats.POST("/messages", proxyToChatService("POST", "/api/v1/chats/messages"))
			chats.GET("/order/:order_id", proxyToChatService("GET", "/api/v1/chats/order/:order_id"))
			chats.GET("/order/:order_id/unread", proxyToChatService("GET", "/api/v1/chats/order/:order_id/unread"))
			chats.PATCH("/order/:order_id/read", proxyToChatService("PATCH", "/api/v1/chats/order/:order_id/read"))
		}
	}

	// WebSocket Routes
	wsRoutes := r.Group("/api/v1/ws")
	{
		// Order WebSocket - proxy to order service
		wsRoutes.GET("/orders/:order_id", proxyWebSocketToOrderService())

		// Location WebSocket - proxy to location service
		wsRoutes.GET("/locations/:order_id", proxyWebSocketToLocationService())

		// Chat WebSocket - proxy to chat service
		wsRoutes.GET("/chat/:order_id", proxyWebSocketToChatService())
	}

	log.Println("🚀 API Gateway running on http://localhost:5000")
	log.Println("📚 Available endpoints:")
	log.Println("  POST /api/v1/auth/register     - Register new user")
	log.Println("  POST /api/v1/auth/login        - Login user")
	log.Println("  POST /api/v1/auth/verify-otp   - Verify OTP")
	log.Println("  POST /api/v1/auth/resend-otp   - Resend OTP")
	log.Println("  POST /api/v1/auth/refresh-token - Refresh JWT token")
	log.Println("  POST /api/v1/auth/google-oauth - Google OAuth login")
	log.Println("  POST /api/v1/auth/request-reset-password - Request password reset")
	log.Println("  POST /api/v1/auth/verify-reset-password - Verify reset password")
	log.Println("  GET  /api/v1/user/profile      - Get user profile (protected)")
	log.Println("  PUT  /api/v1/user/profile      - Update user profile (protected)")
	log.Println("  POST /api/v1/orders            - Create new order")
	log.Println("  GET  /api/v1/orders/:id        - Get order by ID")
	log.Println("  PATCH /api/v1/orders/:id/accept - Accept order")
	log.Println("  PATCH /api/v1/orders/:id/on-the-way - Update to on the way")
	log.Println("  PATCH /api/v1/orders/:id/arrived - Update to arrived")
	log.Println("  PATCH /api/v1/orders/:id/start-job - Start job")
	log.Println("  PATCH /api/v1/orders/:id/complete-job - Complete job")
	log.Println("  PATCH /api/v1/orders/:id/cancel - Cancel order")
	log.Println("  GET  /api/v1/orders/client/:client_id - Get client orders")
	log.Println("  GET  /api/v1/orders/provider/:provider_id - Get provider orders")
	log.Println("  POST /api/v1/locations/track   - Update location")
	log.Println("  GET  /api/v1/locations/order/:order_id - Get order location")
	log.Println("  GET  /api/v1/locations/order/:order_id/history - Get location history")
	log.Println("  GET  /api/v1/locations/provider/:order_id - Get provider location")
	log.Println("  WS   /api/v1/ws/orders/:order_id - WebSocket for order updates")
	log.Println("  WS   /api/v1/ws/locations/:order_id - WebSocket for location updates")
	log.Println("  GET  /health                   - Health check")

	r.Run(":5000")
}

// proxyToUserService creates a proxy handler for user service
func proxyToUserService(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// Replace URL parameters with actual values
		actualPath := path
		for _, param := range c.Params {
			actualPath = strings.Replace(actualPath, ":"+param.Key, param.Value, -1)
		}

		// Add query parameters
		queryParams := c.Request.URL.Query()
		if len(queryParams) > 0 {
			actualPath += "?" + queryParams.Encode()
		}

		// Create new request to user service
		url := UserServiceURL + actualPath
		req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Make request to user service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to connect to user service at %s: %v", url, err)
			c.JSON(500, gin.H{"error": "User service unavailable", "details": err.Error()})
			return
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Return response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

// proxyToOrderService creates a proxy handler for order service
func proxyToOrderService(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// Replace URL parameters with actual values
		actualPath := path
		for _, param := range c.Params {
			actualPath = strings.Replace(actualPath, ":"+param.Key, param.Value, -1)
		}

		// Add query parameters
		queryParams := c.Request.URL.Query()
		if len(queryParams) > 0 {
			actualPath += "?" + queryParams.Encode()
		}

		// Create new request to order service
		url := OrderServiceURL + actualPath
		req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Make request to order service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to connect to order service at %s: %v", url, err)
			c.JSON(500, gin.H{"error": "Order service unavailable", "details": err.Error()})
			return
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Return response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

// proxyToLocationService creates a proxy handler for location service
func proxyToLocationService(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// Replace URL parameters with actual values
		actualPath := path
		for _, param := range c.Params {
			actualPath = strings.Replace(actualPath, ":"+param.Key, param.Value, -1)
		}

		// Add query parameters
		queryParams := c.Request.URL.Query()
		if len(queryParams) > 0 {
			actualPath += "?" + queryParams.Encode()
		}

		// Create new request to location service
		url := LocationServiceURL + actualPath
		req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Make request to location service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to connect to location service at %s: %v", url, err)
			c.JSON(500, gin.H{"error": "Location service unavailable", "details": err.Error()})
			return
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Return response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// proxyWebSocketToOrderService proxies WebSocket connections to order service
func proxyWebSocketToOrderService() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")

		// Upgrade client connection
		clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade client connection: %v", err)
			return
		}
		defer clientConn.Close()

		// Connect to order service
		wsURL := strings.Replace(OrderServiceURL, "http://", "ws://", 1)
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
		serviceURL, _ := url.Parse(wsURL + "/api/v1/ws/orders/" + orderID)

		serviceConn, _, err := websocket.DefaultDialer.Dial(serviceURL.String(), nil)
		if err != nil {
			log.Printf("Failed to connect to order service WebSocket: %v", err)
			return
		}
		defer serviceConn.Close()

		log.Printf("✅ WebSocket proxy established for order: %s", orderID)

		// Proxy bidirectionally
		go proxyWebSocket(clientConn, serviceConn)
		proxyWebSocket(serviceConn, clientConn)
	}
}

// proxyWebSocketToLocationService proxies WebSocket connections to location service
func proxyWebSocketToLocationService() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")

		// Upgrade client connection
		clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade client connection: %v", err)
			return
		}
		defer clientConn.Close()

		// Connect to location service
		wsURL := strings.Replace(LocationServiceURL, "http://", "ws://", 1)
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
		serviceURL, _ := url.Parse(wsURL + "/api/v1/ws/locations/" + orderID)

		serviceConn, _, err := websocket.DefaultDialer.Dial(serviceURL.String(), nil)
		if err != nil {
			log.Printf("Failed to connect to location service WebSocket: %v", err)
			return
		}
		defer serviceConn.Close()

		log.Printf("✅ WebSocket proxy established for location: %s", orderID)

		// Proxy bidirectionally
		go proxyWebSocket(clientConn, serviceConn)
		proxyWebSocket(serviceConn, clientConn)
	}
}

// proxyWebSocketToChatService proxies WebSocket connections to chat service
func proxyWebSocketToChatService() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")
		userID := c.Query("user_id")

		// Upgrade client connection
		clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade client connection: %v", err)
			return
		}
		defer clientConn.Close()

		// Connect to chat service
		wsURL := strings.Replace(ChatServiceURL, "http://", "ws://", 1)
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
		serviceURL, _ := url.Parse(wsURL + "/api/v1/ws/chat/" + orderID + "?user_id=" + userID)

		serviceConn, _, err := websocket.DefaultDialer.Dial(serviceURL.String(), nil)
		if err != nil {
			log.Printf("Failed to connect to chat service WebSocket: %v", err)
			return
		}
		defer serviceConn.Close()

		log.Printf("✅ WebSocket proxy established for chat: %s", orderID)

		// Proxy bidirectionally
		go proxyWebSocket(clientConn, serviceConn)
		proxyWebSocket(serviceConn, clientConn)
	}
}

// proxyToChatService creates a proxy handler for chat service
func proxyToChatService(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// Replace URL parameters with actual values
		actualPath := path
		for _, param := range c.Params {
			actualPath = strings.Replace(actualPath, ":"+param.Key, param.Value, -1)
		}

		// Add query parameters
		queryParams := c.Request.URL.Query()
		if len(queryParams) > 0 {
			actualPath += "?" + queryParams.Encode()
		}

		// Create new request to chat service
		url := ChatServiceURL + actualPath
		req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Make request to chat service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to connect to chat service at %s: %v", url, err)
			c.JSON(500, gin.H{"error": "Chat service unavailable", "details": err.Error()})
			return
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Return response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

// proxyWebSocket forwards messages between two WebSocket connections
func proxyWebSocket(src, dst *websocket.Conn) {
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		err = dst.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Failed to write message: %v", err)
			break
		}
	}
}
