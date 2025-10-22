package handlers

import (
	"chat-service/internal/services"
	"chat-service/internal/websocket"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WebSocketHandler struct {
	hub     *websocket.Hub
	service *services.ChatService
}

func NewWebSocketHandler(hub *websocket.Hub, service *services.ChatService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:     hub,
		service: service,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	orderID := c.Param("order_id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := websocket.NewClient(h.hub, conn, orderID, userID)
	h.hub.RegisterClient(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Printf("✅ WebSocket connected: Order=%s, User=%s", orderID, userID)
}
