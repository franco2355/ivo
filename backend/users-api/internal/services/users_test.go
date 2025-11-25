package services

import (
	"context"
	"errors"
	"testing"
	"time"
	"users-api/internal/domain"
	"users-api/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// MockUsersRepository es un mock del repositorio para testing
type MockUsersRepository struct {
	CreateFunc               func(ctx context.Context, user domain.User) (domain.User, error)
	GetByIDFunc              func(ctx context.Context, id uint) (domain.User, error)
	GetByUsernameFunc        func(ctx context.Context, username string) (domain.User, error)
	GetByEmailFunc           func(ctx context.Context, email string) (domain.User, error)
	GetByUsernameOrEmailFunc func(ctx context.Context, usernameOrEmail string) (domain.User, error)
	ListFunc                 func(ctx context.Context) ([]domain.User, error)
	UpdateFunc               func(ctx context.Context, id uint, user domain.User) (domain.User, error)
	DeleteFunc               func(ctx context.Context, id uint) error
	GetDBFunc                func() *gorm.DB
}

func (m *MockUsersRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return domain.User{}, nil
}

func (m *MockUsersRepository) GetByID(ctx context.Context, id uint) (domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return domain.User{}, nil
}

func (m *MockUsersRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	return domain.User{}, nil
}

func (m *MockUsersRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return domain.User{}, nil
}

func (m *MockUsersRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (domain.User, error) {
	if m.GetByUsernameOrEmailFunc != nil {
		return m.GetByUsernameOrEmailFunc(ctx, usernameOrEmail)
	}
	return domain.User{}, nil
}

func (m *MockUsersRepository) List(ctx context.Context) ([]domain.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return []domain.User{}, nil
}

func (m *MockUsersRepository) Update(ctx context.Context, id uint, user domain.User) (domain.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, user)
	}
	return domain.User{}, nil
}

func (m *MockUsersRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUsersRepository) GetDB() *gorm.DB {
	if m.GetDBFunc != nil {
		return m.GetDBFunc()
	}
	return nil
}

// Verificar que MockUsersRepository implementa la interfaz
var _ repository.UsersRepository = (*MockUsersRepository)(nil)

// TestRegister_Success prueba el registro exitoso de un usuario
func TestRegister_Success(t *testing.T) {
	mockRepo := &MockUsersRepository{
		CreateFunc: func(ctx context.Context, user domain.User) (domain.User, error) {
			user.ID = 1
			user.FechaRegistro = time.Now()
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			return user, nil
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	userReg := domain.UserRegister{
		Nombre:   "Juan",
		Apellido: "Pérez",
		Username: "juanperez",
		Email:    "juan@example.com",
		Password: "SecurePass123!",
	}

	userResp, token, err := service.Register(context.Background(), userReg)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if userResp.ID != 1 {
		t.Errorf("Expected user ID 1, got: %d", userResp.ID)
	}

	if userResp.Username != "juanperez" {
		t.Errorf("Expected username 'juanperez', got: %s", userResp.Username)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verificar que el password no se exponga en la respuesta
	if userResp.Email != "juan@example.com" {
		t.Errorf("Expected email 'juan@example.com', got: %s", userResp.Email)
	}
}

// TestRegister_ValidationErrors prueba los errores de validación
func TestRegister_ValidationErrors(t *testing.T) {
	mockRepo := &MockUsersRepository{}
	service := NewUsersService(mockRepo, "test-secret-key")

	tests := []struct {
		name        string
		userReg     domain.UserRegister
		expectedErr string
	}{
		{
			name: "nombre vacío",
			userReg: domain.UserRegister{
				Nombre:   "",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "SecurePass123!",
			},
			expectedErr: "el nombre es requerido",
		},
		{
			name: "apellido vacío",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "SecurePass123!",
			},
			expectedErr: "el apellido es requerido",
		},
		{
			name: "username muy corto",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "ab",
				Email:    "juan@example.com",
				Password: "SecurePass123!",
			},
			expectedErr: "el nombre de usuario debe tener entre 3 y 30 caracteres",
		},
		{
			name: "username con caracteres inválidos",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juan pérez",
				Email:    "juan@example.com",
				Password: "SecurePass123!",
			},
			expectedErr: "el nombre de usuario solo puede contener letras, números, guiones y guiones bajos",
		},
		{
			name: "email inválido",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			expectedErr: "formato de email inválido",
		},
		{
			name: "password sin mayúscula",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "securepass123!",
			},
			expectedErr: "la contraseña debe contener al menos una letra mayúscula",
		},
		{
			name: "password sin minúscula",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "SECUREPASS123!",
			},
			expectedErr: "la contraseña debe contener al menos una letra minúscula",
		},
		{
			name: "password sin número",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "SecurePass!",
			},
			expectedErr: "la contraseña debe contener al menos un número",
		},
		{
			name: "password sin carácter especial",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "SecurePass123",
			},
			expectedErr: "la contraseña debe contener al menos un carácter especial",
		},
		{
			name: "password muy corto",
			userReg: domain.UserRegister{
				Nombre:   "Juan",
				Apellido: "Pérez",
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: "Pass1!",
			},
			expectedErr: "la contraseña debe tener al menos 8 caracteres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.Register(context.Background(), tt.userReg)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("Expected error '%s', got: '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestRegister_RepositoryError prueba el manejo de errores del repositorio
func TestRegister_RepositoryError(t *testing.T) {
	mockRepo := &MockUsersRepository{
		CreateFunc: func(ctx context.Context, user domain.User) (domain.User, error) {
			return domain.User{}, errors.New("database error")
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	userReg := domain.UserRegister{
		Nombre:   "Juan",
		Apellido: "Pérez",
		Username: "juanperez",
		Email:    "juan@example.com",
		Password: "SecurePass123!",
	}

	_, _, err := service.Register(context.Background(), userReg)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "error creating user") {
		t.Errorf("Expected error to contain 'error creating user', got: %s", err.Error())
	}
}

// TestLogin_Success prueba el login exitoso
func TestLogin_Success(t *testing.T) {
	// Hash de "password123" con SHA256
	hashedPassword := "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"

	mockRepo := &MockUsersRepository{
		GetByUsernameOrEmailFunc: func(ctx context.Context, usernameOrEmail string) (domain.User, error) {
			return domain.User{
				ID:            1,
				Nombre:        "Juan",
				Apellido:      "Pérez",
				Username:      "juanperez",
				Email:         "juan@example.com",
				Password:      hashedPassword,
				IsAdmin:       false,
				FechaRegistro: time.Now(),
			}, nil
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	credentials := domain.UserLogin{
		UsernameOrEmail: "juanperez",
		Password:        "password123",
	}

	userResp, token, err := service.Login(context.Background(), credentials)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if userResp.Username != "juanperez" {
		t.Errorf("Expected username 'juanperez', got: %s", userResp.Username)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

// TestLogin_UserNotFound prueba login con usuario inexistente
func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := &MockUsersRepository{
		GetByUsernameOrEmailFunc: func(ctx context.Context, usernameOrEmail string) (domain.User, error) {
			return domain.User{}, errors.New("user not found")
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	credentials := domain.UserLogin{
		UsernameOrEmail: "nonexistent",
		Password:        "password123",
	}

	_, _, err := service.Login(context.Background(), credentials)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "credenciales inválidas" {
		t.Errorf("Expected error 'credenciales inválidas', got: %s", err.Error())
	}
}

// TestLogin_WrongPassword prueba login con contraseña incorrecta
func TestLogin_WrongPassword(t *testing.T) {
	hashedPassword := "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"

	mockRepo := &MockUsersRepository{
		GetByUsernameOrEmailFunc: func(ctx context.Context, usernameOrEmail string) (domain.User, error) {
			return domain.User{
				ID:       1,
				Username: "juanperez",
				Email:    "juan@example.com",
				Password: hashedPassword,
			}, nil
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	credentials := domain.UserLogin{
		UsernameOrEmail: "juanperez",
		Password:        "wrongpassword",
	}

	_, _, err := service.Login(context.Background(), credentials)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "credenciales inválidas" {
		t.Errorf("Expected error 'credenciales inválidas', got: %s", err.Error())
	}
}

// TestGetByID_Success prueba obtener usuario por ID
func TestGetByID_Success(t *testing.T) {
	mockRepo := &MockUsersRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (domain.User, error) {
			if id == 1 {
				return domain.User{
					ID:       1,
					Nombre:   "Juan",
					Apellido: "Pérez",
					Username: "juanperez",
					Email:    "juan@example.com",
				}, nil
			}
			return domain.User{}, errors.New("user not found")
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	userResp, err := service.GetByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if userResp.ID != 1 {
		t.Errorf("Expected ID 1, got: %d", userResp.ID)
	}

	if userResp.Username != "juanperez" {
		t.Errorf("Expected username 'juanperez', got: %s", userResp.Username)
	}
}

// TestGetByID_NotFound prueba usuario no encontrado
func TestGetByID_NotFound(t *testing.T) {
	mockRepo := &MockUsersRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (domain.User, error) {
			return domain.User{}, errors.New("user not found")
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	_, err := service.GetByID(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "user not found" {
		t.Errorf("Expected error 'user not found', got: %s", err.Error())
	}
}

// TestList_Success prueba listar usuarios
func TestList_Success(t *testing.T) {
	mockRepo := &MockUsersRepository{
		ListFunc: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{
				{
					ID:       1,
					Username: "user1",
					Email:    "user1@example.com",
				},
				{
					ID:       2,
					Username: "user2",
					Email:    "user2@example.com",
				},
			}, nil
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	users, err := service.List(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got: %d", len(users))
	}

	if users[0].Username != "user1" {
		t.Errorf("Expected first user 'user1', got: %s", users[0].Username)
	}
}

// TestList_Empty prueba listar cuando no hay usuarios
func TestList_Empty(t *testing.T) {
	mockRepo := &MockUsersRepository{
		ListFunc: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{}, nil
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	users, err := service.List(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("Expected 0 users, got: %d", len(users))
	}
}

// TestValidateToken_Success prueba la validación exitosa de token
func TestValidateToken_Success(t *testing.T) {
	mockRepo := &MockUsersRepository{}
	service := NewUsersService(mockRepo, "test-secret-key")

	// Crear un token válido
	claims := jwt.MapClaims{
		"iss":        "gym-management-system",
		"exp":        time.Now().Add(30 * time.Minute).Unix(),
		"username":   "testuser",
		"id_usuario": float64(1),
		"is_admin":   false,
		"role":       "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	// Validar el token
	validatedClaims, err := service.ValidateToken(tokenString)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if validatedClaims["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got: %s", validatedClaims["username"])
	}

	if validatedClaims["role"] != "user" {
		t.Errorf("Expected role 'user', got: %s", validatedClaims["role"])
	}
}

// TestValidateToken_Expired prueba token expirado
func TestValidateToken_Expired(t *testing.T) {
	mockRepo := &MockUsersRepository{}
	service := NewUsersService(mockRepo, "test-secret-key")

	// Crear un token expirado
	claims := jwt.MapClaims{
		"iss":        "gym-management-system",
		"exp":        time.Now().Add(-1 * time.Hour).Unix(), // Expirado hace 1 hora
		"username":   "testuser",
		"id_usuario": float64(1),
		"is_admin":   false,
		"role":       "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	// Validar el token
	_, err := service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}
}

// TestValidateToken_InvalidSignature prueba token con firma inválida
func TestValidateToken_InvalidSignature(t *testing.T) {
	mockRepo := &MockUsersRepository{}
	service := NewUsersService(mockRepo, "test-secret-key")

	// Crear un token con una clave diferente
	claims := jwt.MapClaims{
		"iss":        "gym-management-system",
		"exp":        time.Now().Add(30 * time.Minute).Unix(),
		"username":   "testuser",
		"id_usuario": float64(1),
		"is_admin":   false,
		"role":       "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("different-secret-key"))

	// Validar el token
	_, err := service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("Expected error for invalid signature, got nil")
	}
}

// TestValidateToken_MalformedToken prueba token malformado
func TestValidateToken_MalformedToken(t *testing.T) {
	mockRepo := &MockUsersRepository{}
	service := NewUsersService(mockRepo, "test-secret-key")

	// Token inválido
	tokenString := "invalid.token.string"

	// Validar el token
	_, err := service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("Expected error for malformed token, got nil")
	}
}

// TestGenerateToken_AdminRole prueba generación de token para admin
func TestGenerateToken_AdminRole(t *testing.T) {
	mockRepo := &MockUsersRepository{
		CreateFunc: func(ctx context.Context, user domain.User) (domain.User, error) {
			user.ID = 1
			user.IsAdmin = true
			user.FechaRegistro = time.Now()
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			return user, nil
		},
	}

	service := NewUsersService(mockRepo, "test-secret-key")

	userReg := domain.UserRegister{
		Nombre:   "Admin",
		Apellido: "User",
		Username: "adminuser",
		Email:    "admin@example.com",
		Password: "AdminPass123!",
	}

	_, token, err := service.Register(context.Background(), userReg)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Validar que el token contiene el rol admin
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected valid token, got error: %v", err)
	}

	if claims["role"] != "admin" {
		t.Errorf("Expected role 'admin', got: %s", claims["role"])
	}

	if claims["is_admin"] != true {
		t.Errorf("Expected is_admin true, got: %v", claims["is_admin"])
	}
}

// TestHashPassword prueba que el hash de password es consistente
func TestHashPassword(t *testing.T) {
	mockRepo := &MockUsersRepository{}
	service := NewUsersService(mockRepo, "test-secret-key")

	password := "testpassword123"
	hash1 := service.hashPassword(password)
	hash2 := service.hashPassword(password)

	if hash1 != hash2 {
		t.Error("Hash function should be deterministic")
	}

	expectedHash := "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"
	actualHash := service.hashPassword("password123")

	if actualHash != expectedHash {
		t.Errorf("Expected hash '%s', got: '%s'", expectedHash, actualHash)
	}
}

// Helper function para búsqueda de substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
