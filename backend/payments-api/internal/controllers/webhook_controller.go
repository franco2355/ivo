package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/payments-api/internal/gateways"
	"github.com/yourusername/payments-api/internal/repository"
	"github.com/yourusername/payments-api/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookController - Controlador para recibir notificaciones de gateways de pago
// Los webhooks son llamadas HTTP asíncronas que envían los gateways cuando cambia el estado de un pago
type WebhookController struct {
	gatewayFactory *gateways.GatewayFactory
	paymentRepo    repository.PaymentRepository
	paymentService *services.PaymentService // Para obtener el payment completo y publicar eventos
}

// NewWebhookController - Constructor con DI
func NewWebhookController(
	gatewayFactory *gateways.GatewayFactory,
	paymentRepo repository.PaymentRepository,
	paymentService *services.PaymentService,
) *WebhookController {
	return &WebhookController{
		gatewayFactory: gatewayFactory,
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// HandleMercadoPagoWebhook - Procesa webhooks de Mercado Pago
// POST /webhooks/mercadopago
//
// Mercado Pago envía notificaciones como:
// {
//   "action": "payment.updated",
//   "type": "payment",
//   "data": {
//     "id": "123456789"
//   }
// }
func (wc *WebhookController) HandleMercadoPagoWebhook(c *gin.Context) {
	log.Println("📩 Webhook recibido de Mercado Pago")

	// 1. Leer el payload completo
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Error leyendo webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error leyendo payload"})
		return
	}

	// 2. Extraer headers (algunos gateways envían firmas de seguridad)
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 3. Obtener el gateway de Mercado Pago
	gateway, err := wc.gatewayFactory.CreateGateway("mercadopago")
	if err != nil {
		log.Printf("❌ Error creando gateway: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}

	// 4. Procesar el webhook usando el gateway
	webhookEvent, err := gateway.ProcessWebhook(context.Background(), bodyBytes, headers)
	if err != nil {
		log.Printf("❌ Error procesando webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error procesando webhook"})
		return
	}

	log.Printf("✅ Webhook procesado: TransactionID=%s, Status=%s, EventType=%s",
		webhookEvent.TransactionID, webhookEvent.Status, webhookEvent.EventType)

	// 5. Buscar el pago en nuestra BD por transaction_id
	payment, err := wc.findPaymentByTransactionID(webhookEvent.TransactionID)
	if err != nil {
		log.Printf("⚠️  Pago no encontrado en BD: %s", webhookEvent.TransactionID)
		// Aún así retornamos 200 OK para que el gateway no reintente
		c.JSON(http.StatusOK, gin.H{"status": "payment not found, but webhook received"})
		return
	}

	// 6. Actualizar el estado del pago si cambió (Y PUBLICAR EVENTOS ⭐)
	if payment.Status != webhookEvent.Status {
		err = wc.paymentService.UpdatePaymentStatusFromWebhook(
			context.Background(),
			payment.ID,
			webhookEvent.Status,
			webhookEvent.TransactionID,
		)
		if err != nil {
			log.Printf("❌ Error actualizando estado del pago: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando pago"})
			return
		}

		log.Printf("✅ Pago actualizado vía webhook: ID=%s, OldStatus=%s, NewStatus=%s",
			payment.ID.Hex(), payment.Status, webhookEvent.Status)
		log.Printf("📤 Evento publicado a RabbitMQ (si está configurado)")
	}

	// 7. Responder 200 OK (importante para que el gateway no reintente)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Webhook procesado correctamente",
	})
}

// HandleStripeWebhook - Procesa webhooks de Stripe (futuro)
// POST /webhooks/stripe
func (wc *WebhookController) HandleStripeWebhook(c *gin.Context) {
	log.Println("📩 Webhook recibido de Stripe")

	// TODO: Implementar cuando se agregue Stripe
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Stripe webhooks aún no implementados",
	})
}

// HandleGenericWebhook - Procesa webhooks genéricos
// POST /webhooks/:gateway
// Ejemplo: POST /webhooks/mercadopago, POST /webhooks/stripe
func (wc *WebhookController) HandleGenericWebhook(c *gin.Context) {
	gatewayName := c.Param("gateway")
	log.Printf("📩 Webhook recibido de gateway: %s", gatewayName)

	// 1. Validar que el gateway sea soportado
	if !wc.gatewayFactory.ValidateGatewayName(gatewayName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Gateway no soportado: %s", gatewayName),
		})
		return
	}

	// 2. Leer payload
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error leyendo payload"})
		return
	}

	// 3. Extraer headers
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 4. Obtener gateway
	gateway, err := wc.gatewayFactory.CreateGateway(gatewayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando gateway"})
		return
	}

	// 5. Procesar webhook
	webhookEvent, err := gateway.ProcessWebhook(context.Background(), bodyBytes, headers)
	if err != nil {
		log.Printf("❌ Error procesando webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error procesando webhook"})
		return
	}

	// 6. Buscar y actualizar pago (Y PUBLICAR EVENTOS ⭐)
	payment, err := wc.findPaymentByTransactionID(webhookEvent.TransactionID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "payment not found"})
		return
	}

	if payment.Status != webhookEvent.Status {
		err = wc.paymentService.UpdatePaymentStatusFromWebhook(
			context.Background(),
			payment.ID,
			webhookEvent.Status,
			webhookEvent.TransactionID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando pago"})
			return
		}

		log.Printf("✅ Pago actualizado vía webhook genérico: ID=%s, Gateway=%s, Status=%s",
			payment.ID.Hex(), gatewayName, webhookEvent.Status)
		log.Printf("📤 Evento publicado a RabbitMQ (si está configurado)")
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Webhook procesado correctamente",
	})
}

// findPaymentByTransactionID - Helper para buscar un pago por transaction_id
// Como el repository no tiene este método, tenemos que buscar por status y filtrar
// En producción deberías agregar un método FindByTransactionID al repository
func (wc *WebhookController) findPaymentByTransactionID(transactionID string) (*struct {
	ID            primitive.ObjectID
	Status        string
	TransactionID string
}, error) {
	// TODO: Optimizar esto agregando un índice en transaction_id y un método específico
	// Por ahora, buscamos todos los pagos pending y completed
	ctx := context.Background()

	// Buscar en pending
	pendingPayments, err := wc.paymentRepo.FindByStatus(ctx, "pending")
	if err == nil {
		for _, p := range pendingPayments {
			if p.TransactionID == transactionID {
				return &struct {
					ID            primitive.ObjectID
					Status        string
					TransactionID string
				}{
					ID:            p.ID,
					Status:        p.Status,
					TransactionID: p.TransactionID,
				}, nil
			}
		}
	}

	// Buscar en completed
	completedPayments, err := wc.paymentRepo.FindByStatus(ctx, "completed")
	if err == nil {
		for _, p := range completedPayments {
			if p.TransactionID == transactionID {
				return &struct {
					ID            primitive.ObjectID
					Status        string
					TransactionID string
				}{
					ID:            p.ID,
					Status:        p.Status,
					TransactionID: p.TransactionID,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("payment not found")
}
