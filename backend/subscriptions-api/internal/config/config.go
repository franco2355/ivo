package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	MongoURI          string
	MongoDatabase     string
	RabbitMQURL       string
	RabbitMQExchange  string
	UsersAPIURL       string
	PaymentsAPIURL    string
	JWTSecret         string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Port:             getEnv("PORT", "8081"),
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:    getEnv("MONGO_DATABASE", "gym_subscriptions"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "gym_events"),
		UsersAPIURL:      getEnv("USERS_API_URL", "http://localhost:8080"),
		PaymentsAPIURL:   getEnv("PAYMENTS_API_URL", "http://localhost:8083"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
