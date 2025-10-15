package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// Email notification types
	NotificationTypeEmailOTP              NotificationType = "email_otp"
	NotificationTypeEmailPasswordReset    NotificationType = "email_password_reset"
	NotificationTypeEmailWelcome          NotificationType = "email_welcome"
	NotificationTypeEmailOrderConfirmation NotificationType = "email_order_confirmation"
	NotificationTypeEmailPaymentSuccess   NotificationType = "email_payment_success"
	NotificationTypeEmailPaymentFailed    NotificationType = "email_payment_failed"

	// SMS notification types
	NotificationTypeSMSOTP                NotificationType = "sms_otp"
	NotificationTypeSMSPasswordReset      NotificationType = "sms_password_reset"
	NotificationTypeSMSOrderConfirmation  NotificationType = "sms_order_confirmation"
	NotificationTypeSMSPaymentSuccess     NotificationType = "sms_payment_success"

	// Push notification types
	NotificationTypePushOrderUpdate       NotificationType = "push_order_update"
	NotificationTypePushPaymentUpdate     NotificationType = "push_payment_update"
	NotificationTypePushPromotion         NotificationType = "push_promotion"
)

// NotificationChannel represents the channel through which notification is sent
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusDelivered NotificationStatus = "delivered"
	StatusRead      NotificationStatus = "read"
)

// Notification represents a notification record in the database
type Notification struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	UserID      string                 `json:"user_id" gorm:"index;not null"`
	Type        NotificationType       `json:"type" gorm:"not null"`
	Channel     NotificationChannel    `json:"channel" gorm:"not null"`
	Status      NotificationStatus     `json:"status" gorm:"default:'pending'"`
	Recipient   string                 `json:"recipient" gorm:"not null"` // email, phone, device token
	Subject     string                 `json:"subject,omitempty"`
	Content     string                 `json:"content" gorm:"type:text"`
	TemplateID  *uint                  `json:"template_id,omitempty"`
	Template    *NotificationTemplate  `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	Service     string                 `json:"service" gorm:"not null"` // source service (user-service, order-service, etc.)
	EventID     string                 `json:"event_id,omitempty"` // original event ID
	RetryCount  int                    `json:"retry_count" gorm:"default:0"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
}

// NotificationTemplate represents a reusable notification template
type NotificationTemplate struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"uniqueIndex;not null"`
	Type        NotificationType  `json:"type" gorm:"not null"`
	Channel     NotificationChannel `json:"channel" gorm:"not null"`
	Subject     string            `json:"subject,omitempty"`
	Content     string            `json:"content" gorm:"type:text;not null"`
	Variables   []string          `json:"variables" gorm:"type:text[]"` // list of template variables
	IsActive    bool              `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	UserID    string                 `json:"user_id" validate:"required"`
	Type      NotificationType       `json:"type" validate:"required"`
	Channel   NotificationChannel    `json:"channel" validate:"required"`
	Recipient string                 `json:"recipient" validate:"required"`
	Subject   string                 `json:"subject,omitempty"`
	Content   string                 `json:"content,omitempty"`
	TemplateID *uint                 `json:"template_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Service   string                 `json:"service" validate:"required"`
	EventID   string                 `json:"event_id,omitempty"`
}

// NotificationHistory represents notification history for API responses
type NotificationHistory struct {
	ID          uint                `json:"id"`
	UserID      string              `json:"user_id"`
	Type        NotificationType    `json:"type"`
	Channel     NotificationChannel `json:"channel"`
	Status      NotificationStatus  `json:"status"`
	Recipient   string              `json:"recipient"`
	Subject     string              `json:"subject,omitempty"`
	Content     string              `json:"content"`
	Service     string              `json:"service"`
	RetryCount  int                 `json:"retry_count"`
	ErrorMsg    string              `json:"error_msg,omitempty"`
	SentAt      *time.Time          `json:"sent_at,omitempty"`
	DeliveredAt *time.Time          `json:"delivered_at,omitempty"`
	ReadAt      *time.Time          `json:"read_at,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
}

// EventData represents data from various service events
type EventData struct {
	UserID    string                 `json:"user_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	Email     string                 `json:"email,omitempty"`
	Phone     string                 `json:"phone,omitempty"`
	OrderID   string                 `json:"order_id,omitempty"`
	PaymentID string                 `json:"payment_id,omitempty"`
	Amount    float64                `json:"amount,omitempty"`
	Currency  string                 `json:"currency,omitempty"`
	OTP       string                 `json:"otp,omitempty"`
	Token     string                 `json:"token,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
