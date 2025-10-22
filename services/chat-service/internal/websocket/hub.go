package websocket

import (
	"log"
	"sync"
)

type Hub struct {
	clients    map[string]map[*Client]bool // orderID -> clients
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type Message struct {
	OrderID string
	Data    []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if h.clients[client.orderID] == nil {
				h.clients[client.orderID] = make(map[*Client]bool)
			}
			h.clients[client.orderID][client] = true
			log.Printf("✅ Client registered for order: %s (Total: %d)", client.orderID, len(h.clients[client.orderID]))
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if clients, ok := h.clients[client.orderID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					log.Printf("❌ Client unregistered from order: %s (Remaining: %d)", client.orderID, len(clients))

					if len(clients) == 0 {
						delete(h.clients, client.orderID)
					}
				}
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			clients := h.clients[message.OrderID]
			h.mutex.RUnlock()

			for client := range clients {
				select {
				case client.send <- message.Data:
				default:
					close(client.send)
					h.mutex.Lock()
					delete(h.clients[message.OrderID], client)
					h.mutex.Unlock()
				}
			}
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) BroadcastToOrder(orderID string, data []byte) {
	h.broadcast <- &Message{
		OrderID: orderID,
		Data:    data,
	}
}
