package repository

import (
	"activities-api/internal/dao"
	"activities-api/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InscripcionesRepository define la interfaz del repositorio de inscripciones
type InscripcionesRepository interface {
	ListByUser(ctx context.Context, usuarioID uint) ([]domain.Inscripcion, error)
	GetByUserAndActividad(ctx context.Context, usuarioID, actividadID uint) (domain.Inscripcion, error)
	Create(ctx context.Context, inscripcion domain.Inscripcion) (domain.Inscripcion, error)
	Deactivate(ctx context.Context, usuarioID, actividadID uint) error
}

// MySQLInscripcionesRepository implementa InscripcionesRepository usando MySQL/GORM
// Migrado de backend/clients/inscripcion/inscripcion_client.go
type MySQLInscripcionesRepository struct {
	db *gorm.DB
}

// NewMySQLInscripcionesRepository crea una nueva instancia del repository
// Comparte la conexión DB con ActividadesRepository
func NewMySQLInscripcionesRepository(db *gorm.DB) *MySQLInscripcionesRepository {
	// Auto-migration disabled - using SQL initialization scripts instead (BDD/01-init.sql)
	// if err := db.AutoMigrate(&dao.Inscripcion{}); err != nil {
	// 	fmt.Printf("Error auto-migrating Inscripcion table: %v\n", err)
	// }

	return &MySQLInscripcionesRepository{
		db: db,
	}
}

// ListByUser obtiene todas las inscripciones de un usuario
// Solo devuelve inscripciones de actividades NO eliminadas (deleted_at IS NULL)
// Migrado de backend/clients/inscripcion/inscripcion_client.go:12
func (r *MySQLInscripcionesRepository) ListByUser(ctx context.Context, usuarioID uint) ([]domain.Inscripcion, error) {
	var inscripcionesDAO []dao.Inscripcion

	err := r.db.WithContext(ctx).
		Joins("JOIN actividades ON inscripciones.actividad_id = actividades.id_actividad").
		Where("inscripciones.usuario_id = ?", usuarioID).
		Where("actividades.deleted_at IS NULL").
		Find(&inscripcionesDAO).Error

	if err != nil {
		return nil, fmt.Errorf("error listing inscripciones: %w", err)
	}

	// Convertir a Domain
	inscripciones := make([]domain.Inscripcion, len(inscripcionesDAO))
	for i, inscDAO := range inscripcionesDAO {
		inscripciones[i] = inscDAO.ToDomain()
	}

	return inscripciones, nil
}

// GetByUserAndActividad obtiene una inscripción específica
// Solo devuelve la inscripción si la actividad NO está eliminada (deleted_at IS NULL)
func (r *MySQLInscripcionesRepository) GetByUserAndActividad(ctx context.Context, usuarioID, actividadID uint) (domain.Inscripcion, error) {
	var inscripcionDAO dao.Inscripcion

	err := r.db.WithContext(ctx).
		Joins("JOIN actividades ON inscripciones.actividad_id = actividades.id_actividad").
		Where("inscripciones.usuario_id = ? AND inscripciones.actividad_id = ?", usuarioID, actividadID).
		Where("actividades.deleted_at IS NULL").
		First(&inscripcionDAO).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Inscripcion{}, errors.New("inscripcion not found")
		}
		return domain.Inscripcion{}, fmt.Errorf("error getting inscripcion: %w", err)
	}

	return inscripcionDAO.ToDomain(), nil
}

// Create crea una nueva inscripción o reactiva una existente
// Migrado de backend/clients/inscripcion/inscripcion_client.go:27
func (r *MySQLInscripcionesRepository) Create(ctx context.Context, inscripcion domain.Inscripcion) (domain.Inscripcion, error) {
	inscripcionDAO := dao.InscripcionFromDomain(inscripcion)
	inscripcionDAO.FechaInscripcion = time.Now()
	inscripcionDAO.IsActiva = true

	// Intentar buscar inscripción existente (por usuario y actividad)
	var existing dao.Inscripcion
	err := r.db.WithContext(ctx).
		Where("usuario_id = ? AND actividad_id = ?", inscripcionDAO.UsuarioID, inscripcionDAO.ActividadID).
		First(&existing).Error

	if err == nil {
		// La inscripción ya existe
		if existing.IsActiva {
			return domain.Inscripcion{}, errors.New("el usuario ya está inscripto en esta actividad")
		}

		// Reactivar inscripción (ejecuta hook BeforeUpdate)
		existing.IsActiva = true
		if err := r.db.WithContext(ctx).Model(&existing).Update("is_activa", true).Error; err != nil {
			return domain.Inscripcion{}, fmt.Errorf("error reactivating inscripcion: %w", err)
		}

		return existing.ToDomain(), nil
	}

	// No existe, crear nueva (ejecuta hook BeforeCreate)
	if err := r.db.WithContext(ctx).Create(&inscripcionDAO).Error; err != nil {
		return domain.Inscripcion{}, fmt.Errorf("error creating inscripcion: %w", err)
	}

	return inscripcionDAO.ToDomain(), nil
}

// Deactivate desactiva una inscripción (soft delete lógico)
// Migrado de backend/clients/inscripcion/inscripcion_client.go:53
func (r *MySQLInscripcionesRepository) Deactivate(ctx context.Context, usuarioID, actividadID uint) error {
	result := r.db.WithContext(ctx).
		Model(&dao.Inscripcion{}).
		Where("usuario_id = ? AND actividad_id = ?", usuarioID, actividadID).
		Update("is_activa", false)

	if result.Error != nil {
		return fmt.Errorf("error deactivating inscripcion: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("inscripcion not found")
	}

	return nil
}

// TODO: Los compañeros deben agregar:
// - Validación de usuario existe (HTTP call a users-api)
// - Validación de suscripción activa (HTTP call a subscriptions-api)
// - Publicación de eventos a RabbitMQ
