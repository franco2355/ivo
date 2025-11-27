package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Plan representa un plan de suscripción (Entidad de Dominio)
type Plan struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"`
	Nombre                string             `bson:"nombre"`
	Descripcion           string             `bson:"descripcion"`
	PrecioMensual         float64            `bson:"precio_mensual"`
	TipoAcceso            string             `bson:"tipo_acceso"` // "limitado" | "completo"
	DuracionDias          int                `bson:"duracion_dias"`
	Activo                bool               `bson:"activo"`
	ActividadesPermitidas []string           `bson:"actividades_permitidas"`
	ActividadesPorSemana  int                `bson:"actividades_por_semana"` // Límite de actividades por semana (0 = ilimitado)
	CreatedAt             time.Time          `bson:"created_at"`
	UpdatedAt             time.Time          `bson:"updated_at"`
}
