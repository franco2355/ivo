package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// HealthService - Servicio para health checks
type HealthService struct {
	mongoClient    *mongo.Client
	eventPublisher EventPublisher
}

// NewHealthService - Constructor con DI
func NewHealthService(mongoClient *mongo.Client, eventPublisher EventPublisher) *HealthService {
	return &HealthService{
		mongoClient:    mongoClient,
		eventPublisher: eventPublisher,
	}
}

// HealthStatus - Estado de health check
type HealthStatus struct {
	Status   string            `json:"status"`
	Service  string            `json:"service"`
	Checks   map[string]string `json:"checks"`
	Uptime   string            `json:"uptime,omitempty"`
	Version  string            `json:"version,omitempty"`
}

var startTime = time.Now()

// CheckHealth - Verifica el estado del servicio y sus dependencias
func (s *HealthService) CheckHealth(ctx context.Context) *HealthStatus {
	checks := make(map[string]string)
	overallStatus := "healthy"

	// 1. Verificar MongoDB
	mongoStatus := s.checkMongoDB(ctx)
	checks["mongodb"] = mongoStatus
	if mongoStatus != "healthy" {
		overallStatus = "degraded"
	}

	// 2. Verificar RabbitMQ (EventPublisher)
	rabbitStatus := s.checkRabbitMQ()
	checks["rabbitmq"] = rabbitStatus
	if rabbitStatus != "healthy" && rabbitStatus != "disabled" {
		overallStatus = "degraded"
	}

	// 3. Calcular uptime
	uptime := time.Since(startTime).String()

	return &HealthStatus{
		Status:  overallStatus,
		Service: "subscriptions-api",
		Checks:  checks,
		Uptime:  uptime,
		Version: "1.0.0",
	}
}

// checkMongoDB - Verifica conexión a MongoDB
func (s *HealthService) checkMongoDB(ctx context.Context) string {
	if s.mongoClient == nil {
		return "unavailable"
	}

	// Timeout de 2 segundos para el ping
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := s.mongoClient.Ping(ctxWithTimeout, nil)
	if err != nil {
		return "unhealthy"
	}

	return "healthy"
}

// checkRabbitMQ - Verifica estado de RabbitMQ
func (s *HealthService) checkRabbitMQ() string {
	if s.eventPublisher == nil {
		return "disabled"
	}

	// Si el publisher existe, asumimos que está conectado
	// (la conexión falla en el startup si no está disponible)
	return "healthy"
}
