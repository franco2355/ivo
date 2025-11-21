package domain

import "time"

// User representa la entidad de negocio Usuario
// Este modelo es independiente de la base de datos
type User struct {
	ID                uint      `json:"id"`
	Nombre            string    `json:"nombre"`
	Apellido          string    `json:"apellido"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	Password          string    `json:"password,omitempty"` // omitempty para no exponerlo en responses
	IsAdmin           bool      `json:"is_admin"`
	SucursalOrigenID  *uint     `json:"sucursal_origen_id,omitempty"` // Nullable
	FechaRegistro     time.Time `json:"fecha_registro"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UserLogin representa las credenciales de login
type UserLogin struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// UserRegister representa los datos para registrar un nuevo usuario
type UserRegister struct {
	Nombre           string `json:"nombre" binding:"required"`
	Apellido         string `json:"apellido" binding:"required"`
	Username         string `json:"username" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	SucursalOrigenID *uint  `json:"sucursal_origen_id,omitempty"`
}

// UserResponse representa la respuesta pública del usuario (sin password)
type UserResponse struct {
	ID               uint      `json:"id"`
	Nombre           string    `json:"nombre"`
	Apellido         string    `json:"apellido"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	IsAdmin          bool      `json:"is_admin"`
	SucursalOrigenID *uint     `json:"sucursal_origen_id,omitempty"`
	FechaRegistro    time.Time `json:"fecha_registro"`
}

// ToResponse convierte User a UserResponse (sin password)
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:               u.ID,
		Nombre:           u.Nombre,
		Apellido:         u.Apellido,
		Username:         u.Username,
		Email:            u.Email,
		IsAdmin:          u.IsAdmin,
		SucursalOrigenID: u.SucursalOrigenID,
		FechaRegistro:    u.FechaRegistro,
	}
}
