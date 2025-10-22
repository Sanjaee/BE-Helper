package repository

import (
	"chat-service/internal/models"
	"database/sql"
	"log"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveMessage(msg *models.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (order_id, sender_id, sender_type, message, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		msg.OrderID,
		msg.SenderID,
		msg.SenderType,
		msg.Message,
		msg.IsRead,
		msg.CreatedAt,
	).Scan(&msg.ID)

	if err != nil {
		log.Printf("Error saving message: %v", err)
		return err
	}

	return nil
}

func (r *ChatRepository) GetChatHistory(orderID string) ([]models.ChatMessage, error) {
	query := `
		SELECT id, order_id, sender_id, sender_type, message, is_read, created_at, read_at
		FROM chat_messages
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(
			&msg.ID,
			&msg.OrderID,
			&msg.SenderID,
			&msg.SenderType,
			&msg.Message,
			&msg.IsRead,
			&msg.CreatedAt,
			&msg.ReadAt,
		)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *ChatRepository) GetUnreadCount(orderID, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM chat_messages
		WHERE order_id = $1 
		AND sender_id != $2 
		AND is_read = FALSE
	`

	var count int
	err := r.db.QueryRow(query, orderID, userID).Scan(&count)
	if err != nil {
		log.Printf("Error getting unread count: %v", err)
		return 0, err
	}

	return count, nil
}

func (r *ChatRepository) MarkAsRead(orderID, userID string) error {
	query := `
		UPDATE chat_messages
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP
		WHERE order_id = $1 
		AND sender_id != $2 
		AND is_read = FALSE
	`

	_, err := r.db.Exec(query, orderID, userID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		return err
	}

	return nil
}
