package services

import (
	"chat-service/internal/events"
	"chat-service/internal/models"
	"chat-service/internal/repository"
	"chat-service/internal/websocket"
	"encoding/json"
	"time"
)

type ChatService struct {
	repo     *repository.ChatRepository
	rabbitMQ *events.RabbitMQ
	hub      *websocket.Hub
}

func NewChatService(repo *repository.ChatRepository, rabbitMQ *events.RabbitMQ, hub *websocket.Hub) *ChatService {
	return &ChatService{
		repo:     repo,
		rabbitMQ: rabbitMQ,
		hub:      hub,
	}
}

func (s *ChatService) SendMessage(req *models.SendMessageRequest) (*models.ChatMessage, error) {
	// Create message
	msg := &models.ChatMessage{
		OrderID:    req.OrderID,
		SenderID:   req.SenderID,
		SenderType: req.SenderType,
		Message:    req.Message,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	// Save to database
	err := s.repo.SaveMessage(msg)
	if err != nil {
		return nil, err
	}

	// Broadcast via WebSocket
	wsMsg := models.WebSocketMessage{
		Type:    "new_message",
		Message: *msg,
	}
	msgBytes, _ := json.Marshal(wsMsg)
	s.hub.BroadcastToOrder(req.OrderID, msgBytes)

	// Publish to RabbitMQ for notification service
	s.rabbitMQ.PublishChatMessage(req.OrderID, msgBytes)

	return msg, nil
}

func (s *ChatService) GetChatHistory(orderID string) ([]models.ChatMessage, error) {
	return s.repo.GetChatHistory(orderID)
}

func (s *ChatService) GetUnreadCount(orderID, userID string) (int, error) {
	return s.repo.GetUnreadCount(orderID, userID)
}

func (s *ChatService) MarkAsRead(orderID, userID string) error {
	return s.repo.MarkAsRead(orderID, userID)
}

