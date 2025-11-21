package domain

import "time"

// Inscripcion representa la entidad de negocio Inscripcion
type Inscripcion struct {
	ID               uint      `json:"id"`
	UsuarioID        uint      `json:"usuario_id"`
	ActividadID      uint      `json:"actividad_id"`
	FechaInscripcion time.Time `json:"fecha_inscripcion"`
	IsActiva         bool      `json:"is_activa"`
	SuscripcionID    *string   `json:"suscripcion_id,omitempty"` // TODO: Agregar cuando se implemente subscriptions-api
	CreatedAt        time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

// InscripcionCreate representa los datos para crear una inscripción
type InscripcionCreate struct {
	UsuarioID   uint `json:"usuario_id" binding:"required"`
	ActividadID uint `json:"actividad_id" binding:"required"`
	// TODO: Agregar SuscripcionID cuando se implemente subscriptions-api
}

// InscripcionResponse representa la respuesta con datos adicionales
type InscripcionResponse struct {
	ID               uint      `json:"id"`
	UsuarioID        uint      `json:"usuario_id"`
	ActividadID      uint      `json:"actividad_id"`
	FechaInscripcion time.Time `json:"fecha_inscripcion"`
	IsActiva         bool      `json:"is_activa"`
	SuscripcionID    *string   `json:"suscripcion_id,omitempty"`
	// Puede incluir datos de la actividad si se necesita
	ActividadTitulo *string `json:"actividad_titulo,omitempty"`
}

// ToResponse convierte de Inscripcion a InscripcionResponse
func (i Inscripcion) ToResponse() InscripcionResponse {
	return InscripcionResponse{
		ID:               i.ID,
		UsuarioID:        i.UsuarioID,
		ActividadID:      i.ActividadID,
		FechaInscripcion: i.FechaInscripcion,
		IsActiva:         i.IsActiva,
		SuscripcionID:    i.SuscripcionID,
		// ActividadTitulo se puede agregar después si se necesita (JOIN)
	}
}
