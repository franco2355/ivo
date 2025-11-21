package repository

import (
	"context"
	"time"

	"github.com/yourusername/payments-api/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionRepository - Interface para operaciones de suscripciones
// Abstracción para permitir diferentes implementaciones (MongoDB, PostgreSQL, etc.)
type SubscriptionRepository interface {
	// Create crea una nueva suscripción
	Create(ctx context.Context, subscription *entities.Subscription) error

	// FindByID busca una suscripción por su ID
	FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Subscription, error)

	// FindByUser busca todas las suscripciones de un usuario
	FindByUser(ctx context.Context, userID string) ([]*entities.Subscription, error)

	// FindByStatus busca suscripciones por estado
	FindByStatus(ctx context.Context, status string) ([]*entities.Subscription, error)

	// FindBySubscriptionID busca una suscripción por su ID en el gateway externo
	FindBySubscriptionID(ctx context.Context, subscriptionID string) (*entities.Subscription, error)

	// UpdateStatus actualiza el estado de una suscripción
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error

	// UpdateNextPaymentDate actualiza la próxima fecha de cobro
	UpdateNextPaymentDate(ctx context.Context, id primitive.ObjectID, nextPaymentDate *time.Time) error

	// IncrementCharges incrementa el contador de cobros realizados
	IncrementCharges(ctx context.Context, id primitive.ObjectID) error

	// Cancel marca una suscripción como cancelada
	Cancel(ctx context.Context, id primitive.ObjectID) error

	// Count cuenta las suscripciones que cumplen con los filtros
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}
