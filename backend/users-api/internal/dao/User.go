package dao

import (
	"time"
	"users-api/internal/domain"
)

// User representa el modelo de base de datos con tags de GORM
// Este modelo está acoplado a MySQL/GORM
type User struct {
	ID               uint       `gorm:"column:id_usuario;primaryKey;autoIncrement"`
	Nombre           string     `gorm:"type:varchar(30);not null"`
	Apellido         string     `gorm:"type:varchar(30);not null"`
	Username         string     `gorm:"type:varchar(30);unique;not null;index"`
	Email            string     `gorm:"type:varchar(100);unique;not null;index"`
	Password         string     `gorm:"type:char(64);collation:ascii_bin;not null"`
	IsAdmin          bool       `gorm:"column:is_admin;default:false;not null"`
	SucursalOrigenID *uint      `gorm:"column:sucursal_origen_id;index"` // Nullable, referencia lógica
	FechaRegistro    time.Time  `gorm:"column:fecha_registro;type:timestamp;default:CURRENT_TIMESTAMP;not null"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
	DeletedAt        *time.Time `gorm:"index"` // Soft delete
}

// TableName especifica el nombre de la tabla en MySQL
func (User) TableName() string {
	return "usuarios"
}

// ToDomain convierte de DAO (MySQL) a Domain (negocio)
func (u User) ToDomain() domain.User {
	return domain.User{
		ID:               u.ID,
		Nombre:           u.Nombre,
		Apellido:         u.Apellido,
		Username:         u.Username,
		Email:            u.Email,
		Password:         u.Password,
		IsAdmin:          u.IsAdmin,
		SucursalOrigenID: u.SucursalOrigenID,
		FechaRegistro:    u.FechaRegistro,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}

// FromDomain convierte de Domain (negocio) a DAO (MySQL)
func FromDomain(domainUser domain.User) User {
	return User{
		ID:               domainUser.ID,
		Nombre:           domainUser.Nombre,
		Apellido:         domainUser.Apellido,
		Username:         domainUser.Username,
		Email:            domainUser.Email,
		Password:         domainUser.Password,
		IsAdmin:          domainUser.IsAdmin,
		SucursalOrigenID: domainUser.SucursalOrigenID,
		FechaRegistro:    domainUser.FechaRegistro,
		CreatedAt:        domainUser.CreatedAt,
		UpdatedAt:        domainUser.UpdatedAt,
	}
}
