package services

import (
	"chat-service/internal/cache"
	"chat-service/internal/events"
	"chat-service/internal/models"
	"chat-service/internal/repository"
	"chat-service/internal/websocket"
	"encoding/json"
	"log"
	"time"
)

type ChatService struct {
	repo     *repository.ChatRepository
	rabbitMQ *events.RabbitMQ
	hub      *websocket.Hub
	cache    *cache.RedisCache
}

func NewChatService(repo *repository.ChatRepository, rabbitMQ *events.RabbitMQ, hub *websocket.Hub, cache *cache.RedisCache) *ChatService {
	return &ChatService{
		repo:     repo,
		rabbitMQ: rabbitMQ,
		hub:      hub,
		cache:    cache,
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

	// Invalidate cache for this order
	if s.cache != nil {
		if err := s.cache.InvalidateOrderCache(req.OrderID); err != nil {
			log.Printf("Failed to invalidate cache: %v", err)
		}
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
	// Try to get from cache first
	if s.cache != nil {
		cachedMessages, err := s.cache.GetCachedMessages(orderID)
		if err == nil && cachedMessages != nil {
			log.Printf("Cache HIT for order %s", orderID)
			return cachedMessages, nil
		}
		log.Printf("Cache MISS for order %s", orderID)
	}

	// Get from database
	messages, err := s.repo.GetChatHistory(orderID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.cache != nil && len(messages) > 0 {
		if err := s.cache.CacheMessages(orderID, messages); err != nil {
			log.Printf("Failed to cache messages: %v", err)
		}
	}

	return messages, nil
}

func (s *ChatService) GetUnreadCount(orderID, userID string) (int, error) {
	// Try to get from cache first
	if s.cache != nil {
		count, err := s.cache.GetCachedUnreadCount(orderID, userID)
		if err == nil && count >= 0 {
			return count, nil
		}
	}

	// Get from database
	count, err := s.repo.GetUnreadCount(orderID, userID)
	if err != nil {
		return 0, err
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.CacheUnreadCount(orderID, userID, count); err != nil {
			log.Printf("Failed to cache unread count: %v", err)
		}
	}

	return count, nil
}

func (s *ChatService) MarkAsRead(orderID, userID string) error {
	err := s.repo.MarkAsRead(orderID, userID)
	if err != nil {
		return err
	}

	// Clear unread count cache
	if s.cache != nil {
		if err := s.cache.ClearUnreadCount(orderID, userID); err != nil {
			log.Printf("Failed to clear unread count cache: %v", err)
		}
	}

	return nil
}
