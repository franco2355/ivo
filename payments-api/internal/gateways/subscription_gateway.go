package gateways

import (
	"context"
	"time"
)

// SubscriptionGateway - Interfaz para manejar suscripciones recurrentes (débitos automáticos)
// Implementada por: MercadoPago Preapprovals, Stripe Subscriptions, etc.
type SubscriptionGateway interface {
	// GetName retorna el identificador único del gateway
	GetName() string

	// CreateSubscription crea una suscripción recurrente en el gateway
	CreateSubscription(ctx context.Context, request SubscriptionRequest) (*SubscriptionResult, error)

	// GetSubscriptionStatus consulta el estado actual de una suscripción
	GetSubscriptionStatus(ctx context.Context, subscriptionID string) (*SubscriptionStatus, error)

	// CancelSubscription cancela una suscripción activa
	CancelSubscription(ctx context.Context, subscriptionID string) error

	// PauseSubscription pausa una suscripción (opcional, no todos los gateways lo soportan)
	PauseSubscription(ctx context.Context, subscriptionID string) error

	// ResumeSubscription reanuda una suscripción pausada (opcional)
	ResumeSubscription(ctx context.Context, subscriptionID string) error

	// ProcessSubscriptionWebhook procesa webhooks de suscripciones
	ProcessSubscriptionWebhook(ctx context.Context, payload []byte, headers map[string]string) (*SubscriptionWebhookEvent, error)

	// ValidateCredentials verifica que las credenciales sean válidas
	ValidateCredentials(ctx context.Context) error
}

// SubscriptionRequest - DTO para crear una suscripción
type SubscriptionRequest struct {
	Reason         string                 // Descripción de la suscripción (ej: "Cuota mensual gimnasio")
	Amount         float64                // Monto a cobrar en cada período
	Currency       string                 // Moneda (ARS, USD, etc.)
	Frequency      int                    // Frecuencia (1, 2, 3...)
	FrequencyType  string                 // Tipo de frecuencia (months, days, weeks)
	StartDate      *time.Time             // Fecha de inicio (opcional, default: ahora)
	EndDate        *time.Time             // Fecha de fin (opcional, infinito por defecto)
	CustomerEmail  string                 // Email del cliente
	CustomerName   string                 // Nombre del cliente
	CustomerID     string                 // ID del cliente en tu sistema
	ExternalID     string                 // ID de la suscripción en tu sistema
	CallbackURL    string                 // URL de callback
	WebhookURL     string                 // URL para recibir notificaciones de cobros
	Metadata       map[string]interface{} // Datos adicionales
}

// SubscriptionResult - DTO para el resultado de crear suscripción
type SubscriptionResult struct {
	SubscriptionID string                 // ID de la suscripción en el gateway
	Status         string                 // Estado: pending, authorized, paused, cancelled
	InitPoint      string                 // URL para que el usuario autorice la suscripción
	ExternalData   map[string]interface{} // Datos adicionales del gateway
	CreatedAt      time.Time              // Fecha de creación
	Message        string                 // Mensaje descriptivo
}

// SubscriptionStatus - DTO para consultar el estado de una suscripción
type SubscriptionStatus struct {
	SubscriptionID  string                 // ID de la suscripción
	Status          string                 // Estado actual
	Reason          string                 // Descripción
	Amount          float64                // Monto por período
	Currency        string                 // Moneda
	Frequency       int                    // Frecuencia
	FrequencyType   string                 // Tipo de frecuencia
	NextPaymentDate *time.Time             // Próxima fecha de cobro
	LastPaymentDate *time.Time             // Última fecha de cobro
	TotalCharges    int                    // Cantidad de cobros realizados
	ExternalData    map[string]interface{} // Datos adicionales
}

// SubscriptionWebhookEvent - DTO para eventos de webhook de suscripciones
type SubscriptionWebhookEvent struct {
	EventType      string                 // Tipo: subscription.created, subscription.charged, subscription.cancelled
	SubscriptionID string                 // ID de la suscripción
	PaymentID      string                 // ID del pago (solo para eventos de cobro)
	Status         string                 // Nuevo estado
	Amount         float64                // Monto (si aplica)
	Currency       string                 // Moneda
	ProcessedAt    time.Time              // Fecha del evento
	ExternalData   map[string]interface{} // Datos adicionales
	RawPayload     []byte                 // Payload original
}

// Constantes para estados de suscripciones
const (
	SubscriptionStatusPending    = "pending"    // Suscripción creada, esperando autorización
	SubscriptionStatusAuthorized = "authorized" // Suscripción autorizada y activa
	SubscriptionStatusPaused     = "paused"     // Suscripción pausada
	SubscriptionStatusCancelled  = "cancelled"  // Suscripción cancelada
)

// Constantes para tipos de eventos de suscripciones
const (
	EventSubscriptionCreated   = "subscription.created"   // Suscripción creada
	EventSubscriptionAuthorized = "subscription.authorized" // Usuario autorizó la suscripción
	EventSubscriptionCharged   = "subscription.charged"   // Se realizó un cobro
	EventSubscriptionFailed    = "subscription.failed"    // Cobro falló
	EventSubscriptionPaused    = "subscription.paused"    // Suscripción pausada
	EventSubscriptionCancelled = "subscription.cancelled" // Suscripción cancelada
)

// Constantes para tipos de frecuencia
const (
	FrequencyTypeMonths = "months"
	FrequencyTypeDays   = "days"
	FrequencyTypeWeeks  = "weeks"
)
