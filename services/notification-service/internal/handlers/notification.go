package handlers

import (
	"net/http"
	"strconv"

	"notification-service/internal/models"
	"notification-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService *services.NotificationService
	logger              *logrus.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *services.NotificationService, logger *logrus.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// SendNotificationRequest represents the request to send a notification
type SendNotificationRequest struct {
	UserID     string                     `json:"user_id" validate:"required"`
	Type       models.NotificationType    `json:"type" validate:"required"`
	Channel    models.NotificationChannel `json:"channel" validate:"required"`
	Recipient  string                     `json:"recipient" validate:"required"`
	Subject    string                     `json:"subject,omitempty"`
	Content    string                     `json:"content,omitempty"`
	TemplateID *uint                      `json:"template_id,omitempty"`
	Metadata   map[string]interface{}     `json:"metadata,omitempty"`
	Service    string                     `json:"service" validate:"required"`
	EventID    string                     `json:"event_id,omitempty"`
}

// SendNotification sends a notification
func (nh *NotificationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		nh.logger.Errorf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Convert to internal request
	notificationReq := &models.NotificationRequest{
		UserID:     req.UserID,
		Type:       req.Type,
		Channel:    req.Channel,
		Recipient:  req.Recipient,
		Subject:    req.Subject,
		Content:    req.Content,
		TemplateID: req.TemplateID,
		Metadata:   req.Metadata,
		Service:    req.Service,
		EventID:    req.EventID,
	}

	// Send notification
	notification, err := nh.notificationService.SendNotification(c.Request.Context(), notificationReq)
	if err != nil {
		nh.logger.Errorf("Failed to send notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification sent successfully",
		"data":    notification,
	})
}

// GetNotificationHistory retrieves notification history
func (nh *NotificationHandler) GetNotificationHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	// Validate pagination limits
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get notification history
	history, err := nh.notificationService.GetNotificationHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		nh.logger.Errorf("Failed to get notification history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(history),
		},
	})
}

// GetTemplates retrieves notification templates
func (nh *NotificationHandler) GetTemplates(c *gin.Context) {
	var notificationType *models.NotificationType
	var channel *models.NotificationChannel

	// Parse optional filters
	if typeStr := c.Query("type"); typeStr != "" {
		nt := models.NotificationType(typeStr)
		notificationType = &nt
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		ch := models.NotificationChannel(channelStr)
		channel = &ch
	}

	// Get templates
	templates, err := nh.notificationService.GetTemplates(c.Request.Context(), notificationType, channel)
	if err != nil {
		nh.logger.Errorf("Failed to get templates: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
	})
}

// CreateTemplateRequest represents the request to create a template
type CreateTemplateRequest struct {
	Name      string                     `json:"name" validate:"required"`
	Type      models.NotificationType    `json:"type" validate:"required"`
	Channel   models.NotificationChannel `json:"channel" validate:"required"`
	Subject   string                     `json:"subject,omitempty"`
	Content   string                     `json:"content" validate:"required"`
	Variables []string                   `json:"variables,omitempty"`
	IsActive  bool                       `json:"is_active"`
}

// CreateTemplate creates a new notification template
func (nh *NotificationHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		nh.logger.Errorf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Create template
	template := &models.NotificationTemplate{
		Name:      req.Name,
		Type:      req.Type,
		Channel:   req.Channel,
		Subject:   req.Subject,
		Content:   req.Content,
		Variables: req.Variables,
		IsActive:  req.IsActive,
	}

	if err := nh.notificationService.CreateTemplate(c.Request.Context(), template); err != nil {
		nh.logger.Errorf("Failed to create template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Template created successfully",
		"data":    template,
	})
}
