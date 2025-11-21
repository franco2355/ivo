package controllers

import (
	"activities-api/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InscripcionesController maneja las peticiones HTTP relacionadas con inscripciones
// Migrado de backend/controllers/inscripcion/incripcion_controller.go con dependency injection
type InscripcionesController struct {
	service services.InscripcionesService
}

// NewInscripcionesController crea una nueva instancia del controller
func NewInscripcionesController(service services.InscripcionesService) *InscripcionesController {
	return &InscripcionesController{
		service: service,
	}
}

// List obtiene todas las inscripciones del usuario autenticado
// GET /inscripciones (requiere JWT)
// Migrado de backend/controllers/inscripcion/incripcion_controller.go:14
func (c *InscripcionesController) List(ctx *gin.Context) {
	// Obtener el ID del usuario del contexto (seteado por middleware JWT)
	userID, exists := ctx.Get("id_usuario")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Usuario no autenticado"})
		return
	}

	inscripciones, err := c.service.ListByUser(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la consulta"})
		return
	}

	ctx.JSON(http.StatusOK, inscripciones)
}

// Create inscribe al usuario autenticado en una actividad
// POST /inscripciones {"actividad_id": 1} (requiere JWT)
// Migrado de backend/controllers/inscripcion/incripcion_controller.go:31
func (c *InscripcionesController) Create(ctx *gin.Context) {
	// Obtener el ID del usuario del contexto (seteado por middleware JWT)
	userID, exists := ctx.Get("id_usuario")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// Parsear el body
	var inscripcionCreate struct {
		ActividadID uint `json:"actividad_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&inscripcionCreate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos con formato incorrecto", "details": err.Error()})
		return
	}

	// Obtener el token de autorización
	authHeader := ctx.GetHeader("Authorization")

	createdInscripcion, err := c.service.Create(ctx.Request.Context(), userID.(uint), inscripcionCreate.ActividadID, authHeader)
	if err != nil {
		errString := strings.ToLower(err.Error())

		// Detectar errores específicos del hook BeforeCreate
		if strings.Contains(errString, "ya está inscripto") || strings.Contains(errString, "ya esta inscripto") {
			ctx.JSON(http.StatusConflict, gin.H{"error": "El usuario ya está inscripto a esta actividad"})
		} else if strings.Contains(errString, "cupo de la actividad ha sido alcanzado") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No se puede inscribir, el cupo de la actividad ha sido alcanzado"})
		} else if strings.Contains(errString, "actividad no encontrada") || strings.Contains(errString, "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "La actividad no existe"})
		} else if strings.Contains(errString, "no tiene suscripción activa") {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Debe tener una suscripción activa para inscribirse"})
		} else if strings.Contains(errString, "requiere plan premium") {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Esta actividad requiere un plan premium"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al inscribir el usuario", "details": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, createdInscripcion)
}

// Deactivate desinscribe al usuario autenticado de una actividad
// DELETE /inscripciones {"actividad_id": 1} (requiere JWT)
// Migrado de backend/controllers/inscripcion/incripcion_controller.go:66
func (c *InscripcionesController) Deactivate(ctx *gin.Context) {
	// Obtener el ID del usuario del contexto (seteado por middleware JWT)
	userID, exists := ctx.Get("id_usuario")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// Parsear el body
	var deactivateRequest struct {
		ActividadID uint `json:"actividad_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&deactivateRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos con formato incorrecto", "details": err.Error()})
		return
	}

	err := c.service.Deactivate(ctx.Request.Context(), userID.(uint), deactivateRequest.ActividadID)
	if err != nil {
		errString := strings.ToLower(err.Error())

		if strings.Contains(errString, "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Inscripción no encontrada"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al desinscribir al usuario", "details": err.Error()})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
