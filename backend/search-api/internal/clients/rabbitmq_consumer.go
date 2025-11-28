package clients

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"github.com/yourusername/gym-management/search-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/search-api/internal/services"
)

// RabbitMQConsumer - Cliente de RabbitMQ para consumir eventos
type RabbitMQConsumer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	searchService *services.SearchService
	cacheService  *services.CacheService
}

// NewRabbitMQConsumer - Constructor con DI
func NewRabbitMQConsumer(url, exchange, queueName string, searchService *services.SearchService, cacheService *services.CacheService) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
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
		return nil, err
	}

	// Declarar cola
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
		return nil, err
	}

	// Bind a todos los eventos relevantes
	bindings := []string{
		"activity.*",
		"plan.*",
		"subscription.*",
		"inscription.*",
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
	}

	log.Printf("✅ Conectado a RabbitMQ como consumidor (Queue: %s)\n", queueName)

	return &RabbitMQConsumer{
		conn:          conn,
		channel:       channel,
		searchService: searchService,
		cacheService:  cacheService,
	}, nil
}

// Start inicia el consumo de mensajes
func (r *RabbitMQConsumer) Start() error {
	msgs, err := r.channel.Consume(
		"search_indexer_queue", // queue
		"",                     // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return err
	}

	log.Println("🎧 Escuchando eventos de RabbitMQ...")

	go func() {
		for msg := range msgs {
			r.handleMessage(msg)
		}
	}()

	return nil
}

func (r *RabbitMQConsumer) handleMessage(msg amqp.Delivery) {
	var event dtos.RabbitMQEvent

	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		log.Printf("❌ Error decodificando evento: %v\n", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("📥 Evento recibido: %s.%s (ID: %s)\n", event.Type, event.Action, event.ID)

	// Procesar según el tipo de evento
	switch event.Type {
	case "inscription":
		// Cuando cambia una inscripción, reindexar la actividad afectada
		err = r.handleInscriptionEvent(event)
		if err != nil {
			log.Printf("❌ Error procesando evento de inscripción: %v\n", err)
			msg.Nack(false, true) // Requeue
			return
		}
		log.Printf("✅ Actividad reindexada por cambio en inscripción\n")

	case "activity":
		// Para actividades, reindexar desde MySQL para obtener todos los campos
		err = r.handleActivityEvent(event)
		if err != nil {
			log.Printf("❌ Error procesando evento de actividad: %v\n", err)
			msg.Nack(false, true) // Requeue
			return
		}
		log.Printf("✅ Actividad procesada: %s (action: %s)\n", event.ID, event.Action)

	default:
		// Para otros eventos (plan, subscription), procesamiento normal
		switch event.Action {
		case "create", "update":
			err = r.searchService.IndexFromEvent(event)
			if err != nil {
				log.Printf("❌ Error indexando documento: %v\n", err)
				msg.Nack(false, true) // Requeue
				return
			}
			log.Printf("✅ Documento indexado: %s_%s\n", event.Type, event.ID)

		case "delete":
			docID := event.ID // Usar solo el ID numérico para consistencia
			err = r.searchService.DeleteDocument(docID)
			if err != nil {
				log.Printf("❌ Error eliminando documento: %v\n", err)
				msg.Nack(false, true)
				return
			}
			log.Printf("🗑️  Documento eliminado: %s\n", docID)
		}
	}

	// Invalidar caché relacionado
	r.cacheService.InvalidatePattern(event.Type)

	msg.Ack(false)
}

// handleActivityEvent procesa eventos de actividad indexando directamente desde el evento
func (r *RabbitMQConsumer) handleActivityEvent(event dtos.RabbitMQEvent) error {
	switch event.Action {
	case "create", "update":
		// Indexar directamente desde el evento (viene con todos los campos)
		log.Printf("📝 Indexando actividad %s desde evento (action: %s)\n", event.ID, event.Action)
		return r.searchService.IndexFromEvent(event)

	case "delete":
		log.Printf("🗑️  Eliminando actividad %s del índice\n", event.ID)
		return r.searchService.DeleteDocument(event.ID)

	default:
		log.Printf("⚠️  Acción desconocida para actividad: %s\n", event.Action)
		return nil
	}
}

// handleInscriptionEvent procesa eventos de inscripción reindexando la actividad afectada
func (r *RabbitMQConsumer) handleInscriptionEvent(event dtos.RabbitMQEvent) error {
	// Extraer actividad_id del evento
	actividadID, ok := event.Data["actividad_id"]
	if !ok {
		log.Printf("⚠️  Evento de inscripción sin actividad_id\n")
		return nil
	}

	// Convertir a string
	actividadIDStr := ""
	switch v := actividadID.(type) {
	case float64:
		actividadIDStr = fmt.Sprintf("%.0f", v)
	case int:
		actividadIDStr = fmt.Sprintf("%d", v)
	case string:
		actividadIDStr = v
	default:
		log.Printf("⚠️  Tipo de actividad_id no soportado: %T\n", v)
		return nil
	}

	log.Printf("🔄 Reindexando actividad %s por cambio en inscripción\n", actividadIDStr)

	// Reindexar la actividad desde MySQL para obtener cupo actualizado
	return r.searchService.ReindexActivityByID(actividadIDStr)
}

// Close cierra las conexiones
func (r *RabbitMQConsumer) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
