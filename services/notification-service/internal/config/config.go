package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the notification service
type Config struct {
	// Server configuration
	Port    string
	GinMode string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// RabbitMQ configuration
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUsername string
	RabbitMQPassword string

	// SMTP configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string

	// SMS configuration (for future use)
	SMSProvider        string
	TwilioAccountSID   string
	TwilioAuthToken    string
	TwilioPhoneNumber  string

	// Push notification configuration (for future use)
	FCMServerKey string
	APNSKeyID    string
	APNSTeamID   string
	APNSKeyPath  string

	// Logging configuration
	LogLevel  string
	LogFormat string

	// Rate limiting configuration
	RateLimitEnabled            bool
	RateLimitRequestsPerMinute  int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Server configuration
		Port:    getEnv("PORT", "5004"),
		GinMode: getEnv("GIN_MODE", "debug"),

		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "notification_service"),
		DBPassword: getEnv("DB_PASSWORD", "notificationpass"),
		DBName:     getEnv("DB_NAME", "notificationdb"),

		// Redis configuration
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 1),

		// RabbitMQ configuration
		RabbitMQHost:     getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:     getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUsername: getEnv("RABBITMQ_USERNAME", "admin"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "secret123"),

		// SMTP configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),
		FromName:     getEnv("FROM_NAME", "ZACloth"),

		// SMS configuration
		SMSProvider:       getEnv("SMS_PROVIDER", "twilio"),
		TwilioAccountSID:  getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:   getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber: getEnv("TWILIO_PHONE_NUMBER", ""),

		// Push notification configuration
		FCMServerKey: getEnv("FCM_SERVER_KEY", ""),
		APNSKeyID:    getEnv("APNS_KEY_ID", ""),
		APNSTeamID:   getEnv("APNS_TEAM_ID", ""),
		APNSKeyPath:  getEnv("APNS_KEY_PATH", ""),

		// Logging configuration
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),

		// Rate limiting configuration
		RateLimitEnabled:           getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
	}

	// Validate required configuration
	if cfg.SMTPUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME is required")
	}
	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
