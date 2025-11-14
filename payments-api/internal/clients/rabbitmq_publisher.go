package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher - Cliente para publicar eventos de pagos
type RabbitMQPublisher struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	exchange string
}

// PaymentEvent - Estructura del evento de pago
type PaymentEvent struct {
	Action         string                 `json:"action"`          // "payment.created", "payment.completed", "payment.failed", "payment.refunded"
	Type           string                 `json:"type"`            // "payment"
	PaymentID      string                 `json:"payment_id"`      // ID del pago en payments-api
	Status         string                 `json:"status"`          // Estado del pago: "pending", "completed", "failed", "refunded"
	EntityType     string                 `json:"entity_type"`     // "subscription", "inscription"
	EntityID       string                 `json:"entity_id"`       // ID de la entidad (subscription_id, inscription_id)
	UserID         string                 `json:"user_id"`         // ID del usuario que pagó
	Amount         float64                `json:"amount"`          // Monto pagado
	Currency       string                 `json:"currency"`        // ARS, USD, etc.
	TransactionID  string                 `json:"transaction_id"`  // ID de la transacción en el gateway
	PaymentGateway string                 `json:"payment_gateway"` // mercadopago, stripe, etc.
	Timestamp      time.Time              `json:"timestamp"`       // Fecha del evento
	Metadata       map[string]interface{} `json:"metadata"`        // Información adicional
}

// NewRabbitMQPublisher - Constructor con Dependency Injection
func NewRabbitMQPublisher(url, exchange string) (*RabbitMQPublisher, error) {
	// 1. Conectar a RabbitMQ
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("error conectando a RabbitMQ: %w", err)
	}

	// 2. Crear canal
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error creando canal: %w", err)
	}

	// 3. Declarar exchange (Topic Exchange para routing flexible)
	err = channel.ExchangeDeclare(
		exchange, // name: "gym.events"
		"topic",  // type: permite routing con patterns (payment.*, subscription.*)
		true,     // durable: sobrevive a reinicios
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("error declarando exchange: %w", err)
	}

	log.Printf("✅ Conectado a RabbitMQ (Exchange: %s)", exchange)

	return &RabbitMQPublisher{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
	}, nil
}

// PublishPaymentEvent - Publica un evento de pago
func (r *RabbitMQPublisher) PublishPaymentEvent(ctx context.Context, event PaymentEvent) error {
	// 1. Serializar evento a JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	// 2. Construir routing key
	// Ejemplos:
	//   - "payment.completed.subscription"
	//   - "payment.failed.inscription"
	//   - "payment.refunded.subscription"
	routingKey := fmt.Sprintf("payment.%s.%s", getActionName(event.Action), event.EntityType)

	// 3. Publicar mensaje
	err = r.channel.PublishWithContext(
		ctx,
		r.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory (no requerir confirmación de queue existente)
		false,      // immediate (no requerir consumidor inmediato)
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent, // Mensaje persistente (sobrevive a reinicio de RabbitMQ)
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		log.Printf("❌ Error publicando evento a RabbitMQ: %v", err)
		return fmt.Errorf("error publicando evento: %w", err)
	}

	log.Printf("📤 Evento publicado: %s (PaymentID: %s, UserID: %s, Amount: %.2f)",
		routingKey, event.PaymentID, event.UserID, event.Amount)

	return nil
}

// Close - Cierra la conexión
func (r *RabbitMQPublisher) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Helper para extraer nombre de acción sin prefijo "payment."
func getActionName(action string) string {
	// "payment.completed" -> "completed"
	if len(action) > 8 && action[:8] == "payment." {
		return action[8:]
	}
	return action
}
