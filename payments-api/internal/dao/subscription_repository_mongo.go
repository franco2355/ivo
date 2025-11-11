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

// SubscriptionRepositoryMongo - Implementación de SubscriptionRepository con MongoDB
type SubscriptionRepositoryMongo struct {
	collection *mongo.Collection
}

// NewSubscriptionRepositoryMongo - Constructor con Dependency Injection
func NewSubscriptionRepositoryMongo(db *mongo.Database) repository.SubscriptionRepository {
	return &SubscriptionRepositoryMongo{
		collection: db.Collection("subscriptions"),
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

func (r *SubscriptionRepositoryMongo) FindByUser(ctx context.Context, userID string) ([]*entities.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("error al buscar suscripciones del usuario: %w", err)
	}
	defer cursor.Close(ctx)

	var subscriptions []*entities.Subscription
	if err := cursor.All(ctx, &subscriptions); err != nil {
		return nil, fmt.Errorf("error al decodificar suscripciones: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepositoryMongo) FindByStatus(ctx context.Context, status string) ([]*entities.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, fmt.Errorf("error al buscar suscripciones por estado: %w", err)
	}
	defer cursor.Close(ctx)

	var subscriptions []*entities.Subscription
	if err := cursor.All(ctx, &subscriptions); err != nil {
		return nil, fmt.Errorf("error al decodificar suscripciones: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepositoryMongo) FindBySubscriptionID(ctx context.Context, subscriptionID string) (*entities.Subscription, error) {
	var subscription entities.Subscription

	err := r.collection.FindOne(ctx, bson.M{"subscription_id": subscriptionID}).Decode(&subscription)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("suscripción no encontrada")
	}
	if err != nil {
		return nil, fmt.Errorf("error al buscar suscripción: %w", err)
	}

	return &subscription, nil
}

func (r *SubscriptionRepositoryMongo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Si el estado es "cancelled", agregar cancelled_at
	if status == "cancelled" {
		now := time.Now()
		update["$set"].(bson.M)["cancelled_at"] = now
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

func (r *SubscriptionRepositoryMongo) UpdateNextPaymentDate(ctx context.Context, id primitive.ObjectID, nextPaymentDate *time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"next_payment_date": nextPaymentDate,
			"updated_at":        time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("error al actualizar fecha de próximo pago: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("suscripción no encontrada")
	}

	return nil
}

func (r *SubscriptionRepositoryMongo) IncrementCharges(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$inc": bson.M{
			"total_charges": 1,
		},
		"$set": bson.M{
			"last_payment_date": time.Now(),
			"updated_at":        time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("error al incrementar cobros: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("suscripción no encontrada")
	}

	return nil
}

func (r *SubscriptionRepositoryMongo) Cancel(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":       "cancelled",
			"cancelled_at": now,
			"updated_at":   now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("error al cancelar suscripción: %w", err)
	}

	if result.MatchedCount == 0 {
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
