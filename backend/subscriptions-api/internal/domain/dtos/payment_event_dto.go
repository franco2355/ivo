package dtos

import "time"

// PaymentEvent representa un evento de pago recibido desde payments-api
// Este DTO debe ser compatible con el PaymentEvent que publica payments-api
type PaymentEvent struct {
	Action         string                 `json:"action"`          // "payment.created", "payment.completed", "payment.failed", "payment.refunded"
	Type           string                 `json:"type"`            // "payment"
	PaymentID      string                 `json:"payment_id"`      // ID del pago en payments-api
	Status         string                 `json:"status"`          // "pending", "completed", "failed", "refunded"
	EntityType     string                 `json:"entity_type"`     // "subscription", "inscription"
	EntityID       string                 `json:"entity_id"`       // ID de la suscripción
	UserID         string                 `json:"user_id"`         // ID del usuario que realizó el pago
	Amount         float64                `json:"amount"`          // Monto pagado
	Currency       string                 `json:"currency"`        // "ARS", "USD", etc.
	TransactionID  string                 `json:"transaction_id"`  // ID de transacción del gateway de pago
	PaymentGateway string                 `json:"payment_gateway"` // "mercadopago", "cash", etc.
	Timestamp      time.Time              `json:"timestamp"`       // Fecha y hora del evento
	Metadata       map[string]interface{} `json:"metadata"`        // Metadata adicional del pago
}

// IsSubscriptionEvent verifica si el evento está relacionado con una suscripción
func (e *PaymentEvent) IsSubscriptionEvent() bool {
	return e.EntityType == "subscription"
}

// IsCompleted verifica si el pago fue completado exitosamente
func (e *PaymentEvent) IsCompleted() bool {
	return e.Action == "payment.completed" && e.Status == "completed"
}

// IsFailed verifica si el pago falló
func (e *PaymentEvent) IsFailed() bool {
	return e.Action == "payment.failed" && e.Status == "failed"
}

// IsRefunded verifica si el pago fue reembolsado
func (e *PaymentEvent) IsRefunded() bool {
	return e.Action == "payment.refunded" && e.Status == "refunded"
}
