package gateways

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// MockGateway - Implementación simulada de PaymentGateway para testing
// No hace llamadas HTTP reales, simula comportamiento de una pasarela real
// Útil para desarrollo y testing sin credenciales
type MockGateway struct {
	// Puedes agregar configuración si necesitas simular diferentes comportamientos
	shouldFail bool // Para simular errores
}

// NewMockGateway - Constructor del mock gateway
func NewMockGateway() *MockGateway {
	return &MockGateway{
		shouldFail: false,
	}
}

// GetName retorna el identificador del gateway
func (m *MockGateway) GetName() string {
	return "mock"
}

// CreatePayment simula la creación de un pago
func (m *MockGateway) CreatePayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error) {
	// Simular validaciones
	if request.Amount <= 0 {
		return nil, fmt.Errorf("monto inválido")
	}

	if request.Currency == "" {
		return nil, fmt.Errorf("moneda requerida")
	}

	// Simular delay de red
	time.Sleep(100 * time.Millisecond)

	// Simular error si está configurado
	if m.shouldFail {
		return &PaymentResult{
			TransactionID: "",
			Status:        StatusFailed,
			Message:       "Pago rechazado por el gateway (simulado)",
			CreatedAt:     time.Now(),
		}, nil
	}

	// Generar ID de transacción simulado
	transactionID := fmt.Sprintf("MOCK-%d", time.Now().Unix())

	// Retornar resultado exitoso
	return &PaymentResult{
		TransactionID: transactionID,
		Status:        StatusCompleted, // Mock siempre aprueba
		PaymentURL:    fmt.Sprintf("https://mock-gateway.com/payment/%s", transactionID),
		QRCode:        "", // Mock no genera QR
		ExternalData: map[string]interface{}{
			"mock":        true,
			"test_mode":   true,
			"external_id": request.ExternalID,
		},
		CreatedAt: time.Now(),
		Message:   "Pago procesado exitosamente (simulado)",
	}, nil
}

// GetPaymentStatus simula la consulta de estado de un pago
func (m *MockGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	// Simular delay de red
	time.Sleep(50 * time.Millisecond)

	// Validar que el transactionID tenga formato correcto
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID requerido")
	}

	// Retornar estado simulado
	now := time.Now()
	return &PaymentStatus{
		TransactionID: transactionID,
		Status:        StatusCompleted,
		Amount:        100.00, // Monto simulado
		Currency:      "ARS",
		PaymentMethod: "credit_card",
		ProcessedAt:   &now,
		StatusDetail:  "accredited",
		ExternalData: map[string]interface{}{
			"mock": true,
		},
	}, nil
}

// RefundPayment simula el procesamiento de un reembolso
func (m *MockGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error) {
	// Simular delay de red
	time.Sleep(100 * time.Millisecond)

	// Validaciones
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID requerido")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("monto de reembolso inválido")
	}

	// Generar ID de reembolso simulado
	refundID := fmt.Sprintf("REFUND-MOCK-%d", time.Now().Unix())

	return &RefundResult{
		RefundID:      refundID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        StatusCompleted,
		ProcessedAt:   time.Now(),
		ExternalData: map[string]interface{}{
			"mock": true,
		},
		Message: "Reembolso procesado exitosamente (simulado)",
	}, nil
}

// CancelPayment simula la cancelación de un pago pendiente
func (m *MockGateway) CancelPayment(ctx context.Context, transactionID string) error {
	// Simular delay de red
	time.Sleep(50 * time.Millisecond)

	if transactionID == "" {
		return fmt.Errorf("transaction ID requerido")
	}

	// Mock siempre permite cancelar
	return nil
}

// ProcessWebhook simula el procesamiento de un webhook
func (m *MockGateway) ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookEvent, error) {
	// Intentar parsear el payload como JSON genérico
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("payload inválido: %w", err)
	}

	// Extraer información básica (estructura simulada)
	transactionID, _ := data["transaction_id"].(string)
	if transactionID == "" {
		transactionID = "MOCK-WEBHOOK-123"
	}

	return &WebhookEvent{
		EventType:     EventPaymentUpdated,
		TransactionID: transactionID,
		Status:        StatusCompleted,
		Amount:        100.00,
		Currency:      "ARS",
		ProcessedAt:   time.Now(),
		ExternalData: map[string]interface{}{
			"mock": true,
		},
		RawPayload: payload,
	}, nil
}

// ValidateCredentials simula la validación de credenciales
func (m *MockGateway) ValidateCredentials(ctx context.Context) error {
	// Mock siempre retorna credenciales válidas
	return nil
}

// SetShouldFail - Método para testing: forzar errores
func (m *MockGateway) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}
