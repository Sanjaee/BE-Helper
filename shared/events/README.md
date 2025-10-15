# Shared Events Library

This is a shared library for publishing events across all microservices in the system. It provides a standardized way to publish events to RabbitMQ that the notification service can consume.

## Features

- **Standardized Event Structure**: Consistent event format across all services
- **Multiple Event Types**: Support for user, order, and payment events
- **RabbitMQ Integration**: Built-in RabbitMQ connection and publishing
- **Health Checks**: Built-in health check functionality
- **Easy Integration**: Simple API for publishing events

## Usage

### 1. Add to your service's go.mod

```go
require (
    shared/events v0.0.0
)

replace shared/events => ../shared/events
```

### 2. Import and use in your service

```go
package main

import (
    "shared/events"
)

func main() {
    // Initialize event service
    eventService, err := events.NewEventService()
    if err != nil {
        log.Fatal("Failed to initialize event service:", err)
    }
    defer eventService.Close()

    // Publish events
    err = eventService.PublishUserRegistered("user123", "john_doe", "john@example.com", "123456")
    if err != nil {
        log.Printf("Failed to publish user registered event: %v", err)
    }

    err = eventService.PublishOrderCreated("order123", "user123", "john_doe", "john@example.com", 99.99, "USD")
    if err != nil {
        log.Printf("Failed to publish order created event: %v", err)
    }
}
```

## Available Events

### User Events

#### PublishUserRegistered
```go
eventService.PublishUserRegistered(userID, username, email, otp)
```
- **Exchange**: `user.events`
- **Routing Key**: `user.registered`
- **Triggers**: OTP email notification

#### PublishUserLogin
```go
eventService.PublishUserLogin(userID, username, email)
```
- **Exchange**: `user.events`
- **Routing Key**: `user.login`
- **Triggers**: Security notification (optional)

#### PublishPasswordReset
```go
eventService.PublishPasswordReset(userID, username, email, otp)
```
- **Exchange**: `user.events`
- **Routing Key**: `password.reset`
- **Triggers**: Password reset email notification

### Order Events

#### PublishOrderCreated
```go
eventService.PublishOrderCreated(orderID, userID, username, email, amount, currency)
```
- **Exchange**: `order.events`
- **Routing Key**: `order.created`
- **Triggers**: Order confirmation email notification

### Payment Events

#### PublishPaymentSuccess
```go
eventService.PublishPaymentSuccess(paymentID, orderID, userID, username, email, amount, currency)
```
- **Exchange**: `payment.events`
- **Routing Key**: `payment.success`
- **Triggers**: Payment success email notification

#### PublishPaymentFailed
```go
eventService.PublishPaymentFailed(paymentID, orderID, userID, username, email, reason, amount, currency)
```
- **Exchange**: `payment.events`
- **Routing Key**: `payment.failed`
- **Triggers**: Payment failure email notification

## Configuration

The event service uses the following environment variables:

```bash
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=admin
RABBITMQ_PASSWORD=secret123
```

## Integration Examples

### User Service Integration

```go
// In user registration handler
func (uh *UserHandler) Register(c *gin.Context) {
    // ... user registration logic ...
    
    // Publish user registered event
    if err := uh.eventService.PublishUserRegistered(user.ID, user.Username, user.Email, otp); err != nil {
        log.Printf("Failed to publish user registered event: %v", err)
        // Don't fail the registration, just log the error
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
```

### Order Service Integration

```go
// In order creation handler
func (oh *OrderHandler) CreateOrder(c *gin.Context) {
    // ... order creation logic ...
    
    // Publish order created event
    if err := oh.eventService.PublishOrderCreated(order.ID, userID, username, email, totalAmount, "USD"); err != nil {
        log.Printf("Failed to publish order created event: %v", err)
        // Don't fail the order creation, just log the error
    }
    
    c.JSON(http.StatusOK, gin.H{"order_id": order.ID})
}
```

### Payment Service Integration

```go
// In payment success handler
func (ph *PaymentHandler) ProcessPayment(c *gin.Context) {
    // ... payment processing logic ...
    
    if payment.Success {
        // Publish payment success event
        if err := ph.eventService.PublishPaymentSuccess(payment.ID, orderID, userID, username, email, amount, "USD"); err != nil {
            log.Printf("Failed to publish payment success event: %v", err)
        }
    } else {
        // Publish payment failed event
        if err := ph.eventService.PublishPaymentFailed(payment.ID, orderID, userID, username, email, "Payment declined", amount, "USD"); err != nil {
            log.Printf("Failed to publish payment failed event: %v", err)
        }
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
```

## Error Handling

The event service is designed to be resilient. If event publishing fails, it should not break the main business logic. Always handle errors gracefully:

```go
// Good: Log error but don't fail the main operation
if err := eventService.PublishUserRegistered(userID, username, email, otp); err != nil {
    log.Printf("Failed to publish user registered event: %v", err)
    // Continue with normal flow
}

// Bad: Don't fail the main operation due to event publishing
if err := eventService.PublishUserRegistered(userID, username, email, otp); err != nil {
    return fmt.Errorf("user registration failed: %w", err)
}
```

## Health Checks

The event service includes health check functionality:

```go
if err := eventService.HealthCheck(); err != nil {
    log.Printf("Event service health check failed: %v", err)
}
```

## Adding New Events

To add new event types:

1. Define the event struct in `events.go`
2. Create a publish function following the naming convention
3. Update the notification service to handle the new event type
4. Update this documentation

Example:

```go
// 1. Add event struct
type ProductCreatedEvent struct {
    ProductID string `json:"product_id"`
    Name      string `json:"name"`
    Price     float64 `json:"price"`
}

// 2. Add publish function
func (es *EventService) PublishProductCreated(productID, name string, price float64) error {
    event := Event{
        Type: "product.created",
        Data: ProductCreatedEvent{
            ProductID: productID,
            Name:      name,
            Price:     price,
        },
        Timestamp: time.Now().Unix(),
    }
    
    return es.publishEvent("product.events", "product.created", event)
}
```

This shared library ensures consistent event publishing across all services and makes it easy to add new notification types as the system grows.
