package controllers

import (
	"fmt"
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

	// Obtener información del usuario autenticado
	currentUserID, _ := ctx.Get("id_usuario")
	isAdmin, _ := ctx.Get("is_admin")

	payment, err := c.service.GetPaymentByID(ctx.Request.Context(), paymentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// 🔒 VALIDACIÓN DE SEGURIDAD: Solo el dueño del pago o admin puede verlo
	if payment.UserID != fmt.Sprintf("%d", currentUserID.(uint)) && !isAdmin.(bool) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have permission to view this payment",
		})
		return
	}

	ctx.JSON(http.StatusOK, payment)
}

// GetAllPayments obtiene todos los pagos
func (c *PaymentController) GetAllPayments(ctx *gin.Context) {
	payments, err := c.service.GetAllPayments(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

// GetPaymentsByUser obtiene todos los pagos de un usuario
func (c *PaymentController) GetPaymentsByUser(ctx *gin.Context) {
	requestedUserID := ctx.Param("user_id")

	// Obtener información del usuario autenticado
	currentUserID, _ := ctx.Get("id_usuario")
	isAdmin, _ := ctx.Get("is_admin")

	// 🔒 VALIDACIÓN DE SEGURIDAD: Solo puedes ver tus propios pagos (o admin puede ver cualquiera)
	if requestedUserID != fmt.Sprintf("%d", currentUserID.(uint)) && !isAdmin.(bool) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "You can only view your own payments",
		})
		return
	}

	payments, err := c.service.GetPaymentsByUser(ctx.Request.Context(), requestedUserID)
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

// GetPaymentsByStatus obtiene pagos por estado (solo admin)
func (c *PaymentController) GetPaymentsByStatus(ctx *gin.Context) {
	// Obtener status del parámetro de ruta
	status := ctx.Param("status")

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

// ApproveCashPayment aprueba un pago en efectivo (solo admin)
func (c *PaymentController) ApproveCashPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")

	// Actualizar el estado a "completed"
	req := dtos.UpdatePaymentStatusRequest{
		Status: "completed",
	}

	err := c.service.UpdatePaymentStatus(ctx.Request.Context(), paymentID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Pago en efectivo aprobado correctamente"})
}

// RejectCashPayment rechaza un pago en efectivo (solo admin)
func (c *PaymentController) RejectCashPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")

	// Actualizar el estado a "failed"
	updateReq := dtos.UpdatePaymentStatusRequest{
		Status: "failed",
	}

	err := c.service.UpdatePaymentStatus(ctx.Request.Context(), paymentID, updateReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Pago en efectivo rechazado"})
}

// HealthCheck verifica el estado del servicio
func (c *PaymentController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "payments-api",
	})
}
