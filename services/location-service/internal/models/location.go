package models

import (
	"time"

	"github.com/google/uuid"
)

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
type UpdateLocationRequest struct {
	OrderID           uuid.UUID `json:"order_id" binding:"required"`
	ServiceProviderID uuid.UUID `json:"service_provider_id" binding:"required"`
	Latitude          float64   `json:"latitude" binding:"required"`
	Longitude         float64   `json:"longitude" binding:"required"`
	SpeedKmh          float64   `json:"speed_kmh"`
	AccuracyMeters    int       `json:"accuracy_meters"`
	HeadingDegrees    int       `json:"heading_degrees"`
}

type LocationResponse struct {
	OrderID                 uuid.UUID `json:"order_id"`
	ServiceProviderID       uuid.UUID `json:"service_provider_id"`
	Latitude                float64   `json:"latitude"`
	Longitude               float64   `json:"longitude"`
	DistanceKm              float64   `json:"distance_km"`
	EstimatedArrivalMinutes int       `json:"estimated_arrival_minutes"`
	TrackingStatus          string    `json:"tracking_status"`
	LastUpdated             time.Time `json:"last_updated"`
}

type LocationHistoryResponse struct {
	ID                uuid.UUID `json:"id"`
	OrderID           uuid.UUID `json:"order_id"`
	ServiceProviderID uuid.UUID `json:"service_provider_id"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	SpeedKmh          float64   `json:"speed_kmh"`
	AccuracyMeters    int       `json:"accuracy_meters"`
	HeadingDegrees    int       `json:"heading_degrees"`
	RecordedAt        time.Time `json:"recorded_at"`
}
