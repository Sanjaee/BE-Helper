package events

import (
	"encoding/json"
	"log"
)

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

// StartLocationTrackingListener listens for location updates
func StartLocationTrackingListener(rabbitMQ *RabbitMQ) {
	msgs, err := rabbitMQ.Consume("location.tracking.queue", "location.exchange", "location.updated")
	if err != nil {
		log.Printf("Failed to start location tracking listener: %v", err)
		return
	}

	log.Println("🎧 Location tracking listener started")

	for msg := range msgs {
		var event LocationUpdatedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal location updated event: %v", err)
			continue
		}

		log.Printf("📍 Received location update for order: %s", event.OrderID)

		// Process location update
		// This would typically involve updating the tracking data
		// and calculating distance/ETA
	}
}
