package controllers

import (
	"net/http"
	"strconv"
	"users-api/internal/domain"
	"users-api/internal/services"

	"github.com/gin-gonic/gin"
)

// UsersController maneja las peticiones HTTP para usuarios
type UsersController struct {
	service services.UsersService
}

// NewUsersController crea una nueva instancia del controller
// Dependency Injection: recibe el service como parámetro
func NewUsersController(usersService services.UsersService) *UsersController {
	return &UsersController{
		service: usersService,
	}
}

// Register maneja POST /register - Registra un nuevo usuario
// @Summary Registra un nuevo usuario
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.UserRegister true "Datos del usuario"
// @Success 201 {object} map[string]interface{} "user, token"
// @Failure 400 {object} map[string]interface{} "error, details"
// @Failure 500 {object} map[string]interface{} "error, details"
// @Router /register [post]
func (c *UsersController) Register(ctx *gin.Context) {
	var userReg domain.UserRegister

	// Parsear JSON del body
	if err := ctx.ShouldBindJSON(&userReg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// Llamar al service
	user, token, err := c.service.Register(ctx.Request.Context(), userReg)
	if err != nil {
		// Determinar código de estado y mensaje según el error
		statusCode := http.StatusInternalServerError
		errorMessage := "Error al registrar usuario"
		errorCode := err.Error()

		// Errores de validación (400)
		if contains(errorCode, "required") ||
			contains(errorCode, "must") ||
			contains(errorCode, "invalid") ||
			contains(errorCode, "can only contain") {
			statusCode = http.StatusBadRequest
			errorMessage = translateValidationError(errorCode)
		} else if contains(errorCode, "already_exists") {
			// Errores de duplicados (409)
			statusCode = http.StatusConflict
			if errorCode == "username_already_exists" {
				errorMessage = "El nombre de usuario ya está en uso"
			} else if errorCode == "email_already_exists" {
				errorMessage = "El email ya está registrado"
			} else {
				errorMessage = "El usuario o email ya existe"
			}
		} else if contains(errorCode, "creating user") {
			// Error de base de datos (500)
			statusCode = http.StatusInternalServerError
			errorMessage = "Error del servidor. Por favor, intenta más tarde"
		}

		ctx.JSON(statusCode, gin.H{
			"error":   errorMessage,
			"details": errorCode,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

// Login maneja POST /login - Autentica un usuario
// @Summary Login de usuario
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body domain.UserLogin true "Credenciales"
// @Success 200 {object} map[string]interface{} "user, token"
// @Failure 400 {object} map[string]interface{} "error, details"
// @Failure 401 {object} map[string]interface{} "error, details"
// @Router /login [post]
func (c *UsersController) Login(ctx *gin.Context) {
	var credentials domain.UserLogin

	// Parsear JSON del body
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// Llamar al service
	user, token, err := c.service.Login(ctx.Request.Context(), credentials)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid credentials",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetByID maneja GET /users/:id - Obtiene un usuario por ID
// @Summary Obtiene un usuario por ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} domain.UserResponse
// @Failure 400 {object} map[string]interface{} "error, details"
// @Failure 404 {object} map[string]interface{} "error, details"
// @Router /users/{id} [get]
func (c *UsersController) GetByID(ctx *gin.Context) {
	// Extraer ID del path param
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "ID must be a positive integer",
		})
		return
	}

	// Llamar al service
	user, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to get user",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// List maneja GET /users - Lista todos los usuarios
// @Summary Lista todos los usuarios
// @Tags users
// @Produce json
// @Success 200 {array} domain.UserResponse
// @Failure 500 {object} map[string]interface{} "error, details"
// @Router /users [get]
func (c *UsersController) List(ctx *gin.Context) {
	// Llamar al service
	users, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list users",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// Helper functions
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

// translateValidationError traduce errores de validación al español
func translateValidationError(errorCode string) string {
	errorTranslations := map[string]string{
		"nombre is required":                                             "El nombre es obligatorio",
		"apellido is required":                                           "El apellido es obligatorio",
		"username is required":                                           "El nombre de usuario es obligatorio",
		"email is required":                                              "El email es obligatorio",
		"password is required":                                           "La contraseña es obligatoria",
		"nombre must be at most 30 characters":                           "El nombre debe tener máximo 30 caracteres",
		"apellido must be at most 30 characters":                         "El apellido debe tener máximo 30 caracteres",
		"username must be between 3 and 30 characters":                   "El nombre de usuario debe tener entre 3 y 30 caracteres",
		"username can only contain letters, numbers, hyphens and underscores": "El nombre de usuario solo puede contener letras, números, guiones y guiones bajos",
		"invalid email format":                                           "Formato de email inválido",
		"password must be at least 8 characters":                         "La contraseña debe tener al menos 8 caracteres",
		"password must contain at least one uppercase letter":            "La contraseña debe contener al menos una letra mayúscula",
		"password must contain at least one lowercase letter":            "La contraseña debe contener al menos una letra minúscula",
		"password must contain at least one number":                      "La contraseña debe contener al menos un número",
	}

	if translation, ok := errorTranslations[errorCode]; ok {
		return translation
	}
	return errorCode
}
