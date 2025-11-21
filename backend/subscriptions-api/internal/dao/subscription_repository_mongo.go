package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"github.com/yourusername/gym-management/subscriptions-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SubscriptionRepositoryMongo - Implementación con MongoDB
type SubscriptionRepositoryMongo struct {
	collection *mongo.Collection
}

// NewSubscriptionRepositoryMongo - Constructor con DI
func NewSubscriptionRepositoryMongo(db *mongo.Database) repository.SubscriptionRepository {
	return &SubscriptionRepositoryMongo{
		collection: db.Collection("suscripciones"),
	}
}

func (r *SubscriptionRepositoryMongo) Create(ctx context.Context, subscription *entities.Subscription) error {
	result, err := r.collection.InsertOne(ctx, subscription)
	if err != nil {
		return fmt.Errorf("error al crear suscripción: %w", err)
	}

	subscription.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *SubscriptionRepositoryMongo) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Subscription, error) {
	var subscription entities.Subscription

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&subscription)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("suscripción no encontrada")
	}
	if err != nil {
		return nil, fmt.Errorf("error al buscar suscripción: %w", err)
	}

	return &subscription, nil
}

func (r *SubscriptionRepositoryMongo) FindAll(ctx context.Context, filters map[string]interface{}) ([]*entities.Subscription, error) {
	cursor, err := r.collection.Find(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("error al listar suscripciones: %w", err)
	}
	defer cursor.Close(ctx)

	var subscriptions []*entities.Subscription
	if err := cursor.All(ctx, &subscriptions); err != nil {
		return nil, fmt.Errorf("error al decodificar suscripciones: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepositoryMongo) FindActiveByUserID(ctx context.Context, userID string) (*entities.Subscription, error) {
	filter := bson.M{
		"usuario_id":        userID,
		"estado":            "activa",
		"fecha_vencimiento": bson.M{"$gt": time.Now()},
	}

	var subscription entities.Subscription
	err := r.collection.FindOne(ctx, filter).Decode(&subscription)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("no hay suscripción activa")
	}
	if err != nil {
		return nil, fmt.Errorf("error al buscar suscripción activa: %w", err)
	}

	return &subscription, nil
}

func (r *SubscriptionRepositoryMongo) Update(ctx context.Context, id primitive.ObjectID, subscription *entities.Subscription) error {
	subscription.UpdatedAt = time.Now()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": subscription}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error al actualizar suscripción: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("suscripción no encontrada")
	}

	return nil
}

func (r *SubscriptionRepositoryMongo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status, pagoID string) error {
	update := bson.M{
		"$set": bson.M{
			"estado":     status,
			"pago_id":    pagoID,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("error al actualizar estado: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("suscripción no encontrada")
	}

	return nil
}

func (r *SubscriptionRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("error al eliminar suscripción: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("suscripción no encontrada")
	}

	return nil
}

func (r *SubscriptionRepositoryMongo) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("error al contar suscripciones: %w", err)
	}

	return count, nil
}
