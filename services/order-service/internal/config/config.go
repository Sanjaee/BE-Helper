package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// RabbitMQ
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUsername string
	RabbitMQPassword string

	// Gin Mode
	GinMode string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "5002"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "123"),
		DBName:     getEnv("DB_NAME", "orderdb"),

		// RabbitMQ
		RabbitMQHost:     getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:     getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUsername: getEnv("RABBITMQ_USERNAME", "admin"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "secret123"),

		// Gin Mode
		GinMode: getEnv("GIN_MODE", "debug"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
