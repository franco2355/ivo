package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

// SubscriptionEventHandler define la interfaz para manejar eventos de suscripciones
type SubscriptionEventHandler interface {
	HandleSubscriptionCancelled(ctx context.Context, usuarioID uint) error
}

// RabbitMQSubscriptionConsumer consume eventos de suscripciones desde RabbitMQ
type RabbitMQSubscriptionConsumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queue    string
	handler  SubscriptionEventHandler
	done     chan bool
}

// SubscriptionEvent representa un evento de suscripción
type SubscriptionEvent struct {
	Action    string                 `json:"action"`
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewRabbitMQSubscriptionConsumer crea un nuevo consumer de eventos de suscripciones
func NewRabbitMQSubscriptionConsumer(url, exchange string, handler SubscriptionEventHandler) (*RabbitMQSubscriptionConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("error conectando a RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error creando canal: %w", err)
	}

	// Declarar exchange
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
		return nil, fmt.Errorf("error declarando exchange: %w", err)
	}

	// Declarar queue para eventos de suscripciones
	queueName := "activities_subscription_events"
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("error declarando queue: %w", err)
	}

	// Bind queue a eventos de suscripción cancelada
	routingKey := "subscription.cancelled"
	err = channel.QueueBind(
		queue.Name, // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("error bindeando queue: %w", err)
	}

	log.Printf("✅ RabbitMQ Subscription Consumer conectado (Exchange: %s, Queue: %s, Routing: %s)\n", exchange, queueName, routingKey)

	return &RabbitMQSubscriptionConsumer{
		conn:    conn,
		channel: channel,
		queue:   queue.Name,
		handler: handler,
		done:    make(chan bool),
	}, nil
}

// Start inicia el consumo de mensajes
func (c *RabbitMQSubscriptionConsumer) Start() error {
	msgs, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer
		false,   // auto-ack (false para control manual)
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("error consumiendo mensajes: %w", err)
	}

	log.Println("🎧 Activities-API escuchando eventos de suscripciones canceladas...")

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Println("❌ Canal de mensajes cerrado")
					return
				}
				c.processMessage(msg)
			case <-c.done:
				log.Println("🛑 Deteniendo consumer de suscripciones")
				return
			}
		}
	}()

	return nil
}

// processMessage procesa un mensaje individual
func (c *RabbitMQSubscriptionConsumer) processMessage(msg amqp.Delivery) {
	log.Printf("📩 [SubscriptionConsumer] Mensaje recibido: %s\n", string(msg.Body))

	var event SubscriptionEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("❌ [SubscriptionConsumer] Error parseando mensaje: %v\n", err)
		msg.Nack(false, false) // No requeue
		return
	}

	// Verificar que sea un evento de cancelación
	if event.Action != "cancelled" {
		log.Printf("⚠️ [SubscriptionConsumer] Evento ignorado (action: %s)\n", event.Action)
		msg.Ack(false)
		return
	}

	// Obtener usuario_id del evento
	usuarioIDRaw, ok := event.Data["usuario_id"]
	if !ok {
		log.Printf("❌ [SubscriptionConsumer] Evento sin usuario_id\n")
		msg.Nack(false, false)
		return
	}

	// Convertir usuario_id a uint
	var usuarioID uint
	switch v := usuarioIDRaw.(type) {
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			log.Printf("❌ [SubscriptionConsumer] Error parseando usuario_id: %v\n", err)
			msg.Nack(false, false)
			return
		}
		usuarioID = uint(id)
	case float64:
		usuarioID = uint(v)
	default:
		log.Printf("❌ [SubscriptionConsumer] Tipo de usuario_id no soportado: %T\n", v)
		msg.Nack(false, false)
		return
	}

	log.Printf("🔄 [SubscriptionConsumer] Procesando cancelación de suscripción para usuario %d\n", usuarioID)

	// Llamar al handler para desinscribir al usuario
	ctx := context.Background()
	if err := c.handler.HandleSubscriptionCancelled(ctx, usuarioID); err != nil {
		log.Printf("❌ [SubscriptionConsumer] Error manejando evento: %v\n", err)
		msg.Nack(false, true) // Requeue para reintentar
		return
	}

	log.Printf("✅ [SubscriptionConsumer] Evento procesado exitosamente para usuario %d\n", usuarioID)
	msg.Ack(false)
}

// Stop detiene el consumer
func (c *RabbitMQSubscriptionConsumer) Stop() {
	c.done <- true
}

// Close cierra la conexión
func (c *RabbitMQSubscriptionConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
