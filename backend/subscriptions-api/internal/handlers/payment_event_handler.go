package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/services"
)

// PaymentEventHandler maneja eventos de pagos recibidos vía RabbitMQ
type PaymentEventHandler struct {
	subscriptionService *services.SubscriptionService
}

// NewPaymentEventHandler crea una nueva instancia del handler con DI
func NewPaymentEventHandler(subscriptionService *services.SubscriptionService) *PaymentEventHandler {
	return &PaymentEventHandler{
		subscriptionService: subscriptionService,
	}
}

// HandlePaymentCompleted procesa el evento de pago completado
// Activa la suscripción cuando el pago se completa exitosamente
func (h *PaymentEventHandler) HandlePaymentCompleted(ctx context.Context, event dtos.PaymentEvent) error {
	log.Printf("[PaymentEventHandler] 💳 Procesando pago completado: %s para suscripción: %s\n",
		event.PaymentID, event.EntityID)

	// Validar que el evento esté completo
	if event.EntityID == "" {
		return fmt.Errorf("entity_id vacío en evento de pago completado")
	}

	if !event.IsCompleted() {
		return fmt.Errorf("evento no está en estado completed: %s", event.Status)
	}

	// Activar la suscripción
	err := h.subscriptionService.ActivateSubscriptionByPayment(ctx, event.EntityID, event.PaymentID)
	if err != nil {
		log.Printf("[PaymentEventHandler] ❌ Error activando suscripción %s: %v\n", event.EntityID, err)
		return fmt.Errorf("error activando suscripción: %w", err)
	}

	log.Printf("[PaymentEventHandler] ✅ Suscripción %s activada exitosamente por pago %s\n",
		event.EntityID, event.PaymentID)

	return nil
}

// HandlePaymentFailed procesa el evento de pago fallido
// Podría registrar el intento fallido o notificar al usuario
func (h *PaymentEventHandler) HandlePaymentFailed(ctx context.Context, event dtos.PaymentEvent) error {
	log.Printf("[PaymentEventHandler] ❌ Procesando pago fallido: %s para suscripción: %s\n",
		event.PaymentID, event.EntityID)

	// Validar que el evento esté completo
	if event.EntityID == "" {
		return fmt.Errorf("entity_id vacío en evento de pago fallido")
	}

	// Registrar el fallo en la suscripción
	err := h.subscriptionService.RegisterPaymentFailure(ctx, event.EntityID, event.PaymentID)
	if err != nil {
		log.Printf("[PaymentEventHandler] ⚠️  Error registrando fallo de pago para suscripción %s: %v\n",
			event.EntityID, err)
		// No retornamos error para no requeue, solo logging
		return nil
	}

	log.Printf("[PaymentEventHandler] 📝 Fallo de pago registrado para suscripción %s\n", event.EntityID)

	return nil
}

// HandlePaymentRefunded procesa el evento de pago reembolsado
// Cancela o desactiva la suscripción cuando se reembolsa el pago
func (h *PaymentEventHandler) HandlePaymentRefunded(ctx context.Context, event dtos.PaymentEvent) error {
	log.Printf("[PaymentEventHandler] 💰 Procesando pago reembolsado: %s para suscripción: %s\n",
		event.PaymentID, event.EntityID)

	// Validar que el evento esté completo
	if event.EntityID == "" {
		return fmt.Errorf("entity_id vacío en evento de pago reembolsado")
	}

	if !event.IsRefunded() {
		return fmt.Errorf("evento no está en estado refunded: %s", event.Status)
	}

	// Cancelar la suscripción por reembolso
	err := h.subscriptionService.CancelSubscriptionByRefund(ctx, event.EntityID, event.PaymentID)
	if err != nil {
		log.Printf("[PaymentEventHandler] ❌ Error cancelando suscripción %s por reembolso: %v\n",
			event.EntityID, err)
		return fmt.Errorf("error cancelando suscripción por reembolso: %w", err)
	}

	log.Printf("[PaymentEventHandler] ✅ Suscripción %s cancelada exitosamente por reembolso de pago %s\n",
		event.EntityID, event.PaymentID)

	return nil
}
