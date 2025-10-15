package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"notification-service/internal/config"
	"notification-service/internal/events"
	"notification-service/internal/models"
	"notification-service/internal/services"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// EventConsumer handles events from RabbitMQ and processes notifications
type EventConsumer struct {
	conn                 *amqp.Connection
	channel              *amqp.Channel
	config               *config.Config
	notificationService  *services.NotificationService
	logger               *logrus.Logger
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(cfg *config.Config, notificationService *services.NotificationService, logger *logrus.Logger) (*EventConsumer, error) {
	// Initialize database connection for notification service
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	
	if err := notificationService.InitializeDB(dsn); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Get RabbitMQ configuration
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", 
		cfg.RabbitMQUsername, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchanges for different services
	exchanges := []string{"user.events", "order.events", "payment.events", "notification.events"}
	for _, exchange := range exchanges {
		if err := ch.ExchangeDeclare(
			exchange,
			"topic",
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,   // arguments
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	// Declare queue for notification events
	q, err := ch.QueueDeclare(
		"notification_queue",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchanges for multiple event types
	bindings := []struct {
		exchange  string
		routingKey string
	}{
		{"user.events", "user.registered"},
		{"user.events", "user.login"},
		{"user.events", "password.reset"},
		{"order.events", "order.created"},
		{"order.events", "order.updated"},
		{"order.events", "order.cancelled"},
		{"payment.events", "payment.success"},
		{"payment.events", "payment.failed"},
		{"payment.events", "payment.refunded"},
	}

	for _, binding := range bindings {
		if err := ch.QueueBind(
			q.Name,
			binding.routingKey,
			binding.exchange,
			false,
			nil,
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue to %s.%s: %w", binding.exchange, binding.routingKey, err)
		}
	}

	return &EventConsumer{
		conn:                conn,
		channel:             ch,
		config:              cfg,
		notificationService: notificationService,
		logger:              logger,
	}, nil
}

// Start starts consuming events
func (ec *EventConsumer) Start() error {
	ec.logger.Info("🚀 Starting notification event consumer...")

	// Set QoS to process one message at a time
	if err := ec.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := ec.channel.Consume(
		"notification_queue",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Process messages
	go func() {
		for msg := range msgs {
			ec.processMessage(msg)
		}
	}()

	ec.logger.Info("✅ Notification event consumer started successfully")
	return nil
}

// processMessage processes a single message
func (ec *EventConsumer) processMessage(msg amqp.Delivery) {
	ec.logger.Infof("📧 Processing notification event: %s", msg.RoutingKey)

	var event events.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		ec.logger.Errorf("❌ Failed to unmarshal event: %v", err)
		msg.Nack(false, false) // Reject message
		return
	}

	// Convert event data to our event data structure
	eventData := ec.convertToEventData(&event)

	// Process the event
	ctx := context.Background()
	if err := ec.notificationService.ProcessEvent(ctx, event.Type, eventData); err != nil {
		ec.logger.Errorf("❌ Failed to process event %s: %v", event.Type, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	// Acknowledge successful processing
	msg.Ack(false)
	ec.logger.Infof("✅ Successfully processed notification event: %s", event.Type)
}

// convertToEventData converts the generic event data to our EventData structure
func (ec *EventConsumer) convertToEventData(event *events.Event) *models.EventData {
	eventData := &models.EventData{
		UserID: event.UserID,
	}

	// Extract data based on event type
	switch event.Type {
	case "user.registered":
		if data, ok := event.Data.(map[string]interface{}); ok {
			eventData.Username = getStringFromMap(data, "username")
			eventData.Email = getStringFromMap(data, "email")
			eventData.OTP = getStringFromMap(data, "otp")
		}
	case "user.login":
		if data, ok := event.Data.(map[string]interface{}); ok {
			eventData.Username = getStringFromMap(data, "username")
			eventData.Email = getStringFromMap(data, "email")
		}
	case "password.reset":
		if data, ok := event.Data.(map[string]interface{}); ok {
			eventData.Username = getStringFromMap(data, "username")
			eventData.Email = getStringFromMap(data, "email")
			eventData.OTP = getStringFromMap(data, "otp")
		}
	case "order.created":
		if data, ok := event.Data.(map[string]interface{}); ok {
			eventData.Username = getStringFromMap(data, "username")
			eventData.Email = getStringFromMap(data, "email")
			eventData.OrderID = getStringFromMap(data, "order_id")
			eventData.Amount = getFloat64FromMap(data, "amount")
			eventData.Currency = getStringFromMap(data, "currency")
		}
	case "payment.success":
		if data, ok := event.Data.(map[string]interface{}); ok {
			eventData.Username = getStringFromMap(data, "username")
			eventData.Email = getStringFromMap(data, "email")
			eventData.OrderID = getStringFromMap(data, "order_id")
			eventData.PaymentID = getStringFromMap(data, "payment_id")
			eventData.Amount = getFloat64FromMap(data, "amount")
			eventData.Currency = getStringFromMap(data, "currency")
		}
	case "payment.failed":
		if data, ok := event.Data.(map[string]interface{}); ok {
			eventData.Username = getStringFromMap(data, "username")
			eventData.Email = getStringFromMap(data, "email")
			eventData.OrderID = getStringFromMap(data, "order_id")
			eventData.PaymentID = getStringFromMap(data, "payment_id")
			eventData.Message = getStringFromMap(data, "reason")
		}
	}

	return eventData
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getFloat64FromMap safely extracts a float64 value from a map
func getFloat64FromMap(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// Stop stops the event consumer
func (ec *EventConsumer) Stop() error {
	ec.logger.Info("🛑 Stopping notification event consumer...")

	if ec.channel != nil {
		ec.channel.Close()
	}
	if ec.conn != nil {
		return ec.conn.Close()
	}

	ec.logger.Info("✅ Notification event consumer stopped")
	return nil
}

// HealthCheck checks if the event consumer is healthy
func (ec *EventConsumer) HealthCheck() error {
	if ec.conn == nil || ec.channel == nil {
		return fmt.Errorf("event consumer not initialized")
	}

	// Check notification service health
	if err := ec.notificationService.HealthCheck(); err != nil {
		return fmt.Errorf("notification service health check failed: %w", err)
	}

	return nil
}
