package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/payments-api/internal/domain/dtos"
	"github.com/yourusername/payments-api/internal/services"
)

// PaymentController - Controlador HTTP para pagos con DI
type PaymentController struct {
	service *services.PaymentService
}

// NewPaymentController - Constructor con Dependency Injection
func NewPaymentController(service *services.PaymentService) *PaymentController {
	return &PaymentController{
		service: service,
	}
}

// CreatePayment crea un nuevo pago
func (c *PaymentController) CreatePayment(ctx *gin.Context) {
	var req dtos.CreatePaymentRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := c.service.CreatePayment(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, payment)
}

// GetPayment obtiene un pago por ID
func (c *PaymentController) GetPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")

	payment, err := c.service.GetPaymentByID(ctx.Request.Context(), paymentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payment)
}

// GetPaymentsByUser obtiene todos los pagos de un usuario
func (c *PaymentController) GetPaymentsByUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")

	payments, err := c.service.GetPaymentsByUser(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

// GetPaymentsByEntity obtiene pagos asociados a una entidad
func (c *PaymentController) GetPaymentsByEntity(ctx *gin.Context) {
	entityType := ctx.Query("entity_type")
	entityID := ctx.Query("entity_id")

	if entityType == "" || entityID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "entity_type y entity_id son requeridos"})
		return
	}

	payments, err := c.service.GetPaymentsByEntity(ctx.Request.Context(), entityType, entityID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

// GetPaymentsByStatus obtiene pagos por estado
func (c *PaymentController) GetPaymentsByStatus(ctx *gin.Context) {
	status := ctx.Query("status")

	if status == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status es requerido"})
		return
	}

	payments, err := c.service.GetPaymentsByStatus(ctx.Request.Context(), status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

// UpdatePaymentStatus actualiza el estado de un pago
func (c *PaymentController) UpdatePaymentStatus(ctx *gin.Context) {
	paymentID := ctx.Param("id")

	var req dtos.UpdatePaymentStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.service.UpdatePaymentStatus(ctx.Request.Context(), paymentID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Estado actualizado correctamente"})
}

// ProcessPayment procesa un pago
func (c *PaymentController) ProcessPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")

	err := c.service.ProcessPayment(ctx.Request.Context(), paymentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Pago procesado correctamente"})
}

// HealthCheck verifica el estado del servicio
func (c *PaymentController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "payments-api",
	})
}
