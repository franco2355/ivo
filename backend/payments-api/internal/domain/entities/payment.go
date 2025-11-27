package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Payment - Entidad de base de datos para pagos
// Modelo agnóstico del dominio que puede usarse para cualquier tipo de transacción
// Soporta pagos únicos (Checkout Pro) y recurrentes (Preapprovals)
type Payment struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty"`
	EntityType     string                 `bson:"entity_type"`     // subscription, inscription, plan_upgrade, etc.
	EntityID       string                 `bson:"entity_id"`       // ID de la entidad asociada
	UserID         string                 `bson:"user_id"`         // ID del usuario que realiza el pago
	Amount         float64                `bson:"amount"`          // Monto del pago
	Currency       string                 `bson:"currency"`        // USD, ARS, EUR
	Status         string                 `bson:"status"`          // pending, completed, failed, refunded
	PaymentMethod  string                 `bson:"payment_method"`  // credit_card, debit_card, cash, transfer
	PaymentGateway string                 `bson:"payment_gateway"` // stripe, mercadopago, manual
	TransactionID  string                 `bson:"transaction_id"`  // ID de transacción/preapproval del gateway
	PaymentType    string                 `bson:"payment_type,omitempty"` // "one_time" o "recurring" (opcional)
	IdempotencyKey string                 `bson:"idempotency_key,omitempty"` // UUID para prevenir duplicados (idempotencia)
	Metadata       map[string]interface{} `bson:"metadata"`        // Información adicional específica del dominio
	CreatedAt      time.Time              `bson:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at"`
	ProcessedAt    *time.Time             `bson:"processed_at"` // Fecha de procesamiento del pago
}
