package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQEventPublisher - Implementación de EventPublisher con RabbitMQ
type RabbitMQEventPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

// NewRabbitMQEventPublisher - Constructor con DI
func NewRabbitMQEventPublisher(url, exchange string) (*RabbitMQEventPublisher, error) {
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

	log.Printf("✅ Conectado a RabbitMQ (Exchange: %s)\n", exchange)

	return &RabbitMQEventPublisher{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
	}, nil
}

type rabbitMQEvent struct {
	Action    string                 `json:"action"`
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// PublishSubscriptionEvent - Implementa la interface EventPublisher
func (r *RabbitMQEventPublisher) PublishSubscriptionEvent(action, subscriptionID string, data map[string]interface{}) error {
	event := rabbitMQEvent{
		Action:    action,
		Type:      "subscription",
		ID:        subscriptionID,
		Timestamp: time.Now(),
		Data:      data,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	// Routing key: subscription.{action}
	routingKey := fmt.Sprintf("subscription.%s", action)

	err = r.channel.Publish(
		r.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("❌ Error publicando evento: %v\n", err)
		return fmt.Errorf("error publicando evento: %w", err)
	}

	log.Printf("📤 Evento publicado: %s (ID: %s)\n", routingKey, subscriptionID)
	return nil
}

// Close - Cierra la conexión
func (r *RabbitMQEventPublisher) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
