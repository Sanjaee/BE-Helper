package services

import (
	"fmt"
	"order-service/internal/models"
	"order-service/internal/publisher"
	"order-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(req *models.CreateOrderRequest) (*models.Order, error)
	GetOrder(id uuid.UUID) (*models.Order, error)
	AcceptOrder(orderID, providerID uuid.UUID) (*models.Order, error)
	UpdateToOnTheWay(orderID, providerID uuid.UUID) (*models.Order, error)
	UpdateToArrived(orderID, providerID uuid.UUID) (*models.Order, error)
	CancelOrder(orderID, cancelledBy uuid.UUID, reason string) (*models.Order, error)
	GetClientOrders(clientID uuid.UUID) ([]models.Order, error)
	GetProviderOrders(providerID uuid.UUID) ([]models.Order, error)
	GetPendingOrders() ([]models.Order, error)
	UpdateLocation(req *models.UpdateLocationRequest) error
}

type orderService struct {
	orderRepo      repository.OrderRepository
	broadcastRepo  repository.BroadcastRepository
	eventPublisher *publisher.EventPublisher
}

func NewOrderService(orderRepo repository.OrderRepository, broadcastRepo repository.BroadcastRepository, eventPublisher *publisher.EventPublisher) OrderService {
	return &orderService{
		orderRepo:      orderRepo,
		broadcastRepo:  broadcastRepo,
		eventPublisher: eventPublisher,
	}
}

func (s *orderService) CreateOrder(req *models.CreateOrderRequest) (*models.Order, error) {
	// Generate order number
	orderNumber := fmt.Sprintf("ORD-%d", time.Now().Unix())

	// Create order
	order := &models.Order{
		OrderNumber:      orderNumber,
		ClientID:         req.ClientID,
		Status:           models.OrderStatusPending,
		Description:      req.Description,
		ServiceLatitude:  req.ServiceLatitude,
		ServiceLongitude: req.ServiceLongitude,
		ServiceAddress:   req.ServiceAddress,
		RequestedTime:    req.RequestedTime,
		BroadcastTime:    &time.Time{},
	}

	// Save order
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Set broadcast time
	now := time.Now()
	order.BroadcastTime = &now

	// Publish order created event
	if err := s.eventPublisher.PublishOrderCreated(order); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish order created event: %v\n", err)
	}

	return order, nil
}

func (s *orderService) GetOrder(id uuid.UUID) (*models.Order, error) {
	return s.orderRepo.GetByID(id)
}

func (s *orderService) AcceptOrder(orderID, providerID uuid.UUID) (*models.Order, error) {
	// Get order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Check if order is still pending
	if order.Status != models.OrderStatusPending {
		return nil, fmt.Errorf("order is no longer available for acceptance")
	}

	// Update order
	now := time.Now()
	order.ServiceProviderID = &providerID
	order.Status = models.OrderStatusAccepted
	order.AcceptedTime = &now

	if err := s.orderRepo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Mark broadcast as accepted
	if err := s.broadcastRepo.MarkAsAccepted(orderID, providerID); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to mark broadcast as accepted: %v\n", err)
	}

	// Publish order accepted event
	if err := s.eventPublisher.PublishOrderAccepted(order); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish order accepted event: %v\n", err)
	}

	return order, nil
}

func (s *orderService) UpdateToOnTheWay(orderID, providerID uuid.UUID) (*models.Order, error) {
	// Get order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Check if order is accepted and provider matches
	if order.Status != models.OrderStatusAccepted {
		return nil, fmt.Errorf("order is not in accepted status")
	}

	if order.ServiceProviderID == nil || *order.ServiceProviderID != providerID {
		return nil, fmt.Errorf("unauthorized to update this order")
	}

	// Update order
	order.Status = models.OrderStatusOnTheWay

	if err := s.orderRepo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Publish status update event
	if err := s.eventPublisher.PublishOrderStatusUpdated(order); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish order status updated event: %v\n", err)
	}

	return order, nil
}

func (s *orderService) UpdateToArrived(orderID, providerID uuid.UUID) (*models.Order, error) {
	// Get order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Check if order is on the way and provider matches
	if order.Status != models.OrderStatusOnTheWay {
		return nil, fmt.Errorf("order is not on the way")
	}

	if order.ServiceProviderID == nil || *order.ServiceProviderID != providerID {
		return nil, fmt.Errorf("unauthorized to update this order")
	}

	// Update order
	now := time.Now()
	order.Status = models.OrderStatusArrived
	order.ArrivedTime = &now

	if err := s.orderRepo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Publish status update event
	if err := s.eventPublisher.PublishOrderStatusUpdated(order); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish order status updated event: %v\n", err)
	}

	return order, nil
}

func (s *orderService) CancelOrder(orderID, cancelledBy uuid.UUID, reason string) (*models.Order, error) {
	// Get order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Check if order can be cancelled
	if order.Status == models.OrderStatusCompleted || order.Status == models.OrderStatusCancelled {
		return nil, fmt.Errorf("order cannot be cancelled")
	}

	// Check authorization - only client or assigned provider can cancel
	if order.ClientID != cancelledBy && (order.ServiceProviderID == nil || *order.ServiceProviderID != cancelledBy) {
		return nil, fmt.Errorf("unauthorized to cancel this order")
	}

	// Update order
	now := time.Now()
	order.Status = models.OrderStatusCancelled
	order.CancelledTime = &now
	order.CancelledBy = &cancelledBy
	order.CancellationReason = reason

	if err := s.orderRepo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Publish order cancelled event
	if err := s.eventPublisher.PublishOrderCancelled(order); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish order cancelled event: %v\n", err)
	}

	return order, nil
}

func (s *orderService) GetClientOrders(clientID uuid.UUID) ([]models.Order, error) {
	return s.orderRepo.GetClientOrders(clientID)
}

func (s *orderService) GetProviderOrders(providerID uuid.UUID) ([]models.Order, error) {
	return s.orderRepo.GetProviderOrders(providerID)
}

func (s *orderService) GetPendingOrders() ([]models.Order, error) {
	return s.orderRepo.GetPendingOrders()
}

func (s *orderService) UpdateLocation(req *models.UpdateLocationRequest) error {
	// This would typically be handled by the location service
	// For now, we'll just log the location update
	fmt.Printf("Location updated for order %s: lat=%.6f, lng=%.6f\n",
		req.OrderID, req.Latitude, req.Longitude)
	return nil
}
