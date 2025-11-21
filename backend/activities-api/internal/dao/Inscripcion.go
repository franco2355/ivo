package dao

import (
	"activities-api/internal/domain"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Inscripcion representa el modelo de base de datos con tags de GORM
// Migrado de backend/model/inscripcion.go
type Inscripcion struct {
	ID               uint       `gorm:"column:id_inscripcion;primaryKey;autoIncrement"`
	UsuarioID        uint       `gorm:"column:usuario_id;not null;index"`
	ActividadID      uint       `gorm:"column:actividad_id;not null;index"`
	FechaInscripcion time.Time  `gorm:"column:fecha_inscripcion;type:timestamp;default:CURRENT_TIMESTAMP;not null"`
	IsActiva         bool       `gorm:"column:is_activa;default:true;not null"`
	SuscripcionID    *string    `gorm:"column:suscripcion_id;type:varchar(50);index"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;index"` // Soft delete

	// Relaciones
	Actividad Actividad `gorm:"foreignKey:ActividadID;constraint:OnDelete:CASCADE"`
	// Usuario no se mapea aquí porque está en otro microservicio
}

// TableName especifica el nombre de la tabla
func (Inscripcion) TableName() string {
	return "inscripciones"
}

// BeforeCreate es un hook de GORM que valida el cupo antes de crear
// Migrado del código original (backend/model/inscripcion.go:23)
func (ins *Inscripcion) BeforeCreate(tx *gorm.DB) (err error) {
	var lugares int64

	// Consulta la vista para obtener lugares disponibles
	err = tx.Table("actividades_lugares").
		Select("lugares").
		Where("id_actividad = ?", ins.ActividadID).
		Scan(&lugares).Error
	if err != nil {
		return err
	}

	if lugares <= 0 {
		return fmt.Errorf("no se puede inscribir, el cupo de la actividad ha sido alcanzado")
	}

	return nil
}

// BeforeUpdate es un hook de GORM que valida el cupo antes de reactivar
// Migrado del código original (backend/model/inscripcion.go:42)
func (ins *Inscripcion) BeforeUpdate(tx *gorm.DB) (err error) {
	// Solo validar si se está activando la inscripción
	if ins.IsActiva {
		var lugares int64

		err = tx.Table("actividades_lugares").
			Select("lugares").
			Where("id_actividad = ?", ins.ActividadID).
			Scan(&lugares).Error
		if err != nil {
			return err
		}

		if lugares <= 0 {
			return fmt.Errorf("no se puede inscribir, el cupo de la actividad ha sido alcanzado")
		}
	}

	return nil
}

// ToDomain convierte de DAO (MySQL) a Domain (negocio)
func (i Inscripcion) ToDomain() domain.Inscripcion {
	return domain.Inscripcion{
		ID:               i.ID,
		UsuarioID:        i.UsuarioID,
		ActividadID:      i.ActividadID,
		FechaInscripcion: i.FechaInscripcion,
		IsActiva:         i.IsActiva,
		SuscripcionID:    i.SuscripcionID,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
}

// FromDomain convierte de Domain (negocio) a DAO (MySQL)
func InscripcionFromDomain(domainInsc domain.Inscripcion) Inscripcion {
	return Inscripcion{
		ID:               domainInsc.ID,
		UsuarioID:        domainInsc.UsuarioID,
		ActividadID:      domainInsc.ActividadID,
		FechaInscripcion: domainInsc.FechaInscripcion,
		IsActiva:         domainInsc.IsActiva,
		SuscripcionID:    domainInsc.SuscripcionID,
	}
}
