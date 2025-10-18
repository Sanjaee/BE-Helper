package repository

import (
	"order-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BroadcastRepository interface {
	Create(broadcast *models.OrderBroadcast) error
	GetByOrderID(orderID uuid.UUID) ([]models.OrderBroadcast, error)
	GetByProviderID(providerID uuid.UUID) ([]models.OrderBroadcast, error)
	Update(broadcast *models.OrderBroadcast) error
	MarkAsAccepted(orderID, providerID uuid.UUID) error
	GetAcceptedBroadcast(orderID uuid.UUID) (*models.OrderBroadcast, error)
}

type broadcastRepository struct {
	db *gorm.DB
}

func NewBroadcastRepository(db *gorm.DB) BroadcastRepository {
	return &broadcastRepository{db: db}
}

func (r *broadcastRepository) Create(broadcast *models.OrderBroadcast) error {
	return r.db.Create(broadcast).Error
}

func (r *broadcastRepository) GetByOrderID(orderID uuid.UUID) ([]models.OrderBroadcast, error) {
	var broadcasts []models.OrderBroadcast
	err := r.db.Where("order_id = ?", orderID).Find(&broadcasts).Error
	return broadcasts, err
}

func (r *broadcastRepository) GetByProviderID(providerID uuid.UUID) ([]models.OrderBroadcast, error) {
	var broadcasts []models.OrderBroadcast
	err := r.db.Where("provider_id = ?", providerID).Find(&broadcasts).Error
	return broadcasts, err
}

func (r *broadcastRepository) Update(broadcast *models.OrderBroadcast) error {
	return r.db.Save(broadcast).Error
}

func (r *broadcastRepository) MarkAsAccepted(orderID, providerID uuid.UUID) error {
	return r.db.Model(&models.OrderBroadcast{}).
		Where("order_id = ? AND provider_id = ?", orderID, providerID).
		Update("is_accepted", true).Error
}

func (r *broadcastRepository) GetAcceptedBroadcast(orderID uuid.UUID) (*models.OrderBroadcast, error) {
	var broadcast models.OrderBroadcast
	err := r.db.Where("order_id = ? AND is_accepted = ?", orderID, true).First(&broadcast).Error
	if err != nil {
		return nil, err
	}
	return &broadcast, nil
}
