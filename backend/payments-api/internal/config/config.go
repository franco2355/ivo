package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config - Configuración general del microservicio
type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	MercadoPago   MercadoPagoConfig
	Stripe        StripeConfig
	RabbitMQ      RabbitMQConfig
}

// MercadoPagoConfig - Configuración de Mercado Pago
type MercadoPagoConfig struct {
	AccessToken   string // Token de acceso privado (APP_USR-... o TEST-...)
	PublicKey     string // Public key para checkout frontend
	WebhookSecret string // Secret para validar webhooks (opcional)
}

// StripeConfig - Configuración de Stripe (futuro)
type StripeConfig struct {
	SecretKey     string
	PublicKey     string
	WebhookSecret string
}

// RabbitMQConfig - Configuración de RabbitMQ para eventos
type RabbitMQConfig struct {
	URL      string // amqp://user:password@host:port/
	Exchange string // Nombre del exchange (ej: "gym_events")
}

// LoadConfig - Carga la configuración desde variables de entorno
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Port:          getEnv("PORT", "8083"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "payments"),

		// Mercado Pago
		MercadoPago: MercadoPagoConfig{
			AccessToken:   getEnv("MERCADOPAGO_ACCESS_TOKEN", ""),
			PublicKey:     getEnv("MERCADOPAGO_PUBLIC_KEY", ""),
			WebhookSecret: getEnv("MERCADOPAGO_WEBHOOK_SECRET", ""),
		},

		// Stripe (futuro)
		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			PublicKey:     getEnv("STRIPE_PUBLIC_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},

		// RabbitMQ
		RabbitMQ: RabbitMQConfig{
			URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "gym_events"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
