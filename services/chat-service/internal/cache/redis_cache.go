package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chat-service/internal/models"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(host, port, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Cache keys
func (r *RedisCache) orderMessagesKey(orderID string) string {
	return fmt.Sprintf("chat:order:%s:messages", orderID)
}

func (r *RedisCache) unreadCountKey(orderID, userID string) string {
	return fmt.Sprintf("chat:order:%s:unread:%s", orderID, userID)
}

// CacheMessages - Cache messages for an order (TTL: 1 hour)
func (r *RedisCache) CacheMessages(orderID string, messages []models.ChatMessage) error {
	key := r.orderMessagesKey(orderID)
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, key, data, time.Hour).Err()
}

// GetCachedMessages - Get cached messages for an order
func (r *RedisCache) GetCachedMessages(orderID string) ([]models.ChatMessage, error) {
	key := r.orderMessagesKey(orderID)
	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var messages []models.ChatMessage
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// InvalidateOrderCache - Clear cache when new message is sent
func (r *RedisCache) InvalidateOrderCache(orderID string) error {
	key := r.orderMessagesKey(orderID)
	return r.client.Del(r.ctx, key).Err()
}

// CacheUnreadCount - Cache unread count for a user
func (r *RedisCache) CacheUnreadCount(orderID, userID string, count int) error {
	key := r.unreadCountKey(orderID, userID)
	return r.client.Set(r.ctx, key, count, 10*time.Minute).Err()
}

// GetCachedUnreadCount - Get cached unread count
func (r *RedisCache) GetCachedUnreadCount(orderID, userID string) (int, error) {
	key := r.unreadCountKey(orderID, userID)
	count, err := r.client.Get(r.ctx, key).Int()
	if err == redis.Nil {
		return -1, nil // Cache miss
	}
	return count, err
}

// IncrementUnreadCount - Increment unread count for a user
func (r *RedisCache) IncrementUnreadCount(orderID, userID string) error {
	key := r.unreadCountKey(orderID, userID)
	_, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return r.client.Expire(r.ctx, key, 10*time.Minute).Err()
}

// ClearUnreadCount - Clear unread count when messages are read
func (r *RedisCache) ClearUnreadCount(orderID, userID string) error {
	key := r.unreadCountKey(orderID, userID)
	return r.client.Del(r.ctx, key).Err()
}

// Close - Close Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
