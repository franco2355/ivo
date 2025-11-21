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

// MercadoPagoSubscriptionGateway - Implementación de SubscriptionGateway para Mercado Pago
// Integra con la API de Preapprovals (suscripciones recurrentes)
// Documentación: https://www.mercadopago.com.ar/developers/es/reference/subscriptions/_preapproval/post
type MercadoPagoSubscriptionGateway struct {
	accessToken   string
	publicKey     string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

// NewMercadoPagoSubscriptionGateway - Constructor
func NewMercadoPagoSubscriptionGateway(accessToken, publicKey, webhookSecret string) *MercadoPagoSubscriptionGateway {
	return &MercadoPagoSubscriptionGateway{
		accessToken:   accessToken,
		publicKey:     publicKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://api.mercadopago.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName retorna el identificador del gateway
func (mp *MercadoPagoSubscriptionGateway) GetName() string {
	return "mercadopago_subscriptions"
}

// CreateSubscription crea una suscripción recurrente en Mercado Pago
// Endpoint: POST /preapproval
func (mp *MercadoPagoSubscriptionGateway) CreateSubscription(ctx context.Context, request SubscriptionRequest) (*SubscriptionResult, error) {
	// Construir payload de Preapproval
	mpPayload := map[string]interface{}{
		"reason": request.Reason,
		"auto_recurring": map[string]interface{}{
			"frequency":           request.Frequency,
			"frequency_type":      request.FrequencyType,
			"transaction_amount":  request.Amount,
			"currency_id":         request.Currency,
		},
		"payer_email": request.CustomerEmail,
		"back_url":    request.CallbackURL,
		"status":      "pending", // Siempre inicia como pending
	}

	// Agregar external_reference si existe
	if request.ExternalID != "" {
		mpPayload["external_reference"] = request.ExternalID
	}

	// Agregar notification_url si existe
	if request.WebhookURL != "" {
		mpPayload["notification_url"] = request.WebhookURL
	}

	// Agregar fecha de inicio si existe
	if request.StartDate != nil {
		mpPayload["start_date"] = request.StartDate.Format(time.RFC3339)
	}

	// Agregar fecha de fin si existe
	if request.EndDate != nil {
		mpPayload["end_date"] = request.EndDate.Format(time.RFC3339)
	}

	// Hacer request HTTP
	respData, err := mp.doRequest(ctx, "POST", "/preapproval", mpPayload)
	if err != nil {
		return nil, fmt.Errorf("error creando suscripción en Mercado Pago: %w", err)
	}

	// Parsear respuesta
	result, err := mp.parseSubscriptionResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("error parseando respuesta de Mercado Pago: %w", err)
	}

	return result, nil
}

// GetSubscriptionStatus consulta el estado de una suscripción
// Endpoint: GET /preapproval/{id}
func (mp *MercadoPagoSubscriptionGateway) GetSubscriptionStatus(ctx context.Context, subscriptionID string) (*SubscriptionStatus, error) {
	endpoint := fmt.Sprintf("/preapproval/%s", subscriptionID)

	respData, err := mp.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error consultando suscripción en Mercado Pago: %w", err)
	}

	// Parsear respuesta
	status, err := mp.parseSubscriptionStatusResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("error parseando estado de suscripción: %w", err)
	}

	return status, nil
}

// CancelSubscription cancela una suscripción
// Endpoint: PUT /preapproval/{id}
func (mp *MercadoPagoSubscriptionGateway) CancelSubscription(ctx context.Context, subscriptionID string) error {
	endpoint := fmt.Sprintf("/preapproval/%s", subscriptionID)

	payload := map[string]interface{}{
		"status": "cancelled",
	}

	_, err := mp.doRequest(ctx, "PUT", endpoint, payload)
	if err != nil {
		return fmt.Errorf("error cancelando suscripción en Mercado Pago: %w", err)
	}

	return nil
}

// PauseSubscription pausa una suscripción
// Endpoint: PUT /preapproval/{id}
func (mp *MercadoPagoSubscriptionGateway) PauseSubscription(ctx context.Context, subscriptionID string) error {
	endpoint := fmt.Sprintf("/preapproval/%s", subscriptionID)

	payload := map[string]interface{}{
		"status": "paused",
	}

	_, err := mp.doRequest(ctx, "PUT", endpoint, payload)
	if err != nil {
		return fmt.Errorf("error pausando suscripción en Mercado Pago: %w", err)
	}

	return nil
}

// ResumeSubscription reanuda una suscripción pausada
// Endpoint: PUT /preapproval/{id}
func (mp *MercadoPagoSubscriptionGateway) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	endpoint := fmt.Sprintf("/preapproval/%s", subscriptionID)

	payload := map[string]interface{}{
		"status": "authorized",
	}

	_, err := mp.doRequest(ctx, "PUT", endpoint, payload)
	if err != nil {
		return fmt.Errorf("error reanudando suscripción en Mercado Pago: %w", err)
	}

	return nil
}

// ProcessSubscriptionWebhook procesa notificaciones de suscripciones
func (mp *MercadoPagoSubscriptionGateway) ProcessSubscriptionWebhook(ctx context.Context, payload []byte, headers map[string]string) (*SubscriptionWebhookEvent, error) {
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

	// Obtener ID de la suscripción
	preapprovalID, _ := dataObj["id"].(string)
	if preapprovalID == "" {
		return nil, fmt.Errorf("webhook sin preapproval ID")
	}

	// Consultar el estado actual de la suscripción
	status, err := mp.GetSubscriptionStatus(ctx, preapprovalID)
	if err != nil {
		return nil, fmt.Errorf("error consultando estado de suscripción: %w", err)
	}

	// Determinar tipo de evento
	eventType := mp.mapWebhookAction(action, status.Status)

	return &SubscriptionWebhookEvent{
		EventType:      eventType,
		SubscriptionID: preapprovalID,
		PaymentID:      "", // TODO: Extraer payment ID si es un evento de cobro
		Status:         status.Status,
		Amount:         status.Amount,
		Currency:       status.Currency,
		ProcessedAt:    time.Now(),
		ExternalData: map[string]interface{}{
			"action": action,
		},
		RawPayload: payload,
	}, nil
}

// ValidateCredentials verifica que las credenciales sean válidas
func (mp *MercadoPagoSubscriptionGateway) ValidateCredentials(ctx context.Context) error {
	// Hacer una petición simple para verificar credenciales
	_, err := mp.doRequest(ctx, "GET", "/v1/payment_methods", nil)
	if err != nil {
		return fmt.Errorf("credenciales inválidas: %w", err)
	}
	return nil
}

// doRequest - Helper para hacer requests HTTP a Mercado Pago
func (mp *MercadoPagoSubscriptionGateway) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (map[string]interface{}, error) {
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

// parseSubscriptionResponse - Parsea la respuesta de crear suscripción
func (mp *MercadoPagoSubscriptionGateway) parseSubscriptionResponse(data map[string]interface{}) (*SubscriptionResult, error) {
	id, _ := data["id"].(string)
	status, _ := data["status"].(string)
	initPoint, _ := data["init_point"].(string)
	sandboxInitPoint, _ := data["sandbox_init_point"].(string)

	// Determinar qué URL usar
	url := initPoint
	if sandboxInitPoint != "" && mp.accessToken[:5] == "TEST-" {
		url = sandboxInitPoint
	}

	return &SubscriptionResult{
		SubscriptionID: id,
		Status:         mp.mapStatus(status),
		InitPoint:      url,
		ExternalData: map[string]interface{}{
			"raw_status":         status,
			"init_point":         initPoint,
			"sandbox_init_point": sandboxInitPoint,
		},
		CreatedAt: time.Now(),
		Message:   "Suscripción creada. Redirigir al usuario para autorizar",
	}, nil
}

// parseSubscriptionStatusResponse - Parsea la respuesta de consulta de estado
func (mp *MercadoPagoSubscriptionGateway) parseSubscriptionStatusResponse(data map[string]interface{}) (*SubscriptionStatus, error) {
	id, _ := data["id"].(string)
	status, _ := data["status"].(string)
	reason, _ := data["reason"].(string)

	// Extraer auto_recurring
	autoRecurring, _ := data["auto_recurring"].(map[string]interface{})
	amount, _ := autoRecurring["transaction_amount"].(float64)
	currency, _ := autoRecurring["currency_id"].(string)
	frequency, _ := autoRecurring["frequency"].(float64)
	frequencyType, _ := autoRecurring["frequency_type"].(string)

	// Fecha del próximo pago
	var nextPaymentDate *time.Time
	if nextDate, ok := data["next_payment_date"].(string); ok && nextDate != "" {
		if t, err := time.Parse(time.RFC3339, nextDate); err == nil {
			nextPaymentDate = &t
		}
	}

	// Fecha del último pago
	var lastPaymentDate *time.Time
	if lastDate, ok := data["last_modified"].(string); ok && lastDate != "" {
		if t, err := time.Parse(time.RFC3339, lastDate); err == nil {
			lastPaymentDate = &t
		}
	}

	return &SubscriptionStatus{
		SubscriptionID:  id,
		Status:          mp.mapStatus(status),
		Reason:          reason,
		Amount:          amount,
		Currency:        currency,
		Frequency:       int(frequency),
		FrequencyType:   frequencyType,
		NextPaymentDate: nextPaymentDate,
		LastPaymentDate: lastPaymentDate,
		TotalCharges:    0, // MP no devuelve esto directamente
		ExternalData: map[string]interface{}{
			"raw_status": status,
		},
	}, nil
}

// mapStatus - Mapea estados de Mercado Pago a estados genéricos
func (mp *MercadoPagoSubscriptionGateway) mapStatus(mpStatus string) string {
	switch mpStatus {
	case "pending":
		return SubscriptionStatusPending
	case "authorized":
		return SubscriptionStatusAuthorized
	case "paused":
		return SubscriptionStatusPaused
	case "cancelled":
		return SubscriptionStatusCancelled
	default:
		return SubscriptionStatusPending
	}
}

// mapWebhookAction - Mapea acciones de webhook a tipos de evento
func (mp *MercadoPagoSubscriptionGateway) mapWebhookAction(action, status string) string {
	switch action {
	case "created":
		return EventSubscriptionCreated
	case "updated":
		if status == SubscriptionStatusAuthorized {
			return EventSubscriptionAuthorized
		} else if status == SubscriptionStatusCancelled {
			return EventSubscriptionCancelled
		} else if status == SubscriptionStatusPaused {
			return EventSubscriptionPaused
		}
		return EventSubscriptionCreated
	default:
		return EventSubscriptionCreated
	}
}
