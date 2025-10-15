# Notification Service

A comprehensive notification service that handles all types of notifications across the microservices architecture. This service supports email, SMS, and push notifications with event-driven architecture.

## Features

- **Multi-channel Notifications**: Email, SMS, and Push notifications
- **Event-driven Architecture**: Listens to events from all services
- **Template System**: Reusable notification templates
- **Scalable Design**: Easy to add new notification types and channels
- **Standardized API**: Consistent interface for all services
- **Retry Mechanism**: Automatic retry for failed notifications
- **Notification History**: Track all sent notifications
- **Health Monitoring**: Built-in health checks

## Supported Notification Types

### Email Notifications
- User Registration OTP
- Password Reset OTP
- Welcome Email
- Order Confirmation
- Payment Success
- Payment Failed

### SMS Notifications (Future)
- OTP Verification
- Password Reset
- Order Updates
- Payment Confirmations

### Push Notifications (Future)
- Order Updates
- Payment Updates
- Promotional Messages

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Service  │    │  Order Service  │    │ Payment Service │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      RabbitMQ Events      │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │  Notification Service     │
                    │  ┌─────────────────────┐  │
                    │  │   Event Consumer    │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │  Email Service      │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │  SMS Service        │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │  Push Service       │  │
                    │  └─────────────────────┘  │
                    └───────────────────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │      PostgreSQL DB        │
                    └───────────────────────────┘
```

## API Endpoints

### Send Notification
```http
POST /api/v1/notifications/send
Content-Type: application/json

{
  "user_id": "user123",
  "type": "email_otp",
  "channel": "email",
  "recipient": "user@example.com",
  "subject": "OTP Verification",
  "content": "Your OTP is: 123456",
  "service": "user-service",
  "metadata": {
    "username": "john_doe",
    "otp": "123456"
  }
}
```

### Get Notification History
```http
GET /api/v1/notifications/history?user_id=user123&limit=10&offset=0
```

### Get Templates
```http
GET /api/v1/notifications/templates?type=email_otp&channel=email
```

### Create Template
```http
POST /api/v1/notifications/templates
Content-Type: application/json

{
  "name": "Welcome Email Template",
  "type": "email_welcome",
  "channel": "email",
  "subject": "Welcome to ZACloth!",
  "content": "Hello {{username}}, welcome to ZACloth!",
  "variables": ["username"],
  "is_active": true
}
```

### Health Check
```http
GET /health
```

## Event Types

The service listens to the following events:

### User Events
- `user.registered` - New user registration
- `user.login` - User login
- `password.reset` - Password reset request

### Order Events
- `order.created` - New order created
- `order.updated` - Order status updated
- `order.cancelled` - Order cancelled

### Payment Events
- `payment.success` - Payment successful
- `payment.failed` - Payment failed
- `payment.refunded` - Payment refunded

## Configuration

### Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=notification_service
DB_PASSWORD=notificationpass
DB_NAME=notificationdb

# Server Configuration
PORT=5004
GIN_MODE=debug

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1

# RabbitMQ Configuration
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=admin
RABBITMQ_PASSWORD=secret123

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=your-email@gmail.com
FROM_NAME=ZACloth

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

## Usage in Other Services

### Publishing Events

```go
// In user-service
eventService, _ := events.NewEventService()
eventService.PublishUserRegistered(userID, username, email)

// In order-service
eventService.PublishOrderCreated(orderID, userID, username, email, amount, currency)

// In payment-service
eventService.PublishPaymentSuccess(paymentID, orderID, userID, username, email, amount, currency)
```

### Direct API Calls

```go
// Send notification directly
notificationReq := &models.NotificationRequest{
    UserID:    "user123",
    Type:      models.NotificationTypeEmailOTP,
    Channel:   models.ChannelEmail,
    Recipient: "user@example.com",
    Service:   "user-service",
    Metadata: map[string]interface{}{
        "username": "john_doe",
        "otp":      "123456",
    },
}

notification, err := notificationService.SendNotification(ctx, notificationReq)
```

## Running the Service

### Development
```bash
# Copy environment file
cp env.example .env

# Install dependencies
go mod tidy

# Run the service
go run cmd/main.go
```

### Docker
```bash
# Build and run with docker-compose
docker-compose up notification-service
```

## Database Schema

### Notifications Table
```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    recipient VARCHAR(255) NOT NULL,
    subject TEXT,
    content TEXT NOT NULL,
    template_id INTEGER REFERENCES notification_templates(id),
    metadata JSONB,
    service VARCHAR(100) NOT NULL,
    event_id VARCHAR(255),
    retry_count INTEGER DEFAULT 0,
    error_msg TEXT,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

### Notification Templates Table
```sql
CREATE TABLE notification_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    subject TEXT,
    content TEXT NOT NULL,
    variables TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

## Monitoring and Logging

The service includes comprehensive logging and monitoring:

- **Request Logging**: All HTTP requests are logged with details
- **Event Processing**: Event consumption and processing is logged
- **Error Handling**: Detailed error logging with context
- **Health Checks**: Built-in health check endpoint
- **Metrics**: Notification success/failure rates

## Future Enhancements

1. **SMS Integration**: Twilio/other SMS providers
2. **Push Notifications**: FCM and APNS support
3. **Webhook Support**: Send notifications to external webhooks
4. **Advanced Templates**: Dynamic template rendering
5. **Bulk Notifications**: Send notifications to multiple recipients
6. **Notification Preferences**: User preference management
7. **Analytics**: Notification delivery analytics
8. **Rate Limiting**: Advanced rate limiting per user/service
