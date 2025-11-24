package repository

import (
	"activities-api/internal/config"
	"activities-api/internal/dao"
	"activities-api/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ActividadesRepository define la interfaz del repositorio de actividades
type ActividadesRepository interface {
	List(ctx context.Context) ([]domain.Actividad, error)
	GetByID(ctx context.Context, id uint) (domain.Actividad, error)
	GetByParams(ctx context.Context, params map[string]interface{}) ([]domain.Actividad, error)
	Create(ctx context.Context, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error)
	Update(ctx context.Context, id uint, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error)
	Delete(ctx context.Context, id uint) error
}

// MySQLActividadesRepository implementa ActividadesRepository usando MySQL/GORM
// Migrado de backend/clients/actividad/actividad_client.go
type MySQLActividadesRepository struct {
	db *gorm.DB
}

// GetDB retorna la conexión de base de datos para compartir con otros repositorios
func (r *MySQLActividadesRepository) GetDB() *gorm.DB {
	return r.db
}

// NewMySQLActividadesRepository crea una nueva instancia del repository
func NewMySQLActividadesRepository(cfg config.MySQLConfig) *MySQLActividadesRepository {
	// Construir DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Pass,
		cfg.Host,
		cfg.Port,
		cfg.Schema,
	)

	// Conectar a MySQL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
		return nil
	}

	// Configurar connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error getting SQL DB: %v", err)
		return nil
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migration disabled - using SQL initialization scripts instead (BDD/01-init.sql)
	// if err := db.AutoMigrate(&dao.Sucursal{}); err != nil {
	// 	log.Fatalf("Error auto-migrating Sucursal table: %v", err)
	// 	return nil
	// }
	//
	// if err := db.AutoMigrate(&dao.Actividad{}); err != nil {
	// 	log.Fatalf("Error auto-migrating Actividad table: %v", err)
	// 	return nil
	// }

	// Crear vista actividades_lugares si no existe
	createViewSQL := `
		CREATE OR REPLACE VIEW actividades_lugares AS
		SELECT a.*,
		       a.cupo - COALESCE((SELECT COUNT(*)
		                          FROM inscripciones i
		                          WHERE i.actividad_id = a.id_actividad
		                          AND i.is_activa = true
		                          AND i.deleted_at IS NULL), 0) AS lugares
		FROM actividades a
		WHERE a.deleted_at IS NULL
	`
	if err := db.Exec(createViewSQL).Error; err != nil {
		log.Printf("Warning: Could not create view actividades_lugares: %v", err)
	}

	log.Println("✅ Connected to MySQL successfully (Actividades)")

	return &MySQLActividadesRepository{
		db: db,
	}
}

// List obtiene todas las actividades (usando la vista con lugares)
func (r *MySQLActividadesRepository) List(ctx context.Context) ([]domain.Actividad, error) {
	var actividadesDAO []dao.ActividadVista

	if err := r.db.WithContext(ctx).Find(&actividadesDAO).Error; err != nil {
		return nil, fmt.Errorf("error listing actividades: %w", err)
	}

	// Convertir de DAO a Domain
	actividades := make([]domain.Actividad, len(actividadesDAO))
	for i, actDAO := range actividadesDAO {
		actividades[i] = actDAO.ToDomain()
	}

	return actividades, nil
}

// GetByID obtiene una actividad por ID (usando la vista)
func (r *MySQLActividadesRepository) GetByID(ctx context.Context, id uint) (domain.Actividad, error) {
	var actividadDAO dao.ActividadVista

	err := r.db.WithContext(ctx).Where("id_actividad = ?", id).First(&actividadDAO).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Actividad{}, errors.New("actividad not found")
		}
		return domain.Actividad{}, fmt.Errorf("error getting actividad by ID: %w", err)
	}

	return actividadDAO.ToDomain(), nil
}

// GetByParams busca actividades por parámetros
// Migrado de backend/clients/actividad/actividad_client.go:11
func (r *MySQLActividadesRepository) GetByParams(ctx context.Context, params map[string]interface{}) ([]domain.Actividad, error) {
	var actividadesDAO []dao.ActividadVista
	query := r.db.WithContext(ctx).Model(&dao.ActividadVista{})

	// Filtros opcionales
	if id, ok := params["id"]; ok && id != "" {
		query = query.Where("id_actividad = ?", id)
	}
	if titulo, ok := params["titulo"]; ok && titulo != "" {
		query = query.Where("titulo LIKE ?", fmt.Sprintf("%%%s%%", titulo))
	}
	if horario, ok := params["horario"]; ok && horario != "" {
		query = query.Where("TIME(?) BETWEEN TIME(horario_inicio) AND TIME(horario_final)", horario)
	}
	if categoria, ok := params["categoria"]; ok && categoria != "" {
		query = query.Where("categoria LIKE ?", fmt.Sprintf("%%%s%%", categoria))
	}

	if err := query.Find(&actividadesDAO).Error; err != nil {
		return nil, fmt.Errorf("error searching actividades: %w", err)
	}

	// Convertir a Domain
	actividades := make([]domain.Actividad, len(actividadesDAO))
	for i, actDAO := range actividadesDAO {
		actividades[i] = actDAO.ToDomain()
	}

	return actividades, nil
}

// Create inserta una nueva actividad
func (r *MySQLActividadesRepository) Create(ctx context.Context, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error) {
	actividadDAO := dao.ActividadFromDomain(actividad, horaInicio, horaFin)
	actividadDAO.CreatedAt = time.Now()
	actividadDAO.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&actividadDAO).Error; err != nil {
		return domain.Actividad{}, fmt.Errorf("error creating actividad: %w", err)
	}

	return actividadDAO.ToDomain(), nil
}

// Update actualiza una actividad existente
func (r *MySQLActividadesRepository) Update(ctx context.Context, id uint, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error) {
	actividadDAO := dao.ActividadFromDomain(actividad, horaInicio, horaFin)
	actividadDAO.ID = id
	actividadDAO.UpdatedAt = time.Now()

	// GORM ejecutará el hook BeforeUpdate que valida cupos
	result := r.db.WithContext(ctx).Model(&dao.Actividad{ID: id, Cupo: actividadDAO.Cupo}).Updates(&actividadDAO)
	if result.Error != nil {
		return domain.Actividad{}, fmt.Errorf("error updating actividad: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.Actividad{}, errors.New("actividad not found")
	}

	// Obtener la actividad actualizada
	return r.GetByID(ctx, id)
}

// Delete elimina una actividad
func (r *MySQLActividadesRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id_actividad = ?", id).Delete(&dao.Actividad{})
	if result.Error != nil {
		return fmt.Errorf("error deleting actividad: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("actividad not found")
	}

	return nil
}
