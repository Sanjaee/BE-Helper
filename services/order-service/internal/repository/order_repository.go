package repository

import (
	"order-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *models.Order) error
	GetByID(id uuid.UUID) (*models.Order, error)
	GetByOrderNumber(orderNumber string) (*models.Order, error)
	Update(order *models.Order) error
	GetClientOrders(clientID uuid.UUID) ([]models.Order, error)
	GetProviderOrders(providerID uuid.UUID) ([]models.Order, error)
	GetPendingOrders() ([]models.Order, error)
	GetActiveOrders() ([]models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Broadcasts").First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNumber(orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Broadcasts").First(&order, "order_number = ?", orderNumber).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) GetClientOrders(clientID uuid.UUID) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Where("client_id = ?", clientID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetProviderOrders(providerID uuid.UUID) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Where("service_provider_id = ?", providerID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetPendingOrders() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Where("status = ?", models.OrderStatusPending).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetActiveOrders() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Where("status IN ?", []models.OrderStatus{
		models.OrderStatusAccepted,
		models.OrderStatusOnTheWay,
		models.OrderStatusArrived,
		models.OrderStatusInProgress,
	}).Find(&orders).Error
	return orders, err
}
