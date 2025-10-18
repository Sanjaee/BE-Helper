package repository

import (
	"location-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationRepository interface {
	CreateTracking(tracking *models.OrderTracking) error
	UpdateTracking(tracking *models.OrderTracking) error
	GetTrackingByOrderID(orderID uuid.UUID) (*models.OrderTracking, error)
	CreateLocationHistory(history *models.LocationHistory) error
	GetLocationHistory(orderID uuid.UUID) ([]models.LocationHistory, error)
}

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) CreateTracking(tracking *models.OrderTracking) error {
	return r.db.Create(tracking).Error
}

func (r *locationRepository) UpdateTracking(tracking *models.OrderTracking) error {
	return r.db.Save(tracking).Error
}

func (r *locationRepository) GetTrackingByOrderID(orderID uuid.UUID) (*models.OrderTracking, error) {
	var tracking models.OrderTracking
	err := r.db.Where("order_id = ?", orderID).First(&tracking).Error
	if err != nil {
		return nil, err
	}
	return &tracking, nil
}

func (r *locationRepository) CreateLocationHistory(history *models.LocationHistory) error {
	return r.db.Create(history).Error
}

func (r *locationRepository) GetLocationHistory(orderID uuid.UUID) ([]models.LocationHistory, error) {
	var history []models.LocationHistory
	err := r.db.Where("order_id = ?", orderID).Order("recorded_at DESC").Find(&history).Error
	return history, err
}
