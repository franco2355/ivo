package dao

import (
	"activities-api/internal/domain"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Actividad representa el modelo de base de datos con tags de GORM
// Migrado de backend/model/actividad.go
type Actividad struct {
	ID            uint           `gorm:"column:id_actividad;primaryKey;autoIncrement"`
	Titulo        string         `gorm:"type:varchar(50);not null"`
	Descripcion   string         `gorm:"type:varchar(255)"`
	Cupo          uint           `gorm:"type:int;not null"`
	Dia           string         `gorm:"type:enum('Lunes','Martes','Miercoles','Jueves','Viernes','Sabado','Domingo');not null"`
	HorarioInicio time.Time      `gorm:"column:horario_inicio;type:time;not null"`
	HorarioFinal  time.Time      `gorm:"column:horario_final;type:time;not null"`
	FotoUrl       string         `gorm:"column:foto_url;type:varchar(511);not null"`
	Instructor    string         `gorm:"type:varchar(50);not null"`
	Categoria     string         `gorm:"type:varchar(40);not null"`
	SucursalID    *uint          `gorm:"column:sucursal_id;index"` // TODO: Agregar FK cuando se cree Sucursal
	Activa        bool           `gorm:"column:activa;default:true;not null"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relación con Inscripciones
	Inscripciones []Inscripcion `gorm:"foreignKey:ActividadID;constraint:OnDelete:CASCADE"`
}

// TableName especifica el nombre de la tabla
func (Actividad) TableName() string {
	return "actividades"
}

// BeforeUpdate es un hook de GORM que valida el cupo antes de actualizar
// Migrado del código original (backend/model/actividad.go:28)
func (ac *Actividad) BeforeUpdate(tx *gorm.DB) (err error) {
	var inscActivas int64

	err = tx.Model(&Inscripcion{}).
		Where("actividad_id = ? AND is_activa = ?", ac.ID, true).
		Count(&inscActivas).Error
	if err != nil {
		return err
	}

	if ac.Cupo < uint(inscActivas) {
		return fmt.Errorf("no se puede cambiar el cupo, hay inscripciones activas que superan el nuevo límite")
	}

	return nil
}

// ToDomain convierte de DAO (MySQL) a Domain (negocio)
func (a Actividad) ToDomain() domain.Actividad {
	var deletedAt *time.Time
	if a.DeletedAt.Valid {
		deletedAt = &a.DeletedAt.Time
	}

	return domain.Actividad{
		ID:            a.ID,
		Titulo:        a.Titulo,
		Descripcion:   a.Descripcion,
		Cupo:          a.Cupo,
		Dia:           a.Dia,
		HorarioInicio: a.HorarioInicio.Format("15:04"),
		HorarioFinal:  a.HorarioFinal.Format("15:04"),
		FotoUrl:       a.FotoUrl,
		Instructor:    a.Instructor,
		Categoria:     a.Categoria,
		SucursalID:    a.SucursalID,
		Activa:        a.Activa,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

// FromDomain convierte de Domain (negocio) a DAO (MySQL)
func ActividadFromDomain(domainAct domain.Actividad, horaInicio, horaFin time.Time) Actividad {
	return Actividad{
		ID:            domainAct.ID,
		Titulo:        domainAct.Titulo,
		Descripcion:   domainAct.Descripcion,
		Cupo:          domainAct.Cupo,
		Dia:           domainAct.Dia,
		HorarioInicio: horaInicio,
		HorarioFinal:  horaFin,
		FotoUrl:       domainAct.FotoUrl,
		Instructor:    domainAct.Instructor,
		Categoria:     domainAct.Categoria,
		SucursalID:    domainAct.SucursalID,
	}
}

// ActividadVista representa la vista MySQL con cupos calculados
// Migrado de backend/model/actividad.go:45
type ActividadVista struct {
	ID            uint      `gorm:"column:id_actividad;primaryKey"`
	Titulo        string    `gorm:"type:varchar(50)"`
	Descripcion   string    `gorm:"type:varchar(255)"`
	Cupo          uint      `gorm:"type:int"`
	Dia           string    `gorm:"type:varchar(20)"`
	HorarioInicio time.Time `gorm:"column:horario_inicio;type:time"`
	HorarioFinal  time.Time `gorm:"column:horario_final;type:time"`
	FotoUrl       string    `gorm:"column:foto_url;type:varchar(511)"`
	Instructor    string    `gorm:"type:varchar(50)"`
	Categoria     string    `gorm:"type:varchar(40)"`
	Lugares       uint      `gorm:"column:lugares"` // Campo calculado de la vista
	SucursalID    *uint     `gorm:"column:sucursal_id"`
}

// TableName especifica el nombre de la vista
func (ActividadVista) TableName() string {
	return "actividades_lugares" // Vista que calcula lugares disponibles
}

// ToDomain convierte vista a domain (incluye lugares disponibles)
func (av ActividadVista) ToDomain() domain.Actividad {
	return domain.Actividad{
		ID:            av.ID,
		Titulo:        av.Titulo,
		Descripcion:   av.Descripcion,
		Cupo:          av.Cupo,
		Dia:           av.Dia,
		HorarioInicio: av.HorarioInicio.Format("15:04"),
		HorarioFinal:  av.HorarioFinal.Format("15:04"),
		FotoUrl:       av.FotoUrl,
		Instructor:    av.Instructor,
		Categoria:     av.Categoria,
		Lugares:       av.Lugares, // Incluye cupos disponibles
		SucursalID:    av.SucursalID,
	}
}
