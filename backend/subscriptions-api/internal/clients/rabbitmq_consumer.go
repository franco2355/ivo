package clients

import (
	"context"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/handlers"
)

// RabbitMQConsumer - Cliente de RabbitMQ para consumir eventos de pagos
type RabbitMQConsumer struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	queueName       string
	paymentHandler  *handlers.PaymentEventHandler
	stopChan        chan bool
}

// NewRabbitMQConsumer - Constructor con DI
func NewRabbitMQConsumer(
	url string,
	exchange string,
	queueName string,
	paymentHandler *handlers.PaymentEventHandler,
) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declarar exchange (topic para routing por patrones)
	err = channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declarar cola para eventos de pagos relacionados con suscripciones
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable (sobrevive a reinicios)
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Bind a eventos de pagos relacionados con suscripciones
	// Routing keys: payment.{action}.subscription
	bindings := []string{
		"payment.created.subscription",   // Cuando se crea un pago para una suscripción
		"payment.completed.subscription", // Cuando se completa un pago → ACTIVAR SUSCRIPCIÓN
		"payment.failed.subscription",    // Cuando falla un pago
		"payment.refunded.subscription",  // Cuando se reembolsa un pago → DESACTIVAR/CANCELAR
	}

	for _, binding := range bindings {
		err = channel.QueueBind(
			queue.Name, // queue name
			binding,    // routing key
			exchange,   // exchange
			false,
			nil,
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, err
		}
		log.Printf("✅ Queue '%s' vinculada a evento: %s\n", queueName, binding)
	}

	log.Printf("✅ RabbitMQ Consumer conectado (Exchange: %s, Queue: %s)\n", exchange, queueName)

	return &RabbitMQConsumer{
		conn:           conn,
		channel:        channel,
		queueName:      queueName,
		paymentHandler: paymentHandler,
		stopChan:       make(chan bool),
	}, nil
}

// Start inicia el consumo de mensajes en background
func (r *RabbitMQConsumer) Start() error {
	msgs, err := r.channel.Consume(
		r.queueName, // queue
		"",          // consumer (auto-generated)
		false,       // auto-ack (manual ack para mayor control)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return err
	}

	log.Println("🎧 Subscriptions-API escuchando eventos de pagos...")

	// Goroutine para procesar mensajes
	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Println("⚠️  Canal de mensajes cerrado")
					return
				}
				r.handleMessage(msg)

			case <-r.stopChan:
				log.Println("🛑 Deteniendo consumer de RabbitMQ...")
				return
			}
		}
	}()

	return nil
}

// handleMessage procesa cada mensaje recibido
func (r *RabbitMQConsumer) handleMessage(msg amqp.Delivery) {
	var event dtos.PaymentEvent

	// Decodificar el evento
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		log.Printf("❌ Error decodificando evento de pago: %v\n", err)
		msg.Nack(false, false) // No requeue si está mal formado
		return
	}

	log.Printf("📥 Evento de pago recibido: %s | Subscription: %s | Status: %s\n",
		event.Action, event.EntityID, event.Status)

	// Validar que sea un evento de suscripción
	if !event.IsSubscriptionEvent() {
		log.Printf("⚠️  Evento no es de suscripción (EntityType: %s), ignorando\n", event.EntityType)
		msg.Ack(false)
		return
	}

	// Procesar el evento según la acción
	ctx := context.Background()
	var processErr error

	switch event.Action {
	case "payment.created":
		// Pago creado - podríamos registrar el intento o actualizar metadata
		log.Printf("📝 Pago creado para suscripción %s (PaymentID: %s)\n", event.EntityID, event.PaymentID)
		// No hacemos nada crítico aquí, solo logging
		processErr = nil

	case "payment.completed":
		// ✅ PAGO COMPLETADO → ACTIVAR SUSCRIPCIÓN
		log.Printf("✅ Pago completado para suscripción %s - Activando...\n", event.EntityID)
		processErr = r.paymentHandler.HandlePaymentCompleted(ctx, event)

	case "payment.failed":
		// ❌ PAGO FALLIDO → Mantener en pendiente_pago, opcionalmente notificar
		log.Printf("❌ Pago fallido para suscripción %s\n", event.EntityID)
		processErr = r.paymentHandler.HandlePaymentFailed(ctx, event)

	case "payment.refunded":
		// 💰 PAGO REEMBOLSADO → CANCELAR/DESACTIVAR SUSCRIPCIÓN
		log.Printf("💰 Pago reembolsado para suscripción %s - Cancelando...\n", event.EntityID)
		processErr = r.paymentHandler.HandlePaymentRefunded(ctx, event)

	default:
		log.Printf("⚠️  Acción desconocida: %s\n", event.Action)
		processErr = nil // Ignorar acciones desconocidas
	}

	// Manejar el resultado del procesamiento
	if processErr != nil {
		log.Printf("❌ Error procesando evento %s: %v\n", event.Action, processErr)
		// Requeue para reintento (con límite implícito de RabbitMQ)
		msg.Nack(false, true)
		return
	}

	// Confirmar procesamiento exitoso
	msg.Ack(false)
	log.Printf("✅ Evento %s procesado correctamente\n", event.Action)
}

// Stop detiene el consumer gracefully
func (r *RabbitMQConsumer) Stop() {
	close(r.stopChan)
}

// Close cierra las conexiones
func (r *RabbitMQConsumer) Close() error {
	r.Stop()

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
