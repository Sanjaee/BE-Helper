package database

import (
	"chat-service/internal/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS chat_messages (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		order_id UUID NOT NULL,
		sender_id UUID NOT NULL,
		sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('client', 'provider')),
		message TEXT NOT NULL,
		is_read BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		read_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_chat_order_id ON chat_messages(order_id);
	CREATE INDEX IF NOT EXISTS idx_chat_created_at ON chat_messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_chat_unread ON chat_messages(order_id, is_read) WHERE is_read = FALSE;
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating tables: %v", err)
		return err
	}

	log.Println("✅ Chat tables created successfully")
	return nil
}
