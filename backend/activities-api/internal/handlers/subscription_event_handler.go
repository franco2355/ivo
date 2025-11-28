package handlers

import (
	"activities-api/internal/services"
	"context"
	"fmt"
	"log"
)

// SubscriptionEventHandler maneja eventos de suscripciones
type SubscriptionEventHandler struct {
	inscripcionesService services.InscripcionesService
}

// NewSubscriptionEventHandler crea un nuevo handler de eventos de suscripciones
func NewSubscriptionEventHandler(inscripcionesService services.InscripcionesService) *SubscriptionEventHandler {
	return &SubscriptionEventHandler{
		inscripcionesService: inscripcionesService,
	}
}

// HandleSubscriptionCancelled maneja el evento de cancelación de suscripción
func (h *SubscriptionEventHandler) HandleSubscriptionCancelled(ctx context.Context, usuarioID uint) error {
	log.Printf("🔔 [SubscriptionEventHandler] Suscripción cancelada para usuario %d - Desinscribiendo de actividades...\n", usuarioID)

	// Desinscribir al usuario de todas las actividades
	count, err := h.inscripcionesService.DeactivateAllByUser(ctx, usuarioID)
	if err != nil {
		return fmt.Errorf("error desinscribiendo usuario %d: %w", usuarioID, err)
	}

	log.Printf("✅ [SubscriptionEventHandler] Usuario %d desinscrito de %d actividades\n", usuarioID, count)
	return nil
}
