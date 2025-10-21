package models

import (
	"time"

	"github.com/google/uuid"
)

// Rating represents a rating given by a client to a service provider after job completion
type Rating struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	OrderID            uuid.UUID  `gorm:"type:uuid;not null;unique" json:"order_id"`
	ClientID           uuid.UUID  `gorm:"type:uuid;not null" json:"client_id"`
	ServiceProviderID  uuid.UUID  `gorm:"type:uuid;not null" json:"service_provider_id"`
	Rating             int        `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"` // 1-5 stars
	Review             *string    `gorm:"type:text" json:"review,omitempty"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for Rating model
func (Rating) TableName() string {
	return "ratings"
}

// RatingCreateRequest represents the request to create a new rating
type RatingCreateRequest struct {
	OrderID           string  `json:"order_id" validate:"required,uuid"`
	ServiceProviderID string  `json:"service_provider_id" validate:"required,uuid"`
	Rating            int     `json:"rating" validate:"required,min=1,max=5"`
	Review            *string `json:"review" validate:"omitempty,max=1000"`
}

// RatingResponse represents the rating response
type RatingResponse struct {
	ID                uuid.UUID  `json:"id"`
	OrderID           uuid.UUID  `json:"order_id"`
	ClientID          uuid.UUID  `json:"client_id"`
	ServiceProviderID uuid.UUID  `json:"service_provider_id"`
	Rating            int        `json:"rating"`
	Review            *string    `json:"review,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ToResponse converts Rating to RatingResponse
func (r *Rating) ToResponse() RatingResponse {
	return RatingResponse{
		ID:                r.ID,
		OrderID:           r.OrderID,
		ClientID:          r.ClientID,
		ServiceProviderID: r.ServiceProviderID,
		Rating:            r.Rating,
		Review:            r.Review,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// ProviderRatingStats represents provider rating statistics
type ProviderRatingStats struct {
	ServiceProviderID uuid.UUID `json:"service_provider_id"`
	TotalRatings      int64     `json:"total_ratings"`
	AverageRating     float64   `json:"average_rating"`
	FiveStars         int64     `json:"five_stars"`
	FourStars         int64     `json:"four_stars"`
	ThreeStars        int64     `json:"three_stars"`
	TwoStars          int64     `json:"two_stars"`
	OneStar           int64     `json:"one_star"`
}

