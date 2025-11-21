package dao

import (
	"context"
	"fmt"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"github.com/yourusername/gym-management/subscriptions-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PlanRepositoryMongo - Implementación de PlanRepository con MongoDB
type PlanRepositoryMongo struct {
	collection *mongo.Collection
}

// NewPlanRepositoryMongo - Constructor con Dependency Injection
func NewPlanRepositoryMongo(db *mongo.Database) repository.PlanRepository {
	return &PlanRepositoryMongo{
		collection: db.Collection("planes"),
	}
}

func (r *PlanRepositoryMongo) Create(ctx context.Context, plan *entities.Plan) error {
	result, err := r.collection.InsertOne(ctx, plan)
	if err != nil {
		return fmt.Errorf("error al crear plan: %w", err)
	}

	plan.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PlanRepositoryMongo) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
	var plan entities.Plan

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&plan)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("plan no encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("error al buscar plan: %w", err)
	}

	return &plan, nil
}

func (r *PlanRepositoryMongo) FindAll(ctx context.Context, filters map[string]interface{}) ([]*entities.Plan, error) {
	cursor, err := r.collection.Find(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("error al listar planes: %w", err)
	}
	defer cursor.Close(ctx)

	var plans []*entities.Plan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, fmt.Errorf("error al decodificar planes: %w", err)
	}

	return plans, nil
}

// FindAllPaginated - Busca planes con paginación real
func (r *PlanRepositoryMongo) FindAllPaginated(ctx context.Context, filters map[string]interface{}, page, pageSize int64, sortBy string, sortDesc bool) ([]*entities.Plan, error) {
	// Configurar opciones de paginación
	opts := options.Find()

	// Skip: saltar los registros de las páginas anteriores
	skip := (page - 1) * pageSize
	opts.SetSkip(skip)

	// Limit: limitar cantidad de resultados
	opts.SetLimit(pageSize)

	// Ordenamiento
	if sortBy != "" {
		sortOrder := 1 // Ascendente
		if sortDesc {
			sortOrder = -1 // Descendente
		}
		opts.SetSort(bson.D{{Key: sortBy, Value: sortOrder}})
	}

	// Ejecutar query
	cursor, err := r.collection.Find(ctx, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("error al listar planes paginados: %w", err)
	}
	defer cursor.Close(ctx)

	var plans []*entities.Plan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, fmt.Errorf("error al decodificar planes: %w", err)
	}

	return plans, nil
}

func (r *PlanRepositoryMongo) Update(ctx context.Context, id primitive.ObjectID, plan *entities.Plan) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": plan}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error al actualizar plan: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("plan no encontrado")
	}

	return nil
}

func (r *PlanRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("error al eliminar plan: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("plan no encontrado")
	}

	return nil
}

func (r *PlanRepositoryMongo) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("error al contar planes: %w", err)
	}

	return count, nil
}
