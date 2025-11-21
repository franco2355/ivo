package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subscription - Entidad que representa una suscripción recurrente en el sistema
type Subscription struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID         string                 `bson:"user_id" json:"user_id"`                   // ID del usuario suscrito
	Reason         string                 `bson:"reason" json:"reason"`                     // Descripción (ej: "Cuota mensual gimnasio")
	Amount         float64                `bson:"amount" json:"amount"`                     // Monto por período
	Currency       string                 `bson:"currency" json:"currency"`                 // Moneda (ARS, USD, etc.)
	Frequency      int                    `bson:"frequency" json:"frequency"`               // Frecuencia numérica (1, 2, 3...)
	FrequencyType  string                 `bson:"frequency_type" json:"frequency_type"`     // Tipo (months, days, weeks)
	Status         string                 `bson:"status" json:"status"`                     // Estado: pending, authorized, paused, cancelled
	PaymentGateway string                 `bson:"payment_gateway" json:"payment_gateway"`   // Gateway usado (mercadopago, stripe)
	SubscriptionID string                 `bson:"subscription_id" json:"subscription_id"`   // ID en el gateway externo
	NextPaymentDate *time.Time            `bson:"next_payment_date,omitempty" json:"next_payment_date,omitempty"` // Próxima fecha de cobro
	LastPaymentDate *time.Time            `bson:"last_payment_date,omitempty" json:"last_payment_date,omitempty"` // Última fecha de cobro
	TotalCharges   int                    `bson:"total_charges" json:"total_charges"`       // Cantidad de cobros realizados
	Metadata       map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"` // Datos adicionales
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at" json:"updated_at"`
	CancelledAt    *time.Time             `bson:"cancelled_at,omitempty" json:"cancelled_at,omitempty"` // Fecha de cancelación
}
