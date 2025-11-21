package dtos

import "time"

// CreateSubscriptionRequest - DTO para crear una suscripción
type CreateSubscriptionRequest struct {
	UsuarioID        string `json:"usuario_id" binding:"required"`
	PlanID           string `json:"plan_id" binding:"required"`
	SucursalOrigenID string `json:"sucursal_origen_id"`
	MetodoPago       string `json:"metodo_pago" binding:"required"`
	AutoRenovacion   bool   `json:"auto_renovacion"`
	Notas            string `json:"notas"`
}

// UpdateSubscriptionStatusRequest - DTO para actualizar estado
type UpdateSubscriptionStatusRequest struct {
	Estado string `json:"estado" binding:"required,oneof=activa vencida cancelada pendiente_pago"`
	PagoID string `json:"pago_id"`
}

// RenovacionResponse - DTO para historial de renovaciones
type RenovacionResponse struct {
	Fecha  time.Time `json:"fecha"`
	PagoID string    `json:"pago_id"`
	Monto  float64   `json:"monto"`
}

// SubscriptionResponse - DTO para respuesta de una suscripción
type SubscriptionResponse struct {
	ID                    string               `json:"id"`
	UsuarioID             string               `json:"usuario_id"`
	PlanID                string               `json:"plan_id"`
	PlanNombre            string               `json:"plan_nombre,omitempty"`      // Enriquecido
	SucursalOrigenID      string               `json:"sucursal_origen_id,omitempty"`
	FechaInicio           time.Time            `json:"fecha_inicio"`
	FechaVencimiento      time.Time            `json:"fecha_vencimiento"`
	Estado                string               `json:"estado"`
	PagoID                string               `json:"pago_id,omitempty"`
	AutoRenovacion        bool                 `json:"auto_renovacion"`
	MetodoPagoPreferido   string               `json:"metodo_pago_preferido"`
	Notas                 string               `json:"notas,omitempty"`
	HistorialRenovaciones []RenovacionResponse `json:"historial_renovaciones"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
}

// ListSubscriptionsQuery - DTO para query params de listado
type ListSubscriptionsQuery struct {
	UsuarioID string `form:"usuario_id"`
	Estado    string `form:"estado" binding:"omitempty,oneof=activa vencida cancelada pendiente_pago"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
