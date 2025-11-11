package gateways

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MercadoPagoGateway - Implementación de PaymentGateway para Mercado Pago
// Integra con la API REST de Mercado Pago v1
// Documentación: https://www.mercadopago.com.ar/developers/es/reference
type MercadoPagoGateway struct {
	accessToken   string
	publicKey     string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

// NewMercadoPagoGateway - Constructor con credenciales inyectadas
func NewMercadoPagoGateway(accessToken, publicKey, webhookSecret string) *MercadoPagoGateway {
	return &MercadoPagoGateway{
		accessToken:   accessToken,
		publicKey:     publicKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://api.mercadopago.com", // Producción
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName retorna el identificador del gateway
func (mp *MercadoPagoGateway) GetName() string {
	return "mercadopago"
}

// CreatePayment crea una preferencia de pago en Mercado Pago (Checkout Pro)
// Endpoint: POST /checkout/preferences
// Docs: https://www.mercadopago.com.ar/developers/es/reference/preferences/_checkout_preferences/post
func (mp *MercadoPagoGateway) CreatePayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error) {
	// Construir payload de Checkout Pro
	mpPayload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"title":        request.Description,
				"quantity":     1,
				"unit_price":   request.Amount,
				"currency_id":  request.Currency,
			},
		},
		"payer": map[string]interface{}{
			"email": request.CustomerEmail,
		},
		"back_urls": map[string]interface{}{
			"success": request.CallbackURL + "?status=success",
			"failure": request.CallbackURL + "?status=failure",
			"pending": request.CallbackURL + "?status=pending",
		},
		"auto_return": "approved", // Redirección automática cuando se aprueba
	}

	// Agregar external_reference si existe
	if request.ExternalID != "" {
		mpPayload["external_reference"] = request.ExternalID
	}

	// Agregar notification_url si existe
	if request.WebhookURL != "" {
		mpPayload["notification_url"] = request.WebhookURL
	}

	// Agregar metadata
	if request.Metadata != nil {
		mpPayload["metadata"] = request.Metadata
	}

	// Hacer request HTTP
	respData, err := mp.doRequest(ctx, "POST", "/checkout/preferences", mpPayload)
	if err != nil {
		return nil, fmt.Errorf("error creando preferencia en Mercado Pago: %w", err)
	}

	// Parsear respuesta
	result, err := mp.parsePreferenceResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("error parseando respuesta de Mercado Pago: %w", err)
	}

	return result, nil
}

// GetPaymentStatus consulta el estado de un pago
// Endpoint: GET /v1/payments/{id}
func (mp *MercadoPagoGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	endpoint := fmt.Sprintf("/v1/payments/%s", transactionID)

	respData, err := mp.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error consultando pago en Mercado Pago: %w", err)
	}

	// Parsear respuesta
	status, err := mp.parsePaymentStatusResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("error parseando estado del pago: %w", err)
	}

	return status, nil
}

// RefundPayment procesa un reembolso
// Endpoint: POST /v1/payments/{id}/refunds
func (mp *MercadoPagoGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error) {
	endpoint := fmt.Sprintf("/v1/payments/%s/refunds", transactionID)

	payload := map[string]interface{}{
		"amount": amount,
	}

	respData, err := mp.doRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("error procesando reembolso en Mercado Pago: %w", err)
	}

	// Parsear respuesta
	result, err := mp.parseRefundResponse(respData, transactionID)
	if err != nil {
		return nil, fmt.Errorf("error parseando respuesta de reembolso: %w", err)
	}

	return result, nil
}

// CancelPayment cancela un pago pendiente
// Endpoint: PUT /v1/payments/{id}
func (mp *MercadoPagoGateway) CancelPayment(ctx context.Context, transactionID string) error {
	endpoint := fmt.Sprintf("/v1/payments/%s", transactionID)

	payload := map[string]interface{}{
		"status": "cancelled",
	}

	_, err := mp.doRequest(ctx, "PUT", endpoint, payload)
	if err != nil {
		return fmt.Errorf("error cancelando pago en Mercado Pago: %w", err)
	}

	return nil
}

// ProcessWebhook procesa notificaciones de Mercado Pago
// Docs: https://www.mercadopago.com.ar/developers/es/guides/notifications/webhooks
func (mp *MercadoPagoGateway) ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookEvent, error) {
	// Parsear webhook de Mercado Pago
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("error parseando webhook: %w", err)
	}

	// Extraer información del webhook
	action, _ := webhookData["action"].(string)
	dataObj, _ := webhookData["data"].(map[string]interface{})
	if dataObj == nil {
		return nil, fmt.Errorf("webhook sin campo 'data'")
	}

	// Obtener ID del pago
	paymentIDFloat, _ := dataObj["id"].(float64)
	paymentID := fmt.Sprintf("%.0f", paymentIDFloat)

	if paymentID == "" {
		return nil, fmt.Errorf("webhook sin payment ID")
	}

	// Consultar el estado actual del pago
	status, err := mp.GetPaymentStatus(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("error consultando estado del pago: %w", err)
	}

	// Determinar tipo de evento
	eventType := mp.mapWebhookAction(action, status.Status)

	return &WebhookEvent{
		EventType:     eventType,
		TransactionID: paymentID,
		Status:        status.Status,
		Amount:        status.Amount,
		Currency:      status.Currency,
		ProcessedAt:   time.Now(),
		ExternalData: map[string]interface{}{
			"action":        action,
			"status_detail": status.StatusDetail,
		},
		RawPayload: payload,
	}, nil
}

// ValidateCredentials verifica que las credenciales sean válidas
func (mp *MercadoPagoGateway) ValidateCredentials(ctx context.Context) error {
	// Hacer una petición simple para verificar credenciales
	_, err := mp.doRequest(ctx, "GET", "/v1/payment_methods", nil)
	if err != nil {
		return fmt.Errorf("credenciales inválidas: %w", err)
	}
	return nil
}

// doRequest - Función helper para hacer requests HTTP a Mercado Pago
func (mp *MercadoPagoGateway) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	url := mp.baseURL + endpoint

	var reqBody io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error serializando payload: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	// Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mp.accessToken))

	// Ejecutar request
	resp, err := mp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	// Verificar código de estado
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Mercado Pago error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parsear JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}

	return result, nil
}

// parsePreferenceResponse - Parsea la respuesta de crear preferencia (Checkout Pro)
func (mp *MercadoPagoGateway) parsePreferenceResponse(data map[string]interface{}) (*PaymentResult, error) {
	// Extraer datos de la preferencia
	preferenceID, _ := data["id"].(string)
	initPoint, _ := data["init_point"].(string)
	sandboxInitPoint, _ := data["sandbox_init_point"].(string)

	// Determinar qué URL usar (sandbox para testing, init_point para producción)
	paymentURL := initPoint
	if sandboxInitPoint != "" && mp.accessToken[:5] == "TEST-" {
		paymentURL = sandboxInitPoint
	}

	return &PaymentResult{
		TransactionID: preferenceID, // El ID de la preferencia (no del pago aún)
		Status:        StatusPending, // Siempre pending hasta que el usuario pague
		PaymentURL:    paymentURL,    // URL para redirigir al usuario
		QRCode:        "",            // Checkout Pro no usa QR
		ExternalData: map[string]interface{}{
			"preference_id":      preferenceID,
			"init_point":         initPoint,
			"sandbox_init_point": sandboxInitPoint,
		},
		CreatedAt: time.Now(),
		Message:   "Preferencia creada. Redirigir al usuario a payment_url",
	}, nil
}

// parsePaymentStatusResponse - Parsea la respuesta de consulta de estado
func (mp *MercadoPagoGateway) parsePaymentStatusResponse(data map[string]interface{}) (*PaymentStatus, error) {
	id, _ := data["id"].(float64)
	status, _ := data["status"].(string)
	statusDetail, _ := data["status_detail"].(string)
	amount, _ := data["transaction_amount"].(float64)
	currency, _ := data["currency_id"].(string)
	paymentMethod, _ := data["payment_method_id"].(string)

	// Fecha de procesamiento
	var processedAt *time.Time
	if dateApproved, ok := data["date_approved"].(string); ok && dateApproved != "" {
		if t, err := time.Parse(time.RFC3339, dateApproved); err == nil {
			processedAt = &t
		}
	}

	return &PaymentStatus{
		TransactionID: fmt.Sprintf("%.0f", id),
		Status:        mp.mapStatus(status),
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: paymentMethod,
		ProcessedAt:   processedAt,
		StatusDetail:  statusDetail,
		ExternalData: map[string]interface{}{
			"raw_status": status,
		},
	}, nil
}

// parseRefundResponse - Parsea la respuesta de reembolso
func (mp *MercadoPagoGateway) parseRefundResponse(data map[string]interface{}, transactionID string) (*RefundResult, error) {
	id, _ := data["id"].(float64)
	amount, _ := data["amount"].(float64)
	status, _ := data["status"].(string)

	return &RefundResult{
		RefundID:      fmt.Sprintf("%.0f", id),
		TransactionID: transactionID,
		Amount:        amount,
		Status:        mp.mapStatus(status),
		ProcessedAt:   time.Now(),
		Message:       "Reembolso procesado en Mercado Pago",
	}, nil
}

// mapStatus - Mapea estados de Mercado Pago a estados genéricos
func (mp *MercadoPagoGateway) mapStatus(mpStatus string) string {
	switch mpStatus {
	case "approved":
		return StatusCompleted
	case "pending", "in_process", "in_mediation":
		return StatusPending
	case "rejected", "cancelled":
		return StatusFailed
	case "refunded", "charged_back":
		return StatusRefunded
	default:
		return StatusPending
	}
}

// mapWebhookAction - Mapea acciones de webhook a tipos de evento genéricos
func (mp *MercadoPagoGateway) mapWebhookAction(action, status string) string {
	switch action {
	case "payment.created":
		return EventPaymentCreated
	case "payment.updated":
		if status == StatusCompleted {
			return EventPaymentUpdated
		} else if status == StatusFailed {
			return EventPaymentFailed
		}
		return EventPaymentUpdated
	default:
		return EventPaymentUpdated
	}
}
