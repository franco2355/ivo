package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/payments-api/internal/domain/entities"
)

// PaymentEventPublisherRabbitMQ - Implementación de EventPublisher usando RabbitMQ
// Adaptador que traduce llamadas del Service a eventos de RabbitMQ
type PaymentEventPublisherRabbitMQ struct {
	rabbitMQ *RabbitMQPublisher
}

// NewPaymentEventPublisher - Constructor
func NewPaymentEventPublisher(rabbitMQ *RabbitMQPublisher) *PaymentEventPublisherRabbitMQ {
	return &PaymentEventPublisherRabbitMQ{
		rabbitMQ: rabbitMQ,
	}
}

// PublishPaymentCreated - Publica evento cuando un pago es creado (estado: pending)
func (p *PaymentEventPublisherRabbitMQ) PublishPaymentCreated(ctx context.Context, payment *entities.Payment) error {
	event := PaymentEvent{
		Action:         "payment.created",
		Type:           "payment",
		PaymentID:      payment.ID.Hex(),
		Status:         payment.Status, // "pending" cuando se crea
		EntityType:     payment.EntityType,
		EntityID:       payment.EntityID,
		UserID:         payment.UserID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		TransactionID:  payment.TransactionID,
		PaymentGateway: payment.PaymentGateway,
		Timestamp:      time.Now(),
		Metadata:       payment.Metadata,
	}

	return p.rabbitMQ.PublishPaymentEvent(ctx, event)
}

// PublishPaymentCompleted - Publica evento cuando un pago se completa
func (p *PaymentEventPublisherRabbitMQ) PublishPaymentCompleted(ctx context.Context, payment *entities.Payment) error {
	event := PaymentEvent{
		Action:         "payment.completed",
		Type:           "payment",
		PaymentID:      payment.ID.Hex(),
		Status:         "completed", // Estado actualizado
		EntityType:     payment.EntityType,     // "subscription", "inscription"
		EntityID:       payment.EntityID,       // ID de la suscripción o inscripción
		UserID:         payment.UserID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		TransactionID:  payment.TransactionID,
		PaymentGateway: payment.PaymentGateway,
		Timestamp:      time.Now(),
		Metadata:       payment.Metadata,
	}

	return p.rabbitMQ.PublishPaymentEvent(ctx, event)
}

// PublishPaymentFailed - Publica evento cuando un pago falla
func (p *PaymentEventPublisherRabbitMQ) PublishPaymentFailed(ctx context.Context, payment *entities.Payment) error {
	event := PaymentEvent{
		Action:         "payment.failed",
		Type:           "payment",
		PaymentID:      payment.ID.Hex(),
		Status:         "failed", // Estado actualizado
		EntityType:     payment.EntityType,
		EntityID:       payment.EntityID,
		UserID:         payment.UserID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		TransactionID:  payment.TransactionID,
		PaymentGateway: payment.PaymentGateway,
		Timestamp:      time.Now(),
		Metadata:       payment.Metadata,
	}

	return p.rabbitMQ.PublishPaymentEvent(ctx, event)
}

// PublishPaymentRefunded - Publica evento cuando un pago se reembolsa
func (p *PaymentEventPublisherRabbitMQ) PublishPaymentRefunded(ctx context.Context, payment *entities.Payment, refundAmount float64) error {
	// Agregar monto de reembolso a metadata
	metadata := payment.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["refund_amount"] = refundAmount

	event := PaymentEvent{
		Action:         "payment.refunded",
		Type:           "payment",
		PaymentID:      payment.ID.Hex(),
		EntityType:     payment.EntityType,
		EntityID:       payment.EntityID,
		UserID:         payment.UserID,
		Amount:         refundAmount, // Monto reembolsado
		Currency:       payment.Currency,
		TransactionID:  payment.TransactionID,
		PaymentGateway: payment.PaymentGateway,
		Timestamp:      time.Now(),
		Metadata:       metadata,
	}

	return p.rabbitMQ.PublishPaymentEvent(ctx, event)
}

// NoOpEventPublisher - Implementación vacía de EventPublisher (para testing o cuando no se usa RabbitMQ)
type NoOpEventPublisher struct{}

func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

func (n *NoOpEventPublisher) PublishPaymentCreated(ctx context.Context, payment *entities.Payment) error {
	fmt.Println("📭 No event publisher configured (payment.created event skipped)")
	return nil
}

func (n *NoOpEventPublisher) PublishPaymentCompleted(ctx context.Context, payment *entities.Payment) error {
	fmt.Println("📭 No event publisher configured (payment.completed event skipped)")
	return nil
}

func (n *NoOpEventPublisher) PublishPaymentFailed(ctx context.Context, payment *entities.Payment) error {
	fmt.Println("📭 No event publisher configured (payment.failed event skipped)")
	return nil
}

func (n *NoOpEventPublisher) PublishPaymentRefunded(ctx context.Context, payment *entities.Payment, refundAmount float64) error {
	fmt.Println("📭 No event publisher configured (payment.refunded event skipped)")
	return nil
}
