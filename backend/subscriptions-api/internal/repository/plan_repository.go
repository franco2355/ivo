package repository

import (
	"context"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlanRepository - Interface del repositorio de planes (Inversión de Dependencias)
type PlanRepository interface {
	Create(ctx context.Context, plan *entities.Plan) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]*entities.Plan, error)
	FindAllPaginated(ctx context.Context, filters map[string]interface{}, page, pageSize int64, sortBy string, sortDesc bool) ([]*entities.Plan, error)
	Update(ctx context.Context, id primitive.ObjectID, plan *entities.Plan) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}
