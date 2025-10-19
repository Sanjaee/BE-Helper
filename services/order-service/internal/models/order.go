package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusAccepted   OrderStatus = "ACCEPTED"
	OrderStatusOnTheWay   OrderStatus = "ON_THE_WAY"
	OrderStatusArrived    OrderStatus = "ARRIVED"
	OrderStatusInProgress OrderStatus = "IN_PROGRESS"
	OrderStatusCompleted  OrderStatus = "COMPLETED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type Order struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderNumber        string         `json:"order_number" gorm:"uniqueIndex;not null"`
	ClientID           uuid.UUID      `json:"client_id" gorm:"not null"`
	ServiceProviderID  *uuid.UUID     `json:"service_provider_id" gorm:"null"`
	Status             OrderStatus    `json:"status" gorm:"not null;default:'PENDING'"`
	Description        string         `json:"description" gorm:"type:text"`
	ServiceLatitude    float64        `json:"service_latitude" gorm:"not null"`
	ServiceLongitude   float64        `json:"service_longitude" gorm:"not null"`
	ServiceAddress     string         `json:"service_address" gorm:"not null"`
	RequestedTime      time.Time      `json:"requested_time" gorm:"not null"`
	BroadcastTime      *time.Time     `json:"broadcast_time" gorm:"null"`
	AcceptedTime       *time.Time     `json:"accepted_time" gorm:"null"`
	ArrivedTime        *time.Time     `json:"arrived_time" gorm:"null"`
	StartedTime        *time.Time     `json:"started_time" gorm:"null"`
	CompletedTime      *time.Time     `json:"completed_time" gorm:"null"`
	CancelledTime      *time.Time     `json:"cancelled_time" gorm:"null"`
	DurationMinutes    int            `json:"duration_minutes" gorm:"default:0"`
	BaseAmount         float64        `json:"base_amount" gorm:"default:0"`
	ServiceFee         float64        `json:"service_fee" gorm:"default:0"`
	TotalAmount        float64        `json:"total_amount" gorm:"default:0"`
	CancellationReason string         `json:"cancellation_reason" gorm:"type:text"`
	CancelledBy        *uuid.UUID     `json:"cancelled_by" gorm:"null"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	Broadcasts []OrderBroadcast `json:"broadcasts" gorm:"foreignKey:OrderID"`
}

type OrderBroadcast struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID    uuid.UUID  `json:"order_id" gorm:"not null"`
	ProviderID uuid.UUID  `json:"provider_id" gorm:"not null"`
	NotifiedAt time.Time  `json:"notified_at" gorm:"default:now()"`
	SeenAt     *time.Time `json:"seen_at" gorm:"null"`
	IsAccepted bool       `json:"is_accepted" gorm:"default:false"`

	// Relations
	Order Order `json:"order" gorm:"foreignKey:OrderID"`
}

type OrderTracking struct {
	ID                      uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID                 uuid.UUID `json:"order_id" gorm:"not null;uniqueIndex"`
	ServiceProviderID       uuid.UUID `json:"service_provider_id" gorm:"not null"`
	CurrentLatitude         float64   `json:"current_latitude" gorm:"not null"`
	CurrentLongitude        float64   `json:"current_longitude" gorm:"not null"`
	DistanceKm              float64   `json:"distance_km" gorm:"default:0"`
	EstimatedArrivalMinutes int       `json:"estimated_arrival_minutes" gorm:"default:0"`
	TrackingStatus          string    `json:"tracking_status" gorm:"default:'ACTIVE'"`
	LastUpdated             time.Time `json:"last_updated" gorm:"default:now()"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type LocationHistory struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID           uuid.UUID `json:"order_id" gorm:"not null"`
	ServiceProviderID uuid.UUID `json:"service_provider_id" gorm:"not null"`
	Latitude          float64   `json:"latitude" gorm:"not null"`
	Longitude         float64   `json:"longitude" gorm:"not null"`
	SpeedKmh          float64   `json:"speed_kmh" gorm:"default:0"`
	AccuracyMeters    int       `json:"accuracy_meters" gorm:"default:0"`
	HeadingDegrees    int       `json:"heading_degrees" gorm:"default:0"`
	RecordedAt        time.Time `json:"recorded_at" gorm:"default:now()"`
}

// Request/Response DTOs
type CreateOrderRequest struct {
	ClientID         uuid.UUID `json:"client_id" binding:"required"`
	Description      string    `json:"description" binding:"required"`
	ServiceLatitude  float64   `json:"service_latitude" binding:"required"`
	ServiceLongitude float64   `json:"service_longitude" binding:"required"`
	ServiceAddress   string    `json:"service_address" binding:"required"`
	RequestedTime    time.Time `json:"requested_time" binding:"required"`
}

// UnmarshalJSON custom unmarshal for CreateOrderRequest to handle time parsing
func (c *CreateOrderRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateOrderRequest
	aux := &struct {
		RequestedTime string `json:"requested_time"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}

	var err error
	for _, format := range formats {
		c.RequestedTime, err = time.Parse(format, aux.RequestedTime)
		if err == nil {
			return nil
		}
	}

	return err
}

type AcceptOrderRequest struct {
	ProviderID uuid.UUID `json:"provider_id" binding:"required"`
}

type CancelOrderRequest struct {
	CancelledBy uuid.UUID `json:"cancelled_by" binding:"required"`
	Reason      string    `json:"reason" binding:"required"`
}

type UpdateLocationRequest struct {
	OrderID           uuid.UUID `json:"order_id" binding:"required"`
	ServiceProviderID uuid.UUID `json:"service_provider_id" binding:"required"`
	Latitude          float64   `json:"latitude" binding:"required"`
	Longitude         float64   `json:"longitude" binding:"required"`
	SpeedKmh          float64   `json:"speed_kmh"`
	AccuracyMeters    int       `json:"accuracy_meters"`
	HeadingDegrees    int       `json:"heading_degrees"`
}

type OrderResponse struct {
	ID                 uuid.UUID   `json:"id"`
	OrderNumber        string      `json:"order_number"`
	ClientID           uuid.UUID   `json:"client_id"`
	ServiceProviderID  *uuid.UUID  `json:"service_provider_id"`
	Status             OrderStatus `json:"status"`
	Description        string      `json:"description"`
	ServiceLatitude    float64     `json:"service_latitude"`
	ServiceLongitude   float64     `json:"service_longitude"`
	ServiceAddress     string      `json:"service_address"`
	RequestedTime      time.Time   `json:"requested_time"`
	BroadcastTime      *time.Time  `json:"broadcast_time"`
	AcceptedTime       *time.Time  `json:"accepted_time"`
	ArrivedTime        *time.Time  `json:"arrived_time"`
	StartedTime        *time.Time  `json:"started_time"`
	CompletedTime      *time.Time  `json:"completed_time"`
	CancelledTime      *time.Time  `json:"cancelled_time"`
	DurationMinutes    int         `json:"duration_minutes"`
	BaseAmount         float64     `json:"base_amount"`
	ServiceFee         float64     `json:"service_fee"`
	TotalAmount        float64     `json:"total_amount"`
	CancellationReason string      `json:"cancellation_reason"`
	CancelledBy        *uuid.UUID  `json:"cancelled_by"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}
