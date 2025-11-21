package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(uri, database string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verificar conexión
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("✅ Conectado a MongoDB exitosamente")

	db := client.Database(database)

	mongoDB := &MongoDB{
		Client:   client,
		Database: db,
	}

	// Crear índices
	if err := mongoDB.createIndexes(ctx); err != nil {
		log.Printf("⚠️  Warning: Error creando índices: %v", err)
		// No retornamos error para permitir que continúe la aplicación
	}

	return mongoDB, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// createIndexes - Crea índices para optimizar queries
func (m *MongoDB) createIndexes(ctx context.Context) error {
	log.Println("📊 Creando índices de MongoDB...")

	// Índices para la colección de planes
	planesCollection := m.Database.Collection("planes")
	planIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"activo", 1}},
			Options: options.Index().SetName("idx_planes_activo"),
		},
		{
			Keys: bson.D{{"nombre", 1}},
			Options: options.Index().SetName("idx_planes_nombre"),
		},
		{
			Keys: bson.D{{"precio_mensual", 1}},
			Options: options.Index().SetName("idx_planes_precio"),
		},
		{
			Keys: bson.D{{"created_at", -1}},
			Options: options.Index().SetName("idx_planes_created_at"),
		},
	}

	if _, err := planesCollection.Indexes().CreateMany(ctx, planIndexes); err != nil {
		log.Printf("❌ Error creando índices de planes: %v", err)
		return err
	}
	log.Println("✅ Índices de planes creados")

	// Índices para la colección de suscripciones
	subscriptionCollection := m.Database.Collection("suscripciones")
	subscriptionIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"usuario_id", 1}},
			Options: options.Index().SetName("idx_suscripciones_usuario_id"),
		},
		{
			Keys: bson.D{{"estado", 1}},
			Options: options.Index().SetName("idx_suscripciones_estado"),
		},
		{
			Keys: bson.D{{"fecha_vencimiento", -1}},
			Options: options.Index().SetName("idx_suscripciones_fecha_vencimiento"),
		},
		{
			Keys: bson.D{
				{"usuario_id", 1},
				{"estado", 1},
				{"fecha_vencimiento", -1},
			},
			Options: options.Index().SetName("idx_suscripciones_activa_usuario"),
		},
		{
			Keys: bson.D{{"plan_id", 1}},
			Options: options.Index().SetName("idx_suscripciones_plan_id"),
		},
		{
			Keys: bson.D{{"sucursal_origen_id", 1}},
			Options: options.Index().SetName("idx_suscripciones_sucursal"),
		},
		{
			Keys: bson.D{{"pago_id", 1}},
			Options: options.Index().SetName("idx_suscripciones_pago_id"),
		},
		{
			Keys: bson.D{{"created_at", -1}},
			Options: options.Index().SetName("idx_suscripciones_created_at"),
		},
	}

	if _, err := subscriptionCollection.Indexes().CreateMany(ctx, subscriptionIndexes); err != nil {
		log.Printf("❌ Error creando índices de suscripciones: %v", err)
		return err
	}
	log.Println("✅ Índices de suscripciones creados")

	return nil
}
