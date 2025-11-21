package controllers

import (
	"activities-api/internal/domain"
	"activities-api/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ActividadesController maneja las peticiones HTTP relacionadas con actividades
// Migrado de backend/controllers/actividad/actividad_controller.go con dependency injection
type ActividadesController struct {
	service services.ActividadesService
}

// NewActividadesController crea una nueva instancia del controller
func NewActividadesController(service services.ActividadesService) *ActividadesController {
	return &ActividadesController{
		service: service,
	}
}

// List obtiene todas las actividades
// GET /actividades
// Migrado de backend/controllers/actividad/actividad_controller.go:29
func (c *ActividadesController) List(ctx *gin.Context) {
	actividades, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar actividades"})
		return
	}

	ctx.JSON(http.StatusOK, actividades)
}

// Search busca actividades por parámetros
// GET /actividades/buscar?id=1&titulo=yoga&horario=10:00&categoria=fitness
// Migrado de backend/controllers/actividad/actividad_controller.go:15
func (c *ActividadesController) Search(ctx *gin.Context) {
	params := map[string]interface{}{
		"id":        ctx.Query("id"),
		"titulo":    ctx.Query("titulo"),
		"horario":   ctx.Query("horario"),
		"categoria": ctx.Query("categoria"),
	}

	actividades, err := c.service.Search(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar actividades"})
		return
	}

	ctx.JSON(http.StatusOK, actividades)
}

// GetByID obtiene una actividad por ID
// GET /actividades/:id
// Migrado de backend/controllers/actividad/actividad_controller.go:39
func (c *ActividadesController) GetByID(ctx *gin.Context) {
	idActividad, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El id debe ser un número"})
		return
	}

	actividad, err := c.service.GetByID(ctx.Request.Context(), uint(idActividad))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "La actividad no existe"})
		return
	}

	ctx.JSON(http.StatusOK, actividad)
}

// Create crea una nueva actividad
// POST /actividades (admin only)
// Migrado de backend/controllers/actividad/actividad_controller.go:56
func (c *ActividadesController) Create(ctx *gin.Context) {
	var actividadCreate domain.ActividadCreate
	if err := ctx.ShouldBindJSON(&actividadCreate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos con formato incorrecto", "details": err.Error()})
		return
	}

	createdActividad, err := c.service.Create(ctx.Request.Context(), actividadCreate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear la actividad", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdActividad)
}

// Update actualiza una actividad existente
// PUT /actividades/:id (admin only)
// Migrado de backend/controllers/actividad/actividad_controller.go:73
func (c *ActividadesController) Update(ctx *gin.Context) {
	idActividad, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El id debe ser un número"})
		return
	}

	var actividadUpdate domain.ActividadUpdate
	if err := ctx.ShouldBindJSON(&actividadUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos con formato incorrecto", "details": err.Error()})
		return
	}

	updatedActividad, err := c.service.Update(ctx.Request.Context(), uint(idActividad), actividadUpdate)
	if err != nil {
		errString := err.Error()

		// Detectar errores específicos del hook BeforeUpdate
		if strings.Contains(errString, "inscripciones activas que superan el nuevo límite") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if strings.Contains(errString, "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Actividad no encontrada"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, updatedActividad)
}

// Delete elimina una actividad
// DELETE /actividades/:id (admin only)
// Migrado de backend/controllers/actividad/actividad_controller.go:109
func (c *ActividadesController) Delete(ctx *gin.Context) {
	idActividad, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El id debe ser un número"})
		return
	}

	err = c.service.Delete(ctx.Request.Context(), uint(idActividad))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Actividad no encontrada"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
