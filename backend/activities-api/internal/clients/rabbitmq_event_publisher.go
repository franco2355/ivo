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
	Conn     *amqp.Connection // Exportado para health checks
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
		Conn:     conn,
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

// PublishActivityEvent - Publica eventos de actividades
func (r *RabbitMQEventPublisher) PublishActivityEvent(action, activityID string, data map[string]interface{}) error {
	event := rabbitMQEvent{
		Action:    action,
		Type:      "activity",
		ID:        activityID,
		Timestamp: time.Now(),
		Data:      data,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	// Routing key: activity.{action}
	routingKey := fmt.Sprintf("activity.%s", action)

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

	log.Printf("📤 Evento publicado: %s (ID: %s)\n", routingKey, activityID)
	return nil
}

// PublishInscriptionEvent - Publica eventos de inscripciones
func (r *RabbitMQEventPublisher) PublishInscriptionEvent(action, inscriptionID string, data map[string]interface{}) error {
	event := rabbitMQEvent{
		Action:    action,
		Type:      "inscription",
		ID:        inscriptionID,
		Timestamp: time.Now(),
		Data:      data,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	// Routing key: inscription.{action}
	routingKey := fmt.Sprintf("inscription.%s", action)

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

	log.Printf("📤 Evento publicado: %s (ID: %s)\n", routingKey, inscriptionID)
	return nil
}

// Close - Cierra la conexión
func (r *RabbitMQEventPublisher) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.Conn != nil {
		return r.Conn.Close()
	}
	return nil
}
