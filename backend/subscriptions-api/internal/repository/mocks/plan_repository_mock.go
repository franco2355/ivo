package mocks

import (
	"context"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockPlanRepository - Mock para tests
type MockPlanRepository struct {
	CreateFunc            func(ctx context.Context, plan *entities.Plan) error
	FindByIDFunc          func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error)
	FindAllFunc           func(ctx context.Context, filters map[string]interface{}) ([]*entities.Plan, error)
	FindAllPaginatedFunc  func(ctx context.Context, filters map[string]interface{}, page, pageSize int64, sortBy string, sortDesc bool) ([]*entities.Plan, error)
	UpdateFunc            func(ctx context.Context, id primitive.ObjectID, plan *entities.Plan) error
	DeleteFunc            func(ctx context.Context, id primitive.ObjectID) error
	CountFunc             func(ctx context.Context, filters map[string]interface{}) (int64, error)
}

func (m *MockPlanRepository) Create(ctx context.Context, plan *entities.Plan) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, plan)
	}
	return nil
}

func (m *MockPlanRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPlanRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]*entities.Plan, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, filters)
	}
	return []*entities.Plan{}, nil
}

func (m *MockPlanRepository) FindAllPaginated(ctx context.Context, filters map[string]interface{}, page, pageSize int64, sortBy string, sortDesc bool) ([]*entities.Plan, error) {
	if m.FindAllPaginatedFunc != nil {
		return m.FindAllPaginatedFunc(ctx, filters, page, pageSize, sortBy, sortDesc)
	}
	return []*entities.Plan{}, nil
}

func (m *MockPlanRepository) Update(ctx context.Context, id primitive.ObjectID, plan *entities.Plan) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, plan)
	}
	return nil
}

func (m *MockPlanRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockPlanRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, filters)
	}
	return 0, nil
}
