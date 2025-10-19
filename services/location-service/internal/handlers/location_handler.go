package handlers

import (
	"location-service/internal/models"
	"location-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LocationHandler struct {
	locationService services.LocationService
}

func NewLocationHandler(locationService services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

// UpdateLocation updates the location of a service provider
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	location, err := h.locationService.UpdateLocation(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update location",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Location updated successfully",
		"data":    location,
	})
}

// GetOrderLocation gets the current location for an order
func (h *LocationHandler) GetOrderLocation(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	location, err := h.locationService.GetOrderLocation(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Location not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": location,
	})
}

// GetLocationHistory gets the location history for an order
func (h *LocationHandler) GetLocationHistory(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	history, err := h.locationService.GetLocationHistory(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get location history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
	})
}

// GetProviderLocation gets the current location of the provider for an order
func (h *LocationHandler) GetProviderLocation(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	location, err := h.locationService.GetOrderLocation(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Provider location not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": location,
	})
}