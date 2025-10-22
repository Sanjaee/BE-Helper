package config

import "os"

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUsername string
	RabbitMQPassword string
	Port             string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "123"),
		DBName:           getEnv("DB_NAME", "chatdb"),
		RabbitMQHost:     getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:     getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUsername: getEnv("RABBITMQ_USERNAME", "admin"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "secret123"),
		Port:             getEnv("PORT", "5005"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

