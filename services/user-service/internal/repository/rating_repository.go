package repository

import (
	"user-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RatingRepository handles rating data operations
type RatingRepository struct {
	db *gorm.DB
}

// NewRatingRepository creates a new rating repository
func NewRatingRepository(db *gorm.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

// Create creates a new rating
func (r *RatingRepository) Create(rating *models.Rating) error {
	return r.db.Create(rating).Error
}

// GetByOrderID gets a rating by order ID
func (r *RatingRepository) GetByOrderID(orderID uuid.UUID) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.Where("order_id = ?", orderID).First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

// GetByProviderID gets all ratings for a provider
func (r *RatingRepository) GetByProviderID(providerID uuid.UUID, limit, offset int) ([]models.Rating, error) {
	var ratings []models.Rating
	query := r.db.Where("service_provider_id = ?", providerID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&ratings).Error
	return ratings, err
}

// GetProviderStats gets rating statistics for a provider
func (r *RatingRepository) GetProviderStats(providerID uuid.UUID) (*models.ProviderRatingStats, error) {
	var stats models.ProviderRatingStats
	stats.ServiceProviderID = providerID

	// Get total ratings
	r.db.Model(&models.Rating{}).
		Where("service_provider_id = ?", providerID).
		Count(&stats.TotalRatings)

	if stats.TotalRatings == 0 {
		return &stats, nil
	}

	// Get average rating
	var avgRating *float64
	r.db.Model(&models.Rating{}).
		Select("AVG(rating)").
		Where("service_provider_id = ?", providerID).
		Scan(&avgRating)

	if avgRating != nil {
		stats.AverageRating = *avgRating
	}

	// Get rating distribution
	r.db.Model(&models.Rating{}).
		Where("service_provider_id = ? AND rating = ?", providerID, 5).
		Count(&stats.FiveStars)

	r.db.Model(&models.Rating{}).
		Where("service_provider_id = ? AND rating = ?", providerID, 4).
		Count(&stats.FourStars)

	r.db.Model(&models.Rating{}).
		Where("service_provider_id = ? AND rating = ?", providerID, 3).
		Count(&stats.ThreeStars)

	r.db.Model(&models.Rating{}).
		Where("service_provider_id = ? AND rating = ?", providerID, 2).
		Count(&stats.TwoStars)

	r.db.Model(&models.Rating{}).
		Where("service_provider_id = ? AND rating = ?", providerID, 1).
		Count(&stats.OneStar)

	return &stats, nil
}

// CheckIfRated checks if an order has been rated
func (r *RatingRepository) CheckIfRated(orderID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Rating{}).
		Where("order_id = ?", orderID).
		Count(&count).Error

	return count > 0, err
}

// GetByClientID gets all ratings given by a client
func (r *RatingRepository) GetByClientID(clientID uuid.UUID, limit, offset int) ([]models.Rating, error) {
	var ratings []models.Rating
	query := r.db.Where("client_id = ?", clientID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&ratings).Error
	return ratings, err
}
