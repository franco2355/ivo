package services

// EventPublisher define la interfaz para publicar eventos
// Permite dependency injection y facilita testing
type EventPublisher interface {
	PublishActivityEvent(action, activityID string, data map[string]interface{}) error
	PublishInscriptionEvent(action, inscriptionID string, data map[string]interface{}) error
}
