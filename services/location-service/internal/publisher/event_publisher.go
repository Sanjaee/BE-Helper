package publisher

import (
	"location-service/internal/events"
	"location-service/internal/models"
)

type EventPublisher struct {
	rabbitMQ *events.RabbitMQ
}

func NewEventPublisher(rabbitMQ *events.RabbitMQ) *EventPublisher {
	return &EventPublisher{
		rabbitMQ: rabbitMQ,
	}
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

// PublishLocationUpdated publishes location update event
func (p *EventPublisher) PublishLocationUpdated(tracking *models.OrderTracking) error {
	event := LocationUpdatedEvent{
		OrderID:           tracking.OrderID.String(),
		ServiceProviderID: tracking.ServiceProviderID.String(),
		Latitude:          tracking.CurrentLatitude,
		Longitude:         tracking.CurrentLongitude,
		UpdatedAt:         tracking.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}

	return p.rabbitMQ.Publish("location.exchange", "location.updated", event)
}
