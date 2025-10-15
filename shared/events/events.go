package events

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

// EventService handles RabbitMQ event publishing
type EventService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Event represents a generic event structure
type Event struct {
	Type      string      `json:"type"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// UserRegisteredEvent represents user registration event
type UserRegisteredEvent struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	OTP      string `json:"otp,omitempty"`
}

// UserLoginEvent represents user login event
type UserLoginEvent struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// PasswordResetEvent represents password reset event
type PasswordResetEvent struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	OTP      string `json:"otp,omitempty"`
}

// OrderCreatedEvent represents order creation event
type OrderCreatedEvent struct {
	OrderID  string  `json:"order_id"`
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// PaymentSuccessEvent represents payment success event
type PaymentSuccessEvent struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

// PaymentFailedEvent represents payment failure event
type PaymentFailedEvent struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Reason    string  `json:"reason"`
	Amount    float64 `json:"amount,omitempty"`
	Currency  string  `json:"currency,omitempty"`
}

// NewEventService creates a new event service
func NewEventService() (*EventService, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found in events package, using system env")
	}

	// Get RabbitMQ configuration from environment
	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("RABBITMQ_PORT")
	if port == "" {
		port = "5672"
	}

	username := os.Getenv("RABBITMQ_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("RABBITMQ_PASSWORD")
	if password == "" {
		password = "secret123"
	}

	// Create connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchanges for different services
	exchanges := []string{"user.events", "order.events", "payment.events", "notification.events"}
	for _, exchange := range exchanges {
		if err := ch.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	return &EventService{
		conn:    conn,
		channel: ch,
	}, nil
}

// PublishUserRegistered publishes user registration event
func (es *EventService) PublishUserRegistered(userID, username, email, otp string) error {
	event := Event{
		Type:   "user.registered",
		UserID: userID,
		Data: UserRegisteredEvent{
			UserID:   userID,
			Username: username,
			Email:    email,
			OTP:      otp,
		},
		Timestamp: time.Now().Unix(),
	}

	return es.publishEvent("user.events", "user.registered", event)
}

// PublishUserLogin publishes user login event
func (es *EventService) PublishUserLogin(userID, username, email string) error {
	event := Event{
		Type:   "user.login",
		UserID: userID,
		Data: UserLoginEvent{
			UserID:   userID,
			Username: username,
			Email:    email,
		},
		Timestamp: time.Now().Unix(),
	}

	return es.publishEvent("user.events", "user.login", event)
}

// PublishPasswordReset publishes password reset event
func (es *EventService) PublishPasswordReset(userID, username, email, otp string) error {
	event := Event{
		Type:   "password.reset",
		UserID: userID,
		Data: PasswordResetEvent{
			UserID:   userID,
			Username: username,
			Email:    email,
			OTP:      otp,
		},
		Timestamp: time.Now().Unix(),
	}

	return es.publishEvent("user.events", "password.reset", event)
}

// PublishOrderCreated publishes order creation event
func (es *EventService) PublishOrderCreated(orderID, userID, username, email string, amount float64, currency string) error {
	event := Event{
		Type:   "order.created",
		UserID: userID,
		Data: OrderCreatedEvent{
			OrderID:  orderID,
			UserID:   userID,
			Username: username,
			Email:    email,
			Amount:   amount,
			Currency: currency,
		},
		Timestamp: time.Now().Unix(),
	}

	return es.publishEvent("order.events", "order.created", event)
}

// PublishPaymentSuccess publishes payment success event
func (es *EventService) PublishPaymentSuccess(paymentID, orderID, userID, username, email string, amount float64, currency string) error {
	event := Event{
		Type:   "payment.success",
		UserID: userID,
		Data: PaymentSuccessEvent{
			PaymentID: paymentID,
			OrderID:   orderID,
			UserID:    userID,
			Username:  username,
			Email:     email,
			Amount:    amount,
			Currency:  currency,
		},
		Timestamp: time.Now().Unix(),
	}

	return es.publishEvent("payment.events", "payment.success", event)
}

// PublishPaymentFailed publishes payment failure event
func (es *EventService) PublishPaymentFailed(paymentID, orderID, userID, username, email, reason string, amount float64, currency string) error {
	event := Event{
		Type:   "payment.failed",
		UserID: userID,
		Data: PaymentFailedEvent{
			PaymentID: paymentID,
			OrderID:   orderID,
			UserID:    userID,
			Username:  username,
			Email:     email,
			Reason:    reason,
			Amount:    amount,
			Currency:  currency,
		},
		Timestamp: time.Now().Unix(),
	}

	return es.publishEvent("payment.events", "payment.failed", event)
}

// publishEvent publishes a generic event
func (es *EventService) publishEvent(exchange, routingKey string, event Event) error {
	// Marshal event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish message
	err = es.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("✅ Event published: %s.%s", exchange, routingKey)
	return nil
}

// Close closes the RabbitMQ connection
func (es *EventService) Close() error {
	if es.channel != nil {
		es.channel.Close()
	}
	if es.conn != nil {
		return es.conn.Close()
	}
	return nil
}

// GetChannel returns the RabbitMQ channel for consumers
func (es *EventService) GetChannel() *amqp.Channel {
	return es.channel
}

// HealthCheck checks if RabbitMQ connection is healthy
func (es *EventService) HealthCheck() error {
	if es.conn == nil || es.channel == nil {
		return fmt.Errorf("RabbitMQ connection not initialized")
	}

	// Try to declare a temporary queue to test connection
	_, err := es.channel.QueueDeclare(
		"health_check", // name
		false,          // durable
		true,           // delete when unused
		true,           // exclusive
		false,          // no-wait
		nil,            // arguments
	)

	if err != nil {
		return fmt.Errorf("RabbitMQ health check failed: %w", err)
	}

	// Clean up the temporary queue
	es.channel.QueueDelete("health_check", false, false, false)

	return nil
}
