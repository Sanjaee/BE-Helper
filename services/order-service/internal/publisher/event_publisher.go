package publisher

import (
	"order-service/internal/events"
	"order-service/internal/models"
)

type EventPublisher struct {
	rabbitMQ *events.RabbitMQ
}

func NewEventPublisher(rabbitMQ *events.RabbitMQ) *EventPublisher {
	return &EventPublisher{
		rabbitMQ: rabbitMQ,
	}
}

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

// OrderCancelledEvent represents the event when an order is cancelled
type OrderCancelledEvent struct {
	OrderID            string `json:"order_id"`
	ClientID           string `json:"client_id"`
	ServiceProviderID  string `json:"service_provider_id"`
	CancelledBy        string `json:"cancelled_by"`
	CancellationReason string `json:"cancellation_reason"`
	CancelledTime      string `json:"cancelled_time"`
}

// PublishOrderCreated publishes order.created event
func (p *EventPublisher) PublishOrderCreated(order *models.Order) error {
	event := OrderCreatedEvent{
		OrderID:          order.ID.String(),
		ClientID:         order.ClientID.String(),
		Description:      order.Description,
		ServiceLatitude:  order.ServiceLatitude,
		ServiceLongitude: order.ServiceLongitude,
		ServiceAddress:   order.ServiceAddress,
		RequestedTime:    order.RequestedTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	return p.rabbitMQ.Publish("order.exchange", "order.created", event)
}

// PublishOrderAccepted publishes order.accepted event
func (p *EventPublisher) PublishOrderAccepted(order *models.Order) error {
	event := OrderAcceptedEvent{
		OrderID:           order.ID.String(),
		ClientID:          order.ClientID.String(),
		ServiceProviderID: order.ServiceProviderID.String(),
		AcceptedTime:      order.AcceptedTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	return p.rabbitMQ.Publish("order.exchange", "order.accepted", event)
}

// PublishOrderStatusUpdated publishes order status update event
func (p *EventPublisher) PublishOrderStatusUpdated(order *models.Order) error {
	event := OrderStatusUpdatedEvent{
		OrderID:           order.ID.String(),
		ClientID:          order.ClientID.String(),
		ServiceProviderID: order.ServiceProviderID.String(),
		Status:            string(order.Status),
		UpdatedTime:       order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return p.rabbitMQ.Publish("order.exchange", "order.status.updated", event)
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

// PublishOrderCancelled publishes order cancelled event
func (p *EventPublisher) PublishOrderCancelled(order *models.Order) error {
	event := OrderCancelledEvent{
		OrderID:            order.ID.String(),
		ClientID:           order.ClientID.String(),
		ServiceProviderID:  "",
		CancelledBy:        order.CancelledBy.String(),
		CancellationReason: order.CancellationReason,
		CancelledTime:      order.CancelledTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	if order.ServiceProviderID != nil {
		event.ServiceProviderID = order.ServiceProviderID.String()
	}

	return p.rabbitMQ.Publish("order.exchange", "order.cancelled", event)
}
