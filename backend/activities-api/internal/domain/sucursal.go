package domain

import "time"

// Sucursal representa la entidad de negocio Sucursal
// TODO: Para que los compañeros implementen CRUD completo
type Sucursal struct {
	ID        uint      `json:"id"`
	Nombre    string    `json:"nombre"`
	Direccion string    `json:"direccion"`
	Telefono  string    `json:"telefono"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// SucursalCreate representa los datos para crear una sucursal
type SucursalCreate struct {
	Nombre    string `json:"nombre" binding:"required"`
	Direccion string `json:"direccion" binding:"required"`
	Telefono  string `json:"telefono" binding:"required"`
}

// TODO: Los compañeros deben implementar:
// - Repository (internal/repository/sucursales_mysql.go)
// - Service (internal/services/sucursales.go)
// - Controller (internal/controllers/sucursales.go)
// - Endpoints en main.go
