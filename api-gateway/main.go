package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	UserServiceURL         = getEnv("USER_SERVICE_URL", "http://localhost:5001")
	OrderServiceURL        = getEnv("ORDER_SERVICE_URL", "http://localhost:5002")
	LocationServiceURL     = getEnv("LOCATION_SERVICE_URL", "http://localhost:5003")
	NotificationServiceURL = getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:5004")
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
		}
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
	log.Println("  GET  /api/v1/orders/client/:client_id - Get client orders")
	log.Println("  GET  /api/v1/orders/provider/:provider_id - Get provider orders")
	log.Println("  POST /api/v1/locations/track   - Update location")
	log.Println("  GET  /api/v1/locations/order/:order_id - Get order location")
	log.Println("  GET  /api/v1/locations/order/:order_id/history - Get location history")
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
