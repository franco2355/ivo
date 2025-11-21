package mocks

import "context"

// MockUserValidator - Mock para tests
type MockUserValidator struct {
	ValidateUserFunc func(ctx context.Context, userID string) (bool, error)
}

func (m *MockUserValidator) ValidateUser(ctx context.Context, userID string) (bool, error) {
	if m.ValidateUserFunc != nil {
		return m.ValidateUserFunc(ctx, userID)
	}
	return true, nil
}
