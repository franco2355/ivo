package repository

import (
	"context"

	"github.com/yourusername/payments-api/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentRepository - Interface para operaciones de pagos
// Abstracción para permitir diferentes implementaciones (MongoDB, PostgreSQL, etc.)
type PaymentRepository interface {
	// Create crea un nuevo pago
	Create(ctx context.Context, payment *entities.Payment) error

	// FindByID busca un pago por su ID
	FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error)

	// FindAll busca todos los pagos
	FindAll(ctx context.Context) ([]*entities.Payment, error)

	// FindByUser busca todos los pagos de un usuario
	FindByUser(ctx context.Context, userID string) ([]*entities.Payment, error)

	// FindByEntity busca pagos asociados a una entidad específica
	FindByEntity(ctx context.Context, entityType, entityID string) ([]*entities.Payment, error)

	// FindByStatus busca pagos por estado
	FindByStatus(ctx context.Context, status string) ([]*entities.Payment, error)

	// UpdateStatus actualiza el estado de un pago
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status, transactionID string) error

	// Count cuenta los pagos que cumplen con los filtros
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}
