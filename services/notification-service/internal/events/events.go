package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event represents a generic event structure
type Event struct {
	Type      string      `json:"type"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// UserRegisteredEvent represents user registration event
type UserRegisteredEvent struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	OTP      string `json:"otp,omitempty"`
}

// UserLoginEvent represents user login event
type UserLoginEvent struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// PasswordResetEvent represents password reset event
type PasswordResetEvent struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	OTP      string `json:"otp,omitempty"`
}

// OrderCreatedEvent represents order creation event
type OrderCreatedEvent struct {
	OrderID  string  `json:"order_id"`
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// PaymentSuccessEvent represents payment success event
type PaymentSuccessEvent struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

// PaymentFailedEvent represents payment failure event
type PaymentFailedEvent struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Reason    string  `json:"reason"`
	Amount    float64 `json:"amount,omitempty"`
	Currency  string  `json:"currency,omitempty"`
}

// CreateUserRegisteredEvent creates a user registered event
func CreateUserRegisteredEvent(userID, username, email, otp string) *Event {
	return &Event{
		Type:      "user.registered",
		UserID:    userID,
		Data:      UserRegisteredEvent{UserID: userID, Username: username, Email: email, OTP: otp},
		Timestamp: time.Now().Unix(),
	}
}

// CreateUserLoginEvent creates a user login event
func CreateUserLoginEvent(userID, username, email string) *Event {
	return &Event{
		Type:      "user.login",
		UserID:    userID,
		Data:      UserLoginEvent{UserID: userID, Username: username, Email: email},
		Timestamp: time.Now().Unix(),
	}
}

// CreatePasswordResetEvent creates a password reset event
func CreatePasswordResetEvent(userID, username, email, otp string) *Event {
	return &Event{
		Type:      "password.reset",
		UserID:    userID,
		Data:      PasswordResetEvent{UserID: userID, Username: username, Email: email, OTP: otp},
		Timestamp: time.Now().Unix(),
	}
}

// CreateOrderCreatedEvent creates an order created event
func CreateOrderCreatedEvent(orderID, userID, username, email string, amount float64, currency string) *Event {
	return &Event{
		Type:      "order.created",
		UserID:    userID,
		Data:      OrderCreatedEvent{OrderID: orderID, UserID: userID, Username: username, Email: email, Amount: amount, Currency: currency},
		Timestamp: time.Now().Unix(),
	}
}

// CreatePaymentSuccessEvent creates a payment success event
func CreatePaymentSuccessEvent(paymentID, orderID, userID, username, email string, amount float64, currency string) *Event {
	return &Event{
		Type:      "payment.success",
		UserID:    userID,
		Data:      PaymentSuccessEvent{PaymentID: paymentID, OrderID: orderID, UserID: userID, Username: username, Email: email, Amount: amount, Currency: currency},
		Timestamp: time.Now().Unix(),
	}
}

// CreatePaymentFailedEvent creates a payment failed event
func CreatePaymentFailedEvent(paymentID, orderID, userID, username, email, reason string, amount float64, currency string) *Event {
	return &Event{
		Type:      "payment.failed",
		UserID:    userID,
		Data:      PaymentFailedEvent{PaymentID: paymentID, OrderID: orderID, UserID: userID, Username: username, Email: email, Reason: reason, Amount: amount, Currency: currency},
		Timestamp: time.Now().Unix(),
	}
}

// ToJSON converts the event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON
func FromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}
