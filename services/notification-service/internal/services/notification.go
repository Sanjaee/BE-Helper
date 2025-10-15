package services

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NotificationService handles all notification operations
type NotificationService struct {
	db           *gorm.DB
	emailService *EmailService
	logger       *logrus.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailService *EmailService, logger *logrus.Logger) *NotificationService {
	return &NotificationService{
		emailService: emailService,
		logger:       logger,
	}
}

// InitializeDB initializes the database connection
func (ns *NotificationService) InitializeDB(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&models.Notification{}, &models.NotificationTemplate{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	ns.db = db
	return nil
}

// SendNotification sends a notification based on the request
func (ns *NotificationService) SendNotification(ctx context.Context, req *models.NotificationRequest) (*models.Notification, error) {
	// Create notification record
	notification := &models.Notification{
		UserID:     req.UserID,
		Type:       req.Type,
		Channel:    req.Channel,
		Status:     models.StatusPending,
		Recipient:  req.Recipient,
		Subject:    req.Subject,
		Content:    req.Content,
		TemplateID: req.TemplateID,
		Metadata:   req.Metadata,
		Service:    req.Service,
		EventID:    req.EventID,
	}

	// Save to database
	if err := ns.db.WithContext(ctx).Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification record: %w", err)
	}

	// Send notification based on channel
	switch req.Channel {
	case models.ChannelEmail:
		if err := ns.sendEmailNotification(ctx, notification); err != nil {
			ns.updateNotificationStatus(notification.ID, models.StatusFailed, err.Error())
			return notification, err
		}
	case models.ChannelSMS:
		// TODO: Implement SMS sending
		ns.logger.Warn("SMS notifications not yet implemented")
		ns.updateNotificationStatus(notification.ID, models.StatusFailed, "SMS not implemented")
	case models.ChannelPush:
		// TODO: Implement push notifications
		ns.logger.Warn("Push notifications not yet implemented")
		ns.updateNotificationStatus(notification.ID, models.StatusFailed, "Push notifications not implemented")
	default:
		err := fmt.Errorf("unsupported notification channel: %s", req.Channel)
		ns.updateNotificationStatus(notification.ID, models.StatusFailed, err.Error())
		return notification, err
	}

	// Update status to sent
	ns.updateNotificationStatus(notification.ID, models.StatusSent, "")

	// Update sent timestamp
	now := time.Now()
	ns.db.Model(notification).Updates(map[string]interface{}{
		"sent_at": &now,
	})

	return notification, nil
}

// sendEmailNotification sends email notification
func (ns *NotificationService) sendEmailNotification(ctx context.Context, notification *models.Notification) error {
	switch notification.Type {
	case models.NotificationTypeEmailOTP:
		// Extract OTP from metadata
		otp, ok := notification.Metadata["otp"].(string)
		if !ok {
			return fmt.Errorf("OTP not found in metadata")
		}
		username, _ := notification.Metadata["username"].(string)
		return ns.emailService.SendOTPEmail(notification.Recipient, username, otp)

	case models.NotificationTypeEmailPasswordReset:
		// Extract OTP from metadata
		otp, ok := notification.Metadata["otp"].(string)
		if !ok {
			return fmt.Errorf("OTP not found in metadata")
		}
		username, _ := notification.Metadata["username"].(string)
		return ns.emailService.SendPasswordResetEmail(notification.Recipient, username, otp)

	case models.NotificationTypeEmailWelcome:
		username, _ := notification.Metadata["username"].(string)
		return ns.emailService.SendWelcomeEmail(notification.Recipient, username)

	case models.NotificationTypeEmailOrderConfirmation:
		orderID, _ := notification.Metadata["order_id"].(string)
		amount, _ := notification.Metadata["amount"].(float64)
		currency, _ := notification.Metadata["currency"].(string)
		username, _ := notification.Metadata["username"].(string)
		return ns.emailService.SendOrderConfirmationEmail(notification.Recipient, username, orderID, amount, currency)

	case models.NotificationTypeEmailPaymentSuccess:
		orderID, _ := notification.Metadata["order_id"].(string)
		paymentID, _ := notification.Metadata["payment_id"].(string)
		amount, _ := notification.Metadata["amount"].(float64)
		currency, _ := notification.Metadata["currency"].(string)
		username, _ := notification.Metadata["username"].(string)
		return ns.emailService.SendPaymentSuccessEmail(notification.Recipient, username, orderID, paymentID, amount, currency)

	case models.NotificationTypeEmailPaymentFailed:
		orderID, _ := notification.Metadata["order_id"].(string)
		reason, _ := notification.Metadata["reason"].(string)
		username, _ := notification.Metadata["username"].(string)
		return ns.emailService.SendPaymentFailedEmail(notification.Recipient, username, orderID, reason)

	default:
		// Generic email with custom content
		return ns.emailService.SendEmail(EmailData{
			To:      notification.Recipient,
			Subject: notification.Subject,
			Body:    notification.Content,
		})
	}
}

// ProcessEvent processes events from other services and sends appropriate notifications
func (ns *NotificationService) ProcessEvent(ctx context.Context, eventType string, eventData *models.EventData) error {
	ns.logger.Infof("Processing event: %s", eventType)

	switch eventType {
	case "user.registered":
		return ns.handleUserRegistered(ctx, eventData)
	case "user.login":
		return ns.handleUserLogin(ctx, eventData)
	case "password.reset":
		return ns.handlePasswordReset(ctx, eventData)
	case "order.created":
		return ns.handleOrderCreated(ctx, eventData)
	case "payment.success":
		return ns.handlePaymentSuccess(ctx, eventData)
	case "payment.failed":
		return ns.handlePaymentFailed(ctx, eventData)
	default:
		ns.logger.Warnf("Unknown event type: %s", eventType)
		return nil
	}
}

// handleUserRegistered handles user registration event
func (ns *NotificationService) handleUserRegistered(ctx context.Context, data *models.EventData) error {
	// Send OTP email for email verification
	req := &models.NotificationRequest{
		UserID:    data.UserID,
		Type:      models.NotificationTypeEmailOTP,
		Channel:   models.ChannelEmail,
		Recipient: data.Email,
		Subject:   "Verifikasi Email - ZACloth",
		Service:   "user-service",
		Metadata: map[string]interface{}{
			"username": data.Username,
			"otp":      data.OTP,
		},
	}

	_, err := ns.SendNotification(ctx, req)
	return err
}

// handleUserLogin handles user login event
func (ns *NotificationService) handleUserLogin(ctx context.Context, data *models.EventData) error {
	// For now, we don't send notifications for login events
	// This can be enabled for security notifications if needed
	ns.logger.Infof("User login: %s (%s)", data.Username, data.Email)
	return nil
}

// handlePasswordReset handles password reset event
func (ns *NotificationService) handlePasswordReset(ctx context.Context, data *models.EventData) error {
	req := &models.NotificationRequest{
		UserID:    data.UserID,
		Type:      models.NotificationTypeEmailPasswordReset,
		Channel:   models.ChannelEmail,
		Recipient: data.Email,
		Subject:   "Reset Password - ZACloth",
		Service:   "user-service",
		Metadata: map[string]interface{}{
			"username": data.Username,
			"otp":      data.OTP,
		},
	}

	_, err := ns.SendNotification(ctx, req)
	return err
}

// handleOrderCreated handles order creation event
func (ns *NotificationService) handleOrderCreated(ctx context.Context, data *models.EventData) error {
	req := &models.NotificationRequest{
		UserID:    data.UserID,
		Type:      models.NotificationTypeEmailOrderConfirmation,
		Channel:   models.ChannelEmail,
		Recipient: data.Email,
		Service:   "order-service",
		Metadata: map[string]interface{}{
			"username": data.Username,
			"order_id": data.OrderID,
			"amount":   data.Amount,
			"currency": data.Currency,
		},
	}

	_, err := ns.SendNotification(ctx, req)
	return err
}

// handlePaymentSuccess handles payment success event
func (ns *NotificationService) handlePaymentSuccess(ctx context.Context, data *models.EventData) error {
	req := &models.NotificationRequest{
		UserID:    data.UserID,
		Type:      models.NotificationTypeEmailPaymentSuccess,
		Channel:   models.ChannelEmail,
		Recipient: data.Email,
		Service:   "payment-service",
		Metadata: map[string]interface{}{
			"username":   data.Username,
			"order_id":   data.OrderID,
			"payment_id": data.PaymentID,
			"amount":     data.Amount,
			"currency":   data.Currency,
		},
	}

	_, err := ns.SendNotification(ctx, req)
	return err
}

// handlePaymentFailed handles payment failure event
func (ns *NotificationService) handlePaymentFailed(ctx context.Context, data *models.EventData) error {
	req := &models.NotificationRequest{
		UserID:    data.UserID,
		Type:      models.NotificationTypeEmailPaymentFailed,
		Channel:   models.ChannelEmail,
		Recipient: data.Email,
		Service:   "payment-service",
		Metadata: map[string]interface{}{
			"username": data.Username,
			"order_id": data.OrderID,
			"reason":   data.Message,
		},
	}

	_, err := ns.SendNotification(ctx, req)
	return err
}

// updateNotificationStatus updates the status of a notification
func (ns *NotificationService) updateNotificationStatus(id uint, status models.NotificationStatus, errorMsg string) {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
		updates["retry_count"] = gorm.Expr("retry_count + 1")
	}

	if status == models.StatusDelivered {
		now := time.Now()
		updates["delivered_at"] = &now
	}

	ns.db.Model(&models.Notification{}).Where("id = ?", id).Updates(updates)
}

// GetNotificationHistory retrieves notification history for a user
func (ns *NotificationService) GetNotificationHistory(ctx context.Context, userID string, limit, offset int) ([]*models.NotificationHistory, error) {
	var notifications []*models.NotificationHistory

	query := ns.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notification history: %w", err)
	}

	return notifications, nil
}

// GetTemplates retrieves notification templates
func (ns *NotificationService) GetTemplates(ctx context.Context, notificationType *models.NotificationType, channel *models.NotificationChannel) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate

	query := ns.db.WithContext(ctx).Where("is_active = ?", true)

	if notificationType != nil {
		query = query.Where("type = ?", *notificationType)
	}

	if channel != nil {
		query = query.Where("channel = ?", *channel)
	}

	if err := query.Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	return templates, nil
}

// CreateTemplate creates a new notification template
func (ns *NotificationService) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	if err := ns.db.WithContext(ctx).Create(template).Error; err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	return nil
}

// HealthCheck checks if the notification service is healthy
func (ns *NotificationService) HealthCheck() error {
	if ns.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Test database connection
	sqlDB, err := ns.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check email service
	if err := ns.emailService.HealthCheck(); err != nil {
		return fmt.Errorf("email service health check failed: %w", err)
	}

	return nil
}
