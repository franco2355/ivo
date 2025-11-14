package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"users-api/internal/config"
	"users-api/internal/dao"
	"users-api/internal/domain"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// UsersRepository define la interfaz del repositorio de usuarios
// Esta interfaz permite cambiar la implementación (MySQL, Postgres, etc.) sin afectar el servicio
type UsersRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id uint) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, id uint, user domain.User) (domain.User, error)
	Delete(ctx context.Context, id uint) error
}

// MySQLUsersRepository implementa UsersRepository usando MySQL/GORM
type MySQLUsersRepository struct {
	db *gorm.DB
}

// NewMySQLUsersRepository crea una nueva instancia del repository
func NewMySQLUsersRepository(cfg config.MySQLConfig) *MySQLUsersRepository {
	// Construir DSN (Data Source Name)
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

	// Auto-migration (crea/actualiza tabla)
	/*if err := db.AutoMigrate(&dao.User{}); err != nil {
		log.Fatalf("Error auto-migrating User table: %v", err)
		return nil
	}*/
	log.Println("✅ Connected to MySQL successfully")

	return &MySQLUsersRepository{
		db: db,
	}
}

// Create inserta un nuevo usuario en la base de datos
func (r *MySQLUsersRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	userDAO := dao.FromDomain(user)
	userDAO.FechaRegistro = time.Now()
	userDAO.CreatedAt = time.Now()
	userDAO.UpdatedAt = time.Now()

	// Insertar en DB
	if err := r.db.WithContext(ctx).Create(&userDAO).Error; err != nil {
		// Manejar errores de duplicados (username o email ya existe)
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			(err.Error() != "" && (contains(err.Error(), "username") ||
				contains(err.Error(), "email") ||
				contains(err.Error(), "Duplicate entry"))) {
			return domain.User{}, errors.New("username or email already exists")
		}
		return domain.User{}, fmt.Errorf("error creating user: %w", err)
	}

	return userDAO.ToDomain(), nil
}

// GetByID busca un usuario por su ID
func (r *MySQLUsersRepository) GetByID(ctx context.Context, id uint) (domain.User, error) {
	var userDAO dao.User

	err := r.db.WithContext(ctx).First(&userDAO, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, fmt.Errorf("error getting user by ID: %w", err)
	}

	return userDAO.ToDomain(), nil
}

// GetByUsername busca un usuario por su username
func (r *MySQLUsersRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	var userDAO dao.User

	err := r.db.WithContext(ctx).Where("username = ?", username).First(&userDAO).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, fmt.Errorf("error getting user by username: %w", err)
	}

	return userDAO.ToDomain(), nil
}

// GetByEmail busca un usuario por su email
func (r *MySQLUsersRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var userDAO dao.User

	err := r.db.WithContext(ctx).Where("email = ?", email).First(&userDAO).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, fmt.Errorf("error getting user by email: %w", err)
	}

	return userDAO.ToDomain(), nil
}

// GetByUsernameOrEmail busca un usuario por username o email
func (r *MySQLUsersRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (domain.User, error) {
	var userDAO dao.User

	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		First(&userDAO).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, fmt.Errorf("error getting user: %w", err)
	}

	return userDAO.ToDomain(), nil
}

// List obtiene todos los usuarios
func (r *MySQLUsersRepository) List(ctx context.Context) ([]domain.User, error) {
	var usersDAO []dao.User

	if err := r.db.WithContext(ctx).Find(&usersDAO).Error; err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	// Convertir de DAO a Domain
	users := make([]domain.User, len(usersDAO))
	for i, userDAO := range usersDAO {
		users[i] = userDAO.ToDomain()
	}

	return users, nil
}

// Update actualiza un usuario existente
func (r *MySQLUsersRepository) Update(ctx context.Context, id uint, user domain.User) (domain.User, error) {
	userDAO := dao.FromDomain(user)
	userDAO.ID = id
	userDAO.UpdatedAt = time.Now()

	// Actualizar en DB
	result := r.db.WithContext(ctx).Model(&userDAO).Updates(userDAO)
	if result.Error != nil {
		return domain.User{}, fmt.Errorf("error updating user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.User{}, errors.New("user not found")
	}

	// Obtener el usuario actualizado
	return r.GetByID(ctx, id)
}

// Delete elimina un usuario (soft delete)
func (r *MySQLUsersRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&dao.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("error deleting user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
