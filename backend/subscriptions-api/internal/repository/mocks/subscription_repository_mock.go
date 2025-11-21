package mocks

import (
	"context"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockSubscriptionRepository - Mock para tests
type MockSubscriptionRepository struct {
	CreateFunc              func(ctx context.Context, subscription *entities.Subscription) error
	FindByIDFunc            func(ctx context.Context, id primitive.ObjectID) (*entities.Subscription, error)
	FindAllFunc             func(ctx context.Context, filters map[string]interface{}) ([]*entities.Subscription, error)
	FindActiveByUserIDFunc  func(ctx context.Context, userID string) (*entities.Subscription, error)
	UpdateFunc              func(ctx context.Context, id primitive.ObjectID, subscription *entities.Subscription) error
	UpdateStatusFunc        func(ctx context.Context, id primitive.ObjectID, status, pagoID string) error
	DeleteFunc              func(ctx context.Context, id primitive.ObjectID) error
	CountFunc               func(ctx context.Context, filters map[string]interface{}) (int64, error)
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, subscription)
	}
	return nil
}

func (m *MockSubscriptionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Subscription, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSubscriptionRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]*entities.Subscription, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, filters)
	}
	return []*entities.Subscription{}, nil
}

func (m *MockSubscriptionRepository) FindActiveByUserID(ctx context.Context, userID string) (*entities.Subscription, error) {
	if m.FindActiveByUserIDFunc != nil {
		return m.FindActiveByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, id primitive.ObjectID, subscription *entities.Subscription) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, subscription)
	}
	return nil
}

func (m *MockSubscriptionRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status, pagoID string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status, pagoID)
	}
	return nil
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSubscriptionRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, filters)
	}
	return 0, nil
}
