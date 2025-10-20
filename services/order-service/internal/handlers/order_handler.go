package handlers

import (
	"net/http"
	"order-service/internal/models"
	"order-service/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set requested time to now if not provided
	if req.RequestedTime.IsZero() {
		req.RequestedTime = time.Now()
	}

	order, err := h.orderService.CreateOrder(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create order",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"data":    order,
	})
}

// GetOrder gets an order by ID
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order,
	})
}

// AcceptOrder accepts an order by a provider
func (h *OrderHandler) AcceptOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	var req models.AcceptOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	order, err := h.orderService.AcceptOrder(orderID, req.ProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to accept order",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order accepted successfully",
		"data":    order,
	})
}

// UpdateToOnTheWay updates order status to on the way
func (h *OrderHandler) UpdateToOnTheWay(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	// Get provider ID from query parameter or header
	providerIDStr := c.Query("provider_id")
	if providerIDStr == "" {
		providerIDStr = c.GetHeader("X-Provider-ID")
	}
	if providerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provider ID is required",
		})
		return
	}

	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider ID",
		})
		return
	}

	order, err := h.orderService.UpdateToOnTheWay(orderID, providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update order status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated to on the way",
		"data":    order,
	})
}

// UpdateToArrived updates order status to arrived
func (h *OrderHandler) UpdateToArrived(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	// Get provider ID from query parameter or header
	providerIDStr := c.Query("provider_id")
	if providerIDStr == "" {
		providerIDStr = c.GetHeader("X-Provider-ID")
	}
	if providerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provider ID is required",
		})
		return
	}

	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider ID",
		})
		return
	}

	order, err := h.orderService.UpdateToArrived(orderID, providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update order status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated to arrived",
		"data":    order,
	})
}

// GetClientOrders gets orders for a specific client
func (h *OrderHandler) GetClientOrders(c *gin.Context) {
	clientIDStr := c.Param("client_id")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid client ID",
		})
		return
	}

	orders, err := h.orderService.GetClientOrders(clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get client orders",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
	})
}

// GetProviderOrders gets orders for a specific provider
func (h *OrderHandler) GetProviderOrders(c *gin.Context) {
	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider ID",
		})
		return
	}

	orders, err := h.orderService.GetProviderOrders(providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get provider orders",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
	})
}

// GetPendingOrders gets all pending orders
func (h *OrderHandler) GetPendingOrders(c *gin.Context) {
	orders, err := h.orderService.GetPendingOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get pending orders",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
	})
}

// CancelOrder cancels an order
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	var req models.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	order, err := h.orderService.CancelOrder(orderID, req.CancelledBy, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to cancel order",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"data":    order,
	})
}

// StartJob starts the job after provider has arrived
func (h *OrderHandler) StartJob(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	// Get provider ID from query parameter or header
	providerIDStr := c.Query("provider_id")
	if providerIDStr == "" {
		providerIDStr = c.GetHeader("X-Provider-ID")
	}
	if providerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provider ID is required",
		})
		return
	}

	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider ID",
		})
		return
	}

	order, err := h.orderService.StartJob(orderID, providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to start job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job started successfully",
		"data":    order,
	})
}

// CompleteJob marks the job as completed by provider
func (h *OrderHandler) CompleteJob(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	// Get provider ID from query parameter or header
	providerIDStr := c.Query("provider_id")
	if providerIDStr == "" {
		providerIDStr = c.GetHeader("X-Provider-ID")
	}
	if providerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provider ID is required",
		})
		return
	}

	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid provider ID",
		})
		return
	}

	order, err := h.orderService.CompleteJob(orderID, providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to complete job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job completed successfully, waiting for client approval",
		"data":    order,
	})
}
