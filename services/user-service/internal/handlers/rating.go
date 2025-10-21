package handlers

import (
	"net/http"
	"strconv"
	"time"

	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RatingHandler handles rating-related HTTP requests
type RatingHandler struct {
	ratingRepo *repository.RatingRepository
	validator  *validator.Validate
}

// NewRatingHandler creates a new rating handler
func NewRatingHandler(db *gorm.DB) *RatingHandler {
	return &RatingHandler{
		ratingRepo: repository.NewRatingRepository(db),
		validator:  validator.New(),
	}
}

// CreateRating handles creating a new rating
func (rh *RatingHandler) CreateRating(c *gin.Context) {
	userIDStr, _, _, _, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.RatingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := rh.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse UUIDs
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	providerID, err := uuid.Parse(req.ServiceProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service provider ID"})
		return
	}

	// Check if already rated
	alreadyRated, err := rh.ratingRepo.CheckIfRated(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check rating status"})
		return
	}
	if alreadyRated {
		c.JSON(http.StatusConflict, gin.H{"error": "Order has already been rated"})
		return
	}

	// Create rating
	rating := models.Rating{
		OrderID:           orderID,
		ClientID:          userID,
		ServiceProviderID: providerID,
		Rating:            req.Rating,
		Review:            req.Review,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := rh.ratingRepo.Create(&rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rating"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rating created successfully",
		"rating":  rating.ToResponse(),
	})
}

// GetRatingByOrder gets rating for a specific order
func (rh *RatingHandler) GetRatingByOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	rating, err := rh.ratingRepo.GetByOrderID(orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rating": rating.ToResponse()})
}

// GetProviderRatings gets all ratings for a provider
func (rh *RatingHandler) GetProviderRatings(c *gin.Context) {
	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	ratings, err := rh.ratingRepo.GetByProviderID(providerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ratings"})
		return
	}

	// Convert to responses
	responses := make([]models.RatingResponse, len(ratings))
	for i, r := range ratings {
		responses[i] = r.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"ratings": responses,
		"count":   len(responses),
	})
}

// GetProviderStats gets rating statistics for a provider
func (rh *RatingHandler) GetProviderStats(c *gin.Context) {
	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	stats, err := rh.ratingRepo.GetProviderStats(providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// CheckIfRated checks if an order has been rated
func (rh *RatingHandler) CheckIfRated(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	isRated, err := rh.ratingRepo.CheckIfRated(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check rating status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_rated": isRated})
}

// GetClientRatings gets all ratings given by a client
func (rh *RatingHandler) GetClientRatings(c *gin.Context) {
	userIDStr, _, _, _, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	ratings, err := rh.ratingRepo.GetByClientID(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ratings"})
		return
	}

	// Convert to responses
	responses := make([]models.RatingResponse, len(ratings))
	for i, r := range ratings {
		responses[i] = r.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"ratings": responses,
		"count":   len(responses),
	})
}
