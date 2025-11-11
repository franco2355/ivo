package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/services"
)

// SubscriptionController - Controlador HTTP para suscripciones
type SubscriptionController struct {
	subscriptionService *services.SubscriptionService // DI
	healthService       *services.HealthService       // DI
}

// NewSubscriptionController - Constructor con DI
func NewSubscriptionController(subscriptionService *services.SubscriptionService, healthService *services.HealthService) *SubscriptionController {
	return &SubscriptionController{
		subscriptionService: subscriptionService,
		healthService:       healthService,
	}
}

// CreateSubscription - POST /subscriptions
func (c *SubscriptionController) CreateSubscription(ctx *gin.Context) {
	var req dtos.CreateSubscriptionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := c.subscriptionService.CreateSubscription(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, subscription)
}

// GetSubscription - GET /subscriptions/:id
func (c *SubscriptionController) GetSubscription(ctx *gin.Context) {
	id := ctx.Param("id")

	subscription, err := c.subscriptionService.GetSubscriptionByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, subscription)
}

// GetActiveSubscriptionByUser - GET /subscriptions/active/:user_id
func (c *SubscriptionController) GetActiveSubscriptionByUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")

	subscription, err := c.subscriptionService.GetActiveSubscriptionByUserID(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, subscription)
}

// UpdateSubscriptionStatus - PATCH /subscriptions/:id/status
func (c *SubscriptionController) UpdateSubscriptionStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dtos.UpdateSubscriptionStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.subscriptionService.UpdateSubscriptionStatus(ctx.Request.Context(), id, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Estado actualizado correctamente"})
}

// CancelSubscription - DELETE /subscriptions/:id
func (c *SubscriptionController) CancelSubscription(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.subscriptionService.CancelSubscription(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Suscripción cancelada correctamente"})
}

// HealthCheck - GET /healthz
func (c *SubscriptionController) HealthCheck(ctx *gin.Context) {
	healthStatus := c.healthService.CheckHealth(ctx.Request.Context())

	// Determinar código HTTP según el estado
	statusCode := http.StatusOK
	if healthStatus.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	} else if healthStatus.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	ctx.JSON(statusCode, healthStatus)
}
