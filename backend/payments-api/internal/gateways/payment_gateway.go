package gateways

import (
	"context"
	"time"
)

// PaymentGateway - Interfaz que define el contrato que todas las pasarelas deben implementar
// Patrón: Strategy Pattern
// Permite intercambiar algoritmos de pago en runtime sin modificar la lógica de negocio
type PaymentGateway interface {
	// GetName retorna el identificador único del gateway
	GetName() string

	// CreatePayment procesa un pago en la pasarela externa
	CreatePayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error)

	// GetPaymentStatus consulta el estado actual de un pago
	GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error)

	// RefundPayment procesa un reembolso
	RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error)

	// CancelPayment cancela un pago pendiente
	CancelPayment(ctx context.Context, transactionID string) error

	// ProcessWebhook procesa notificaciones asíncronas del gateway
	ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookEvent, error)

	// ValidateCredentials verifica que las credenciales sean válidas
	ValidateCredentials(ctx context.Context) error
}

// PaymentRequest - DTO genérico para crear un pago (independiente del gateway)
type PaymentRequest struct {
	Amount         float64                // Monto a cobrar
	Currency       string                 // Moneda (USD, ARS, EUR, etc.)
	Description    string                 // Descripción del pago
	CustomerEmail  string                 // Email del cliente
	CustomerName   string                 // Nombre del cliente
	PaymentMethod  string                 // Método de pago (credit_card, debit_card, etc.)
	Metadata       map[string]interface{} // Datos adicionales específicos del negocio
	CallbackURL    string                 // URL de callback después del pago
	WebhookURL     string                 // URL para recibir notificaciones
	ExternalID     string                 // ID interno del pago (para referencia)
	CustomerID     string                 // ID del cliente en tu sistema
	ExpirationDate *time.Time             // Fecha de expiración del pago (opcional)
}

// PaymentResult - DTO genérico para el resultado de crear un pago
type PaymentResult struct {
	TransactionID string                 // ID único del pago en el gateway
	Status        string                 // Estado: pending, completed, failed
	PaymentURL    string                 // URL para completar el pago (si aplica)
	QRCode        string                 // Código QR (si aplica)
	ExternalData  map[string]interface{} // Datos adicionales específicos del gateway
	CreatedAt     time.Time              // Fecha de creación
	Message       string                 // Mensaje descriptivo
}

// PaymentStatus - DTO genérico para consultar el estado de un pago
type PaymentStatus struct {
	TransactionID  string                 // ID del pago en el gateway
	Status         string                 // Estado actual: pending, completed, failed, refunded, cancelled
	Amount         float64                // Monto del pago
	Currency       string                 // Moneda
	PaymentMethod  string                 // Método usado
	ProcessedAt    *time.Time             // Fecha de procesamiento (si ya fue procesado)
	ExternalData   map[string]interface{} // Datos adicionales del gateway
	StatusDetail   string                 // Detalle del estado (ej: "accredited", "rejected", etc.)
	FailureReason  string                 // Razón del fallo (si aplica)
}

// RefundResult - DTO genérico para el resultado de un reembolso
type RefundResult struct {
	RefundID      string                 // ID del reembolso en el gateway
	TransactionID string                 // ID del pago original
	Amount        float64                // Monto reembolsado
	Status        string                 // Estado: pending, completed, failed
	ProcessedAt   time.Time              // Fecha de procesamiento
	ExternalData  map[string]interface{} // Datos adicionales
	Message       string                 // Mensaje descriptivo
}

// WebhookEvent - DTO genérico para eventos de webhook
type WebhookEvent struct {
	EventType     string                 // Tipo de evento: payment.created, payment.updated, payment.refunded
	TransactionID string                 // ID del pago en el gateway
	Status        string                 // Nuevo estado
	Amount        float64                // Monto
	Currency      string                 // Moneda
	ProcessedAt   time.Time              // Fecha del evento
	ExternalData  map[string]interface{} // Datos adicionales del webhook
	RawPayload    []byte                 // Payload original (para debugging)
}

// Constantes para estados de pago (estandarizados)
const (
	StatusPending   = "pending"   // Pago creado pero no completado
	StatusCompleted = "completed" // Pago aprobado y acreditado
	StatusFailed    = "failed"    // Pago rechazado o fallido
	StatusRefunded  = "refunded"  // Pago reembolsado
	StatusCancelled = "cancelled" // Pago cancelado
)

// Constantes para tipos de eventos de webhook
const (
	EventPaymentCreated  = "payment.created"
	EventPaymentUpdated  = "payment.updated"
	EventPaymentRefunded = "payment.refunded"
	EventPaymentFailed   = "payment.failed"
)
