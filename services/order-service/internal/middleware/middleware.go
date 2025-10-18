package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Provider-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RequestLogger middleware to log request body
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		// Restore the request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Log the request
		fmt.Printf("🔍 [%s] %s %s\n", c.Request.Method, c.Request.URL.Path, string(body))

		c.Next()
	}
}
