package dtos

import "time"

// CreatePlanRequest - DTO para crear un plan
type CreatePlanRequest struct {
	Nombre                string   `json:"nombre" binding:"required,min=3,max=100"`
	Descripcion           string   `json:"descripcion" binding:"max=500"`
	PrecioMensual         float64  `json:"precio_mensual" binding:"required,gt=0"`
	TipoAcceso            string   `json:"tipo_acceso" binding:"required,oneof=limitado completo"`
	DuracionDias          int      `json:"duracion_dias" binding:"required,gt=0"`
	Activo                bool     `json:"activo"`
	ActividadesPermitidas []string `json:"actividades_permitidas"`
}

// UpdatePlanRequest - DTO para actualizar un plan
type UpdatePlanRequest struct {
	Nombre                *string   `json:"nombre,omitempty" binding:"omitempty,min=3,max=100"`
	Descripcion           *string   `json:"descripcion,omitempty" binding:"omitempty,max=500"`
	PrecioMensual         *float64  `json:"precio_mensual,omitempty" binding:"omitempty,gt=0"`
	TipoAcceso            *string   `json:"tipo_acceso,omitempty" binding:"omitempty,oneof=limitado completo"`
	DuracionDias          *int      `json:"duracion_dias,omitempty" binding:"omitempty,gt=0"`
	Activo                *bool     `json:"activo,omitempty"`
	ActividadesPermitidas *[]string `json:"actividades_permitidas,omitempty"`
}

// PlanResponse - DTO para respuesta de un plan
type PlanResponse struct {
	ID                    string    `json:"id"`
	Nombre                string    `json:"nombre"`
	Descripcion           string    `json:"descripcion"`
	PrecioMensual         float64   `json:"precio_mensual"`
	TipoAcceso            string    `json:"tipo_acceso"`
	DuracionDias          int       `json:"duracion_dias"`
	Activo                bool      `json:"activo"`
	ActividadesPermitidas []string  `json:"actividades_permitidas"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ListPlansQuery - DTO para query params de listado
type ListPlansQuery struct {
	Activo   *bool  `form:"activo"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy   string `form:"sort_by" binding:"omitempty,oneof=nombre precio_mensual created_at"`
	SortDesc bool   `form:"sort_desc"`
}

// PaginatedPlansResponse - DTO para respuesta paginada de planes
type PaginatedPlansResponse struct {
	Plans      []PlanResponse `json:"plans"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
