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
		// Determinar código de estado según el error
		statusCode := http.StatusInternalServerError
		if err.Error() == "username or email already exists" {
			statusCode = http.StatusConflict
		} else if err.Error() == "nombre is required" ||
			err.Error() == "apellido is required" ||
			err.Error() == "username is required" ||
			err.Error() == "email is required" ||
			err.Error() == "password is required" ||
			contains(err.Error(), "must") || contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to register user",
			"details": err.Error(),
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
