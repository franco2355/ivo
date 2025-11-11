package dtos

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateSubscriptionRequest - DTO para crear una nueva suscripción
type CreateSubscriptionRequest struct {
	UserID         string                 `json:"user_id" binding:"required"`
	Reason         string                 `json:"reason" binding:"required"`                // "Cuota mensual gimnasio"
	Amount         float64                `json:"amount" binding:"required,gt=0"`           // Monto por período
	Currency       string                 `json:"currency" binding:"required"`              // ARS, USD, etc.
	Frequency      int                    `json:"frequency" binding:"required,gt=0"`        // 1, 2, 3...
	FrequencyType  string                 `json:"frequency_type" binding:"required"`        // months, days, weeks
	PaymentGateway string                 `json:"payment_gateway" binding:"required"`       // mercadopago, stripe
	CallbackURL    string                 `json:"callback_url,omitempty"`                   // URL de retorno después de autorizar
	StartDate      *time.Time             `json:"start_date,omitempty"`                     // Fecha de inicio (opcional)
	EndDate        *time.Time             `json:"end_date,omitempty"`                       // Fecha de fin (opcional)
	Metadata       map[string]interface{} `json:"metadata,omitempty"`                       // Datos adicionales
}

// SubscriptionResponse - DTO para responder con datos de una suscripción
type SubscriptionResponse struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Reason          string                 `json:"reason"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Frequency       int                    `json:"frequency"`
	FrequencyType   string                 `json:"frequency_type"`
	Status          string                 `json:"status"`
	PaymentGateway  string                 `json:"payment_gateway"`
	SubscriptionID  string                 `json:"subscription_id"`
	InitPoint       string                 `json:"init_point,omitempty"`         // URL para autorizar (solo al crear)
	NextPaymentDate *time.Time             `json:"next_payment_date,omitempty"`
	LastPaymentDate *time.Time             `json:"last_payment_date,omitempty"`
	TotalCharges    int                    `json:"total_charges"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CancelledAt     *time.Time             `json:"cancelled_at,omitempty"`
}

// ToSubscriptionResponse - Convierte entity a DTO de respuesta
func ToSubscriptionResponse(
	id primitive.ObjectID,
	userID string,
	reason string,
	amount float64,
	currency string,
	frequency int,
	frequencyType string,
	status string,
	paymentGateway string,
	subscriptionID string,
	initPoint string,
	nextPaymentDate *time.Time,
	lastPaymentDate *time.Time,
	totalCharges int,
	metadata map[string]interface{},
	createdAt time.Time,
	updatedAt time.Time,
	cancelledAt *time.Time,
) SubscriptionResponse {
	return SubscriptionResponse{
		ID:              id.Hex(),
		UserID:          userID,
		Reason:          reason,
		Amount:          amount,
		Currency:        currency,
		Frequency:       frequency,
		FrequencyType:   frequencyType,
		Status:          status,
		PaymentGateway:  paymentGateway,
		SubscriptionID:  subscriptionID,
		InitPoint:       initPoint,
		NextPaymentDate: nextPaymentDate,
		LastPaymentDate: lastPaymentDate,
		TotalCharges:    totalCharges,
		Metadata:        metadata,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		CancelledAt:     cancelledAt,
	}
}
