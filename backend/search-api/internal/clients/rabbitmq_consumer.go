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
		// Cuando cambia una inscripción, actualizar cupo de la actividad afectada
		err = r.handleInscriptionEvent(event)
		if err != nil {
			log.Printf("❌ Error procesando evento de inscripción: %v\n", err)
			msg.Nack(false, true) // Requeue
			return
		}
		// Flush TODO el caché porque cambió el cupo de una actividad
		r.cacheService.FlushAll()
		log.Printf("✅ Actividad actualizada por cambio en inscripción\n")

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
		// Eventos no manejados específicamente - ignorar
		log.Printf("⏭️  Evento ignorado: %s.%s (ID: %s)\n", event.Type, event.Action, event.ID)
	}

	// Invalidar caché relacionado
	r.cacheService.InvalidatePattern(event.Type)

	msg.Ack(false)
}

// handleActivityEvent procesa eventos de actividad obteniendo datos frescos desde MySQL
func (r *RabbitMQConsumer) handleActivityEvent(event dtos.RabbitMQEvent) error {
	switch event.Action {
	case "create", "update":
		// Obtener datos frescos desde MySQL (más robusto que parsear el evento)
		log.Printf("📝 Reindexando actividad %s desde MySQL (action: %s)\n", event.ID, event.Action)
		return r.searchService.ReindexActivityByID(event.ID)

	case "delete":
		log.Printf("🗑️  Eliminando actividad %s del índice\n", event.ID)
		return r.searchService.DeleteDocument(event.ID)

	default:
		log.Printf("⚠️  Acción desconocida para actividad: %s\n", event.Action)
		return nil
	}
}

// handleInscriptionEvent procesa eventos de inscripción actualizando solo el cupo (partial update)
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

	log.Printf("⚡ Partial update: actualizando cupo_disponible de actividad %s\n", actividadIDStr)

	// Solo actualizar cupo_disponible (partial update optimizado)
	return r.searchService.UpdateActivityCupo(actividadIDStr)
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
