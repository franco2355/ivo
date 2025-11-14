package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/payments-api/internal/domain/entities"
	"github.com/yourusername/payments-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PaymentRepositoryMongo - Implementación de PaymentRepository con MongoDB
type PaymentRepositoryMongo struct {
	collection *mongo.Collection
}

// NewPaymentRepositoryMongo - Constructor con Dependency Injection
func NewPaymentRepositoryMongo(db *mongo.Database) repository.PaymentRepository {
	return &PaymentRepositoryMongo{
		collection: db.Collection("payments"),
	}
}

func (r *PaymentRepositoryMongo) Create(ctx context.Context, payment *entities.Payment) error {
	result, err := r.collection.InsertOne(ctx, payment)
	if err != nil {
		return fmt.Errorf("error al crear pago: %w", err)
	}

	payment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PaymentRepositoryMongo) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
	var payment entities.Payment

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("pago no encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("error al buscar pago: %w", err)
	}

	return &payment, nil
}

func (r *PaymentRepositoryMongo) FindAll(ctx context.Context) ([]*entities.Payment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error al buscar todos los pagos: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("error al decodificar pagos: %w", err)
	}

	return payments, nil
}

func (r *PaymentRepositoryMongo) FindByUser(ctx context.Context, userID string) ([]*entities.Payment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("error al buscar pagos del usuario: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("error al decodificar pagos: %w", err)
	}

	return payments, nil
}

func (r *PaymentRepositoryMongo) FindByEntity(ctx context.Context, entityType, entityID string) ([]*entities.Payment, error) {
	filter := bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error al buscar pagos de la entidad: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("error al decodificar pagos: %w", err)
	}

	return payments, nil
}

func (r *PaymentRepositoryMongo) FindByStatus(ctx context.Context, status string) ([]*entities.Payment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, fmt.Errorf("error al buscar pagos por estado: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("error al decodificar pagos: %w", err)
	}

	return payments, nil
}

func (r *PaymentRepositoryMongo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status, transactionID string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Si el estado es "completed", agregar processed_at
	if status == "completed" {
		now := time.Now()
		update["$set"].(bson.M)["processed_at"] = now
	}

	// Si hay transaction_id, agregarlo
	if transactionID != "" {
		update["$set"].(bson.M)["transaction_id"] = transactionID
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("error al actualizar estado: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("pago no encontrado")
	}

	return nil
}

func (r *PaymentRepositoryMongo) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("error al contar pagos: %w", err)
	}

	return count, nil
}
