package gateways

import (
	"context"
	"fmt"
	"time"
)

// CashGateway - Gateway para pagos en efectivo (manual)
// No requiere integración externa, marca el pago como "pending"
// hasta que un admin lo confirme manualmente en la sucursal
type CashGateway struct{}

// NewCashGateway crea una nueva instancia del gateway de efectivo
func NewCashGateway() *CashGateway {
	return &CashGateway{}
}

// GetName retorna el nombre del gateway
func (c *CashGateway) GetName() string {
	return "cash"
}

// CreatePayment registra un pago en efectivo pendiente de confirmación
func (c *CashGateway) CreatePayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error) {
	// Validaciones básicas
	if request.Amount <= 0 {
		return nil, fmt.Errorf("monto inválido: debe ser mayor a 0")
	}

	if request.CustomerID == "" {
		return nil, fmt.Errorf("usuario requerido")
	}

	// Generar ID de transacción interno único
	transactionID := fmt.Sprintf("CASH-%d-%s", time.Now().Unix(), request.CustomerID)

	// Pago en efectivo siempre queda PENDIENTE hasta confirmación manual
	return &PaymentResult{
		TransactionID: transactionID,
		Status:        StatusPending, // ⚠️ Requiere confirmación manual en sucursal
		PaymentURL:    "",            // No hay URL externa (pago presencial)
		QRCode:        "",            // No hay QR code
		ExternalData: map[string]interface{}{
			"payment_type":          "cash",
			"payment_method":        "efectivo",
			"requires_confirmation": true,
			"confirmation_code":     transactionID,
			"instructions":          fmt.Sprintf("Presentarse en caja con el código: %s", transactionID),
			"valid_until":           time.Now().Add(48 * time.Hour), // Válido 48 horas
		},
		CreatedAt: time.Now(),
		Message:   fmt.Sprintf("Pago en efectivo registrado. Código: %s\n\nAcérquese a la sucursal dentro de las próximas 48 horas para completar el pago.", transactionID),
	}, nil
}

// GetPaymentStatus consulta el estado de un pago en efectivo
// En producción, esto consultaría la base de datos para obtener el estado real
func (c *CashGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transactionID requerido")
	}

	// En una implementación real, aquí consultarías la BD para el estado actual
	// Por ahora retornamos pending (el servicio actualizará el estado desde la BD)
	processedAt := time.Now()
	return &PaymentStatus{
		TransactionID: transactionID,
		Status:        StatusPending,
		StatusDetail:  "awaiting_cash_payment",
		ProcessedAt:   &processedAt,
		ExternalData: map[string]interface{}{
			"payment_method": "efectivo",
			"location":       "sucursal",
		},
	}, nil
}

// RefundPayment procesa un reembolso en efectivo
// También requiere proceso manual en sucursal
func (c *CashGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transactionID requerido")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("monto de reembolso inválido")
	}

	refundID := fmt.Sprintf("REFUND-CASH-%d", time.Now().Unix())

	return &RefundResult{
		RefundID:      refundID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        StatusPending,
		ProcessedAt:   time.Now(),
		Message:       fmt.Sprintf("Reembolso en efectivo registrado. Código: %s\n\nAcérquese a la sucursal para recibir su reembolso.", refundID),
		ExternalData: map[string]interface{}{
			"refund_method": "cash",
			"refund_code":   refundID,
		},
	}, nil
}

// CancelPayment cancela un pago en efectivo pendiente
func (c *CashGateway) CancelPayment(ctx context.Context, transactionID string) error {
	if transactionID == "" {
		return fmt.Errorf("transactionID requerido")
	}

	// Para efectivo, cancelar es simplemente marcar como cancelado
	// El servicio se encargará de actualizar el estado en la BD
	return nil
}

// ProcessWebhook - Los pagos en efectivo no reciben webhooks
func (c *CashGateway) ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookEvent, error) {
	return nil, fmt.Errorf("cash payments do not support webhooks")
}

// ValidateCredentials - Los pagos en efectivo no requieren credenciales externas
func (c *CashGateway) ValidateCredentials(ctx context.Context) error {
	// No requiere validación de credenciales
	return nil
}
