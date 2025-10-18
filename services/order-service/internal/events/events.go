package events

import (
	"encoding/json"
	"log"
)

// OrderCreatedEvent represents the event when a new order is created
type OrderCreatedEvent struct {
	OrderID          string  `json:"order_id"`
	ClientID         string  `json:"client_id"`
	Description      string  `json:"description"`
	ServiceLatitude  float64 `json:"service_latitude"`
	ServiceLongitude float64 `json:"service_longitude"`
	ServiceAddress   string  `json:"service_address"`
	RequestedTime    string  `json:"requested_time"`
}

// OrderAcceptedEvent represents the event when an order is accepted
type OrderAcceptedEvent struct {
	OrderID           string `json:"order_id"`
	ClientID          string `json:"client_id"`
	ServiceProviderID string `json:"service_provider_id"`
	AcceptedTime      string `json:"accepted_time"`
}

// OrderStatusUpdatedEvent represents the event when order status is updated
type OrderStatusUpdatedEvent struct {
	OrderID           string `json:"order_id"`
	ClientID          string `json:"client_id"`
	ServiceProviderID string `json:"service_provider_id"`
	Status            string `json:"status"`
	UpdatedTime       string `json:"updated_time"`
}

// LocationUpdatedEvent represents the event when provider location is updated
type LocationUpdatedEvent struct {
	OrderID           string  `json:"order_id"`
	ServiceProviderID string  `json:"service_provider_id"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	SpeedKmh          float64 `json:"speed_kmh"`
	AccuracyMeters    int     `json:"accuracy_meters"`
	HeadingDegrees    int     `json:"heading_degrees"`
	UpdatedAt         string  `json:"updated_at"`
}

// NotificationOrderBroadcastEvent represents the event for broadcasting order to providers
type NotificationOrderBroadcastEvent struct {
	OrderID          string   `json:"order_id"`
	ClientID         string   `json:"client_id"`
	Description      string   `json:"description"`
	ServiceLatitude  float64  `json:"service_latitude"`
	ServiceLongitude float64  `json:"service_longitude"`
	ServiceAddress   string   `json:"service_address"`
	RequestedTime    string   `json:"requested_time"`
	ProviderIDs      []string `json:"provider_ids"`
}

// NotificationOrderAcceptedEvent represents the event for notifying client about accepted order
type NotificationOrderAcceptedEvent struct {
	OrderID           string `json:"order_id"`
	ClientID          string `json:"client_id"`
	ServiceProviderID string `json:"service_provider_id"`
	ProviderName      string `json:"provider_name"`
	ProviderPhone     string `json:"provider_phone"`
	AcceptedTime      string `json:"accepted_time"`
}

// StartOrderBroadcastListener listens for order.created events and broadcasts to providers
func StartOrderBroadcastListener(rabbitMQ *RabbitMQ) {
	msgs, err := rabbitMQ.Consume("order.broadcast.queue", "order.exchange", "order.created")
	if err != nil {
		log.Printf("Failed to start order broadcast listener: %v", err)
		return
	}

	log.Println("🎧 Order broadcast listener started")

	for msg := range msgs {
		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal order created event: %v", err)
			continue
		}

		log.Printf("📢 Received order created event: %s", event.OrderID)

		// Get all available providers (this would typically come from user service)
		// For now, we'll simulate getting provider IDs
		providerIDs := []string{"provider-1", "provider-2", "provider-3"} // This should come from user service

		// Create broadcast event
		broadcastEvent := NotificationOrderBroadcastEvent{
			OrderID:          event.OrderID,
			ClientID:         event.ClientID,
			Description:      event.Description,
			ServiceLatitude:  event.ServiceLatitude,
			ServiceLongitude: event.ServiceLongitude,
			ServiceAddress:   event.ServiceAddress,
			RequestedTime:    event.RequestedTime,
			ProviderIDs:      providerIDs,
		}

		// Publish to notification service
		if err := rabbitMQ.Publish("notification.exchange", "notification.order.broadcast", broadcastEvent); err != nil {
			log.Printf("Failed to publish order broadcast event: %v", err)
		} else {
			log.Printf("📤 Published order broadcast event for order: %s", event.OrderID)
		}
	}
}

// StartOrderAcceptedListener listens for order.accepted events
func StartOrderAcceptedListener(rabbitMQ *RabbitMQ) {
	msgs, err := rabbitMQ.Consume("order.accepted.queue", "order.exchange", "order.accepted")
	if err != nil {
		log.Printf("Failed to start order accepted listener: %v", err)
		return
	}

	log.Println("🎧 Order accepted listener started")

	for msg := range msgs {
		var event OrderAcceptedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal order accepted event: %v", err)
			continue
		}

		log.Printf("✅ Received order accepted event: %s", event.OrderID)

		// Create notification event for client
		notificationEvent := NotificationOrderAcceptedEvent{
			OrderID:           event.OrderID,
			ClientID:          event.ClientID,
			ServiceProviderID: event.ServiceProviderID,
			ProviderName:      "Provider Name",  // This should come from user service
			ProviderPhone:     "Provider Phone", // This should come from user service
			AcceptedTime:      event.AcceptedTime,
		}

		// Publish to notification service
		if err := rabbitMQ.Publish("notification.exchange", "notification.order.accepted", notificationEvent); err != nil {
			log.Printf("Failed to publish order accepted notification: %v", err)
		} else {
			log.Printf("📤 Published order accepted notification for order: %s", event.OrderID)
		}
	}
}
