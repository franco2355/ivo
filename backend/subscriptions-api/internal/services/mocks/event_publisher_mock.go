package mocks

// MockEventPublisher - Mock para tests
type MockEventPublisher struct {
	PublishSubscriptionEventFunc func(action, subscriptionID string, data map[string]interface{}) error
}

func (m *MockEventPublisher) PublishSubscriptionEvent(action, subscriptionID string, data map[string]interface{}) error {
	if m.PublishSubscriptionEventFunc != nil {
		return m.PublishSubscriptionEventFunc(action, subscriptionID, data)
	}
	return nil
}
