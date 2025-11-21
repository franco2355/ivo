package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"users-api/internal/domain"
	"users-api/internal/repository"

	"github.com/golang-jwt/jwt/v4"
)

// UsersService define la interfaz del servicio de usuarios
type UsersService interface {
	Register(ctx context.Context, userReg domain.UserRegister) (domain.UserResponse, string, error)
	Login(ctx context.Context, credentials domain.UserLogin) (domain.UserResponse, string, error)
	GetByID(ctx context.Context, id uint) (domain.UserResponse, error)
	List(ctx context.Context) ([]domain.UserResponse, error)
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

// UsersServiceImpl implementa UsersService
type UsersServiceImpl struct {
	repository repository.UsersRepository
	jwtSecret  string
}

// NewUsersService crea una nueva instancia del servicio
// Dependency Injection: recibe el repository como parámetro
func NewUsersService(repo repository.UsersRepository, jwtSecret string) *UsersServiceImpl {
	return &UsersServiceImpl{
		repository: repo,
		jwtSecret:  jwtSecret,
	}
}

// Register registra un nuevo usuario
func (s *UsersServiceImpl) Register(ctx context.Context, userReg domain.UserRegister) (domain.UserResponse, string, error) {
	// Validaciones de negocio
	if err := s.validateUserRegistration(userReg); err != nil {
		return domain.UserResponse{}, "", err
	}

	// Hashear password
	hashedPassword := s.hashPassword(userReg.Password)

	// Crear domain user
	user := domain.User{
		Nombre:           userReg.Nombre,
		Apellido:         userReg.Apellido,
		Username:         userReg.Username,
		Email:            userReg.Email,
		Password:         hashedPassword,
		IsAdmin:          false, // Por defecto, usuario normal
		SucursalOrigenID: userReg.SucursalOrigenID,
	}

	// Guardar en repository
	createdUser, err := s.repository.Create(ctx, user)
	if err != nil {
		return domain.UserResponse{}, "", fmt.Errorf("error creating user: %w", err)
	}

	// Generar token JWT
	token, err := s.generateToken(createdUser)
	if err != nil {
		return domain.UserResponse{}, "", fmt.Errorf("error generating token: %w", err)
	}

	return createdUser.ToResponse(), token, nil
}

// Login autentica un usuario y devuelve un token JWT
func (s *UsersServiceImpl) Login(ctx context.Context, credentials domain.UserLogin) (domain.UserResponse, string, error) {
	// Buscar usuario por username o email
	user, err := s.repository.GetByUsernameOrEmail(ctx, credentials.UsernameOrEmail)
	if err != nil {
		return domain.UserResponse{}, "", errors.New("invalid credentials")
	}

	// Verificar password
	hashedPassword := s.hashPassword(credentials.Password)
	if user.Password != hashedPassword {
		return domain.UserResponse{}, "", errors.New("invalid credentials")
	}

	// Generar token JWT
	token, err := s.generateToken(user)
	if err != nil {
		return domain.UserResponse{}, "", fmt.Errorf("error generating token: %w", err)
	}

	return user.ToResponse(), token, nil
}

// GetByID obtiene un usuario por su ID
func (s *UsersServiceImpl) GetByID(ctx context.Context, id uint) (domain.UserResponse, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.UserResponse{}, err
	}

	return user.ToResponse(), nil
}

// List obtiene todos los usuarios
func (s *UsersServiceImpl) List(ctx context.Context) ([]domain.UserResponse, error) {
	users, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}

	// Convertir a UserResponse
	responses := make([]domain.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

// ValidateToken valida un token JWT y devuelve los claims
func (s *UsersServiceImpl) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Validar que el método de firma es HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// generateToken genera un token JWT para el usuario
func (s *UsersServiceImpl) generateToken(user domain.User) (string, error) {
	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	claims := jwt.MapClaims{
		"iss":        "gym-management-system",
		"exp":        time.Now().Add(30 * time.Minute).Unix(), // Expira en 30 minutos
		"username":   user.Username,
		"id_usuario": user.ID,
		"is_admin":   user.IsAdmin,
		"role":       role, // Add role claim for compatibility with other services
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// hashPassword hashea una contraseña usando SHA256
func (s *UsersServiceImpl) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// validateUserRegistration valida los datos de registro
func (s *UsersServiceImpl) validateUserRegistration(userReg domain.UserRegister) error {
	// Validar nombre
	if strings.TrimSpace(userReg.Nombre) == "" {
		return errors.New("nombre is required")
	}
	if len(userReg.Nombre) > 30 {
		return errors.New("nombre must be at most 30 characters")
	}

	// Validar apellido
	if strings.TrimSpace(userReg.Apellido) == "" {
		return errors.New("apellido is required")
	}
	if len(userReg.Apellido) > 30 {
		return errors.New("apellido must be at most 30 characters")
	}

	// Validar username
	if strings.TrimSpace(userReg.Username) == "" {
		return errors.New("username is required")
	}
	if len(userReg.Username) < 3 || len(userReg.Username) > 30 {
		return errors.New("username must be between 3 and 30 characters")
	}
	// Username solo puede contener letras, números, guiones y guiones bajos
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(userReg.Username) {
		return errors.New("username can only contain letters, numbers, hyphens and underscores")
	}

	// Validar email
	if strings.TrimSpace(userReg.Email) == "" {
		return errors.New("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(userReg.Email) {
		return errors.New("invalid email format")
	}

	// Validar password
	if strings.TrimSpace(userReg.Password) == "" {
		return errors.New("password is required")
	}
	if len(userReg.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// Password debe tener al menos una letra mayúscula, una minúscula y un número
	if !regexp.MustCompile(`[A-Z]`).MatchString(userReg.Password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(userReg.Password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(userReg.Password) {
		return errors.New("password must contain at least one number")
	}

	return nil
}
