package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Renovacion representa una renovación de suscripción
type Renovacion struct {
	Fecha  time.Time `bson:"fecha"`
	PagoID string    `bson:"pago_id"`
	Monto  float64   `bson:"monto"`
}

// Metadata representa metadatos adicionales de suscripción
type Metadata struct {
	AutoRenovacion      bool   `bson:"auto_renovacion"`
	MetodoPagoPreferido string `bson:"metodo_pago_preferido"`
	Notas               string `bson:"notas"`
}

// Subscription representa una suscripción de usuario (Entidad de Dominio)
type Subscription struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"`
	UsuarioID             string             `bson:"usuario_id"`
	PlanID                primitive.ObjectID `bson:"plan_id"`
	SucursalOrigenID      string             `bson:"sucursal_origen_id,omitempty"`
	FechaInicio           time.Time          `bson:"fecha_inicio"`
	FechaVencimiento      time.Time          `bson:"fecha_vencimiento"`
	Estado                string             `bson:"estado"` // "activa" | "vencida" | "cancelada" | "pendiente_pago"
	PagoID                string             `bson:"pago_id,omitempty"`
	Metadata              Metadata           `bson:"metadata"`
	HistorialRenovaciones []Renovacion       `bson:"historial_renovaciones"`
	CreatedAt             time.Time          `bson:"created_at"`
	UpdatedAt             time.Time          `bson:"updated_at"`
}
