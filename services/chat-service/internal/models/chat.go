package models

import "time"

type ChatMessage struct {
	ID         string     `json:"id"`
	OrderID    string     `json:"order_id"`
	SenderID   string     `json:"sender_id"`
	SenderType string     `json:"sender_type"` // "client" or "provider"
	Message    string     `json:"message"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
}

type SendMessageRequest struct {
	OrderID    string `json:"order_id" binding:"required"`
	SenderID   string `json:"sender_id" binding:"required"`
	SenderType string `json:"sender_type" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Message ChatMessage `json:"message,omitempty"`
	Count   int         `json:"count,omitempty"`
}
