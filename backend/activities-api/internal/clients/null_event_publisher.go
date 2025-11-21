package clients

import "log"

// NullEventPublisher - Implementación que no hace nada cuando RabbitMQ no está disponible
type NullEventPublisher struct{}

// NewNullEventPublisher - Constructor
func NewNullEventPublisher() *NullEventPublisher {
	log.Println("⚠️  Usando NullEventPublisher - Los eventos no serán publicados")
	return &NullEventPublisher{}
}

// PublishActivityEvent - Implementa la interface EventPublisher sin hacer nada
func (n *NullEventPublisher) PublishActivityEvent(action, activityID string, data map[string]interface{}) error {
	log.Printf("⚠️  [NullEventPublisher] Evento no publicado: activity.%s (ID: %s)", action, activityID)
	return nil
}

// PublishInscriptionEvent - Implementa la interface EventPublisher sin hacer nada
func (n *NullEventPublisher) PublishInscriptionEvent(action, inscriptionID string, data map[string]interface{}) error {
	log.Printf("⚠️  [NullEventPublisher] Evento no publicado: inscription.%s (ID: %s)", action, inscriptionID)
	return nil
}
