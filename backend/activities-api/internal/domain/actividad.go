package domain

import "time"

// Actividad representa la entidad de negocio Actividad
// Independiente de la base de datos
type Actividad struct {
	ID            uint       `json:"id"`
	Titulo        string     `json:"titulo"`
	Descripcion   string     `json:"descripcion"`
	Cupo          uint       `json:"cupo"`
	Dia           string     `json:"dia"`
	HorarioInicio string     `json:"horario_inicio"` // Formato "HH:MM"
	HorarioFinal  string     `json:"horario_final"`  // Formato "HH:MM"
	FotoUrl       string     `json:"foto_url"`
	Instructor    string     `json:"instructor"`
	Categoria     string     `json:"categoria"`
	SucursalID    *uint      `json:"sucursal_id,omitempty"` // TODO: Agregar cuando se cree entidad Sucursal
	Lugares       uint       `json:"lugares,omitempty"`     // Campo calculado (cupos disponibles)
	CreatedAt     time.Time  `json:"created_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at,omitempty"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"` // Soft delete
}

// ActividadCreate representa los datos para crear una actividad
type ActividadCreate struct {
	Titulo        string `json:"titulo" binding:"required"`
	Descripcion   string `json:"descripcion"`
	Cupo          uint   `json:"cupo" binding:"required,min=1"`
	Dia           string `json:"dia" binding:"required"`
	HorarioInicio string `json:"horario_inicio" binding:"required"` // "HH:MM"
	HorarioFinal  string `json:"horario_final" binding:"required"`  // "HH:MM"
	FotoUrl       string `json:"foto_url" binding:"required"`
	Instructor    string `json:"instructor" binding:"required"`
	Categoria     string `json:"categoria" binding:"required"`
	SucursalID    *uint  `json:"sucursal_id,omitempty"` // TODO: Validar que existe
}

// ActividadUpdate representa los datos para actualizar una actividad
type ActividadUpdate struct {
	Titulo        string `json:"titulo" binding:"required"`
	Descripcion   string `json:"descripcion"`
	Cupo          uint   `json:"cupo" binding:"required,min=1"`
	Dia           string `json:"dia" binding:"required"`
	HorarioInicio string `json:"horario_inicio" binding:"required"`
	HorarioFinal  string `json:"horario_final" binding:"required"`
	FotoUrl       string `json:"foto_url" binding:"required"`
	Instructor    string `json:"instructor" binding:"required"`
	Categoria     string `json:"categoria" binding:"required"`
	SucursalID    *uint  `json:"sucursal_id,omitempty"`
}

// ActividadResponse representa la respuesta HTTP de una actividad
type ActividadResponse struct {
	ID            uint   `json:"id"`
	Titulo        string `json:"titulo"`
	Descripcion   string `json:"descripcion"`
	Cupo          uint   `json:"cupo"`
	Dia           string `json:"dia"`
	HorarioInicio string `json:"horario_inicio"` // "HH:MM"
	HorarioFinal  string `json:"horario_final"`  // "HH:MM"
	FotoUrl       string `json:"foto_url"`
	Instructor    string `json:"instructor"`
	Categoria     string `json:"categoria"`
	SucursalID    *uint  `json:"sucursal_id,omitempty"`
	Lugares       uint   `json:"lugares"` // Campo calculado de cupos disponibles
}

// ToResponse convierte de Actividad a ActividadResponse
func (a Actividad) ToResponse() ActividadResponse {
	return ActividadResponse{
		ID:            a.ID,
		Titulo:        a.Titulo,
		Descripcion:   a.Descripcion,
		Cupo:          a.Cupo,
		Dia:           a.Dia,
		HorarioInicio: a.HorarioInicio,
		HorarioFinal:  a.HorarioFinal,
		FotoUrl:       a.FotoUrl,
		Instructor:    a.Instructor,
		Categoria:     a.Categoria,
		SucursalID:    a.SucursalID,
		Lugares:       a.Lugares,
	}
}
