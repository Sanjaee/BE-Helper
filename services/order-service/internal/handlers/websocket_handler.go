package handlers

import (
	"log"
	"net/http"
	ws "order-service/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleOrderWebSocket handles WebSocket connections for order updates
func (h *WebSocketHandler) HandleOrderWebSocket(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := ws.NewClient(h.hub, conn, orderID)
	h.hub.Register() <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()

	log.Printf("✅ WebSocket connection established for order: %s", orderID)
}

// GetHub returns the hub instance (for broadcasting from services)
func (h *WebSocketHandler) GetHub() *ws.Hub {
	return h.hub
}
