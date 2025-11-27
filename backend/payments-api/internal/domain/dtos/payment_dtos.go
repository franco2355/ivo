package dtos

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatePaymentRequest - DTO para crear un nuevo pago
type CreatePaymentRequest struct {
	EntityType     string                 `json:"entity_type" binding:"required"`
	EntityID       string                 `json:"entity_id" binding:"required"`
	UserID         string                 `json:"user_id" binding:"required"`
	Amount         float64                `json:"amount" binding:"required,gt=0"`
	Currency       string                 `json:"currency" binding:"required"`
	PaymentMethod  string                 `json:"payment_method" binding:"required"`
	PaymentGateway string                 `json:"payment_gateway,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"` // UUID para prevenir pagos duplicados
	CallbackURL    string                 `json:"callback_url,omitempty"`
	WebhookURL     string                 `json:"webhook_url,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePaymentStatusRequest - DTO para actualizar el estado de un pago
type UpdatePaymentStatusRequest struct {
	Status        string `json:"status" binding:"required,oneof=pending completed failed refunded"`
	TransactionID string `json:"transaction_id,omitempty"`
}

// PaymentResponse - DTO de respuesta con información del pago
type PaymentResponse struct {
	ID             string                 `json:"id"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	UserID         string                 `json:"user_id"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Status         string                 `json:"status"`
	PaymentMethod  string                 `json:"payment_method"`
	PaymentGateway string                 `json:"payment_gateway,omitempty"`
	TransactionID  string                 `json:"transaction_id,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ProcessedAt    *time.Time             `json:"processed_at,omitempty"`
}

// ToPaymentResponse - Convierte una entidad Payment a PaymentResponse
func ToPaymentResponse(id primitive.ObjectID, entityType, entityID, userID string, amount float64, currency, status, paymentMethod, paymentGateway, transactionID, idempotencyKey string, metadata map[string]interface{}, createdAt, updatedAt time.Time, processedAt *time.Time) PaymentResponse {
	return PaymentResponse{
		ID:             id.Hex(),
		EntityType:     entityType,
		EntityID:       entityID,
		UserID:         userID,
		Amount:         amount,
		Currency:       currency,
		Status:         status,
		PaymentMethod:  paymentMethod,
		PaymentGateway: paymentGateway,
		TransactionID:  transactionID,
		IdempotencyKey: idempotencyKey,
		Metadata:       metadata,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		ProcessedAt:    processedAt,
	}
}
