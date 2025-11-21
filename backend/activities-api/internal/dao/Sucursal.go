package dao

import (
	"activities-api/internal/domain"
	"time"
)

// Sucursal representa el modelo de base de datos con tags de GORM
// TODO: Para que los compañeros implementen
type Sucursal struct {
	ID        uint       `gorm:"column:id_sucursal;primaryKey;autoIncrement"`
	Nombre    string     `gorm:"type:varchar(100);not null"`
	Direccion string     `gorm:"type:varchar(255);not null"`
	Telefono  string     `gorm:"type:varchar(20)"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"` // Soft delete

	// Relación con Actividades
	Actividades []Actividad `gorm:"foreignKey:SucursalID"`
}

// TableName especifica el nombre de la tabla
func (Sucursal) TableName() string {
	return "sucursales"
}

// ToDomain convierte de DAO (MySQL) a Domain (negocio)
func (s Sucursal) ToDomain() domain.Sucursal {
	return domain.Sucursal{
		ID:        s.ID,
		Nombre:    s.Nombre,
		Direccion: s.Direccion,
		Telefono:  s.Telefono,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// FromDomain convierte de Domain (negocio) a DAO (MySQL)
func SucursalFromDomain(domainSuc domain.Sucursal) Sucursal {
	return Sucursal{
		ID:        domainSuc.ID,
		Nombre:    domainSuc.Nombre,
		Direccion: domainSuc.Direccion,
		Telefono:  domainSuc.Telefono,
	}
}

// TODO: Los compañeros deben crear el repository completo
