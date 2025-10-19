package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by order ID
	clients map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to clients
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	OrderID uuid.UUID
	Type    string // "order_update", "order_cancelled", etc.
	Data    interface{}
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.orderID] == nil {
				h.clients[client.orderID] = make(map[*Client]bool)
			}
			h.clients[client.orderID][client] = true
			h.mu.Unlock()
			log.Printf("✅ WebSocket client registered for order: %s (total: %d)", client.orderID, len(h.clients[client.orderID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.orderID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.orderID)
					}
					log.Printf("❌ WebSocket client unregistered for order: %s", client.orderID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			orderIDStr := message.OrderID.String()
			clients := h.clients[orderIDStr]
			h.mu.RUnlock()

			if len(clients) > 0 {
				// Prepare JSON message
				jsonData, err := json.Marshal(map[string]interface{}{
					"type": message.Type,
					"data": message.Data,
				})
				if err != nil {
					log.Printf("Error marshaling broadcast message: %v", err)
					continue
				}

				// Send to all clients for this order
				for client := range clients {
					select {
					case client.send <- jsonData:
						log.Printf("📤 Broadcast sent to client for order: %s, type: %s", orderIDStr, message.Type)
					default:
						// Client's send channel is full, unregister it
						h.mu.Lock()
						close(client.send)
						delete(h.clients[orderIDStr], client)
						if len(h.clients[orderIDStr]) == 0 {
							delete(h.clients, orderIDStr)
						}
						h.mu.Unlock()
					}
				}
			}
		}
	}
}

// Broadcast sends a message to all clients subscribed to an order
func (h *Hub) Broadcast(orderID uuid.UUID, msgType string, data interface{}) {
	h.broadcast <- &BroadcastMessage{
		OrderID: orderID,
		Type:    msgType,
		Data:    data,
	}
}

// GetClientCount returns the number of clients for an order
func (h *Hub) GetClientCount(orderID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[orderID])
}

// Register returns the register channel
func (h *Hub) Register() chan *Client {
	return h.register
}

// Unregister returns the unregister channel
func (h *Hub) Unregister() chan *Client {
	return h.unregister
}
