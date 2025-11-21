package clients

import "log"

// NullEventPublisher - Implementación que no hace nada cuando RabbitMQ no está disponible
type NullEventPublisher struct{}

// NewNullEventPublisher - Constructor
func NewNullEventPublisher() *NullEventPublisher {
	log.Println("⚠️  Usando NullEventPublisher - Los eventos no serán publicados")
	return &NullEventPublisher{}
}

// PublishSubscriptionEvent - Implementa la interface EventPublisher sin hacer nada
func (n *NullEventPublisher) PublishSubscriptionEvent(action, subscriptionID string, data map[string]interface{}) error {
	log.Printf("⚠️  [NullEventPublisher] Evento no publicado: subscription.%s (ID: %s)", action, subscriptionID)
	return nil
}
