package repository

import (
	"context"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionRepository - Interface del repositorio de suscripciones
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *entities.Subscription) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Subscription, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]*entities.Subscription, error)
	FindActiveByUserID(ctx context.Context, userID string) (*entities.Subscription, error)
	FindExpiredSubscriptions(ctx context.Context) ([]*entities.Subscription, error)
	Update(ctx context.Context, id primitive.ObjectID, subscription *entities.Subscription) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status, pagoID string) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}
