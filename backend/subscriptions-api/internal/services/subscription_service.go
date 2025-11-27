package services

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"github.com/yourusername/gym-management/subscriptions-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionService - Servicio de lógica de negocio para suscripciones
type SubscriptionService struct {
	subscriptionRepo repository.SubscriptionRepository // DI
	planRepo         repository.PlanRepository         // DI
	userService      UserValidator                     // DI (Interface para validar usuarios)
	eventPublisher   EventPublisher                    // DI (Interface para publicar eventos)
}

// UserValidator - Interface para validar usuarios (abstrae users-api)
type UserValidator interface {
	ValidateUser(ctx context.Context, userID string) (bool, error)
}

// EventPublisher - Interface para publicar eventos (abstrae RabbitMQ)
type EventPublisher interface {
	PublishSubscriptionEvent(action, subscriptionID string, data map[string]interface{}) error
}

// NewSubscriptionService - Constructor con DI
func NewSubscriptionService(
	subscriptionRepo repository.SubscriptionRepository,
	planRepo repository.PlanRepository,
	userService UserValidator,
	eventPublisher EventPublisher,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
		userService:      userService,
		eventPublisher:   eventPublisher,
	}
}

// CreateSubscription - Crea una nueva suscripción
func (s *SubscriptionService) CreateSubscription(ctx context.Context, req dtos.CreateSubscriptionRequest) (*dtos.SubscriptionResponse, error) {
	// 1. Validar usuario existe
	valid, err := s.userService.ValidateUser(ctx, req.UsuarioID)
	if err != nil || !valid {
		return nil, fmt.Errorf("usuario no válido: %w", err)
	}

	// 2. Verificar si el usuario ya tiene una suscripción activa
	existingSubscription, err := s.subscriptionRepo.FindActiveByUserID(ctx, req.UsuarioID)
	if err == nil && existingSubscription != nil {
		// Encontró una suscripción activa, no permitir crear otra
		return nil, fmt.Errorf("ya tienes una suscripción activa. No puedes crear otra hasta que expire o sea cancelada")
	}

	// Verificar si hay suscripciones pendientes de pago
	filters := map[string]interface{}{
		"usuario_id": req.UsuarioID,
		"estado":     "pendiente_pago",
	}
	pendingSubscriptions, errPending := s.subscriptionRepo.FindAll(ctx, filters)
	if errPending == nil && len(pendingSubscriptions) > 0 {
		return nil, fmt.Errorf("ya tienes una suscripción pendiente de pago. Por favor completa el pago o cancélala antes de crear una nueva")
	}

	// 3. Obtener plan
	planObjID, err := primitive.ObjectIDFromHex(req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("ID de plan inválido")
	}

	plan, err := s.planRepo.FindByID(ctx, planObjID)
	if err != nil {
		return nil, fmt.Errorf("plan no encontrado: %w", err)
	}

	if !plan.Activo {
		return nil, fmt.Errorf("el plan no está activo")
	}

	// 4. Calcular fechas
	now := time.Now()
	fechaVencimiento := now.AddDate(0, 0, plan.DuracionDias)

	// 5. Crear suscripción
	subscription := &entities.Subscription{
		ID:               primitive.NewObjectID(),
		UsuarioID:        req.UsuarioID,
		PlanID:           planObjID,
		SucursalOrigenID: req.SucursalOrigenID,
		FechaInicio:      now,
		FechaVencimiento: fechaVencimiento,
		Estado:           "pendiente_pago",
		Metadata: entities.Metadata{
			MetodoPagoPreferido: req.MetodoPago,
			AutoRenovacion:      req.AutoRenovacion,
			Notas:               req.Notas,
		},
		HistorialRenovaciones: []entities.Renovacion{},
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	// 6. Guardar en repositorio
	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, err
	}

	// 7. Publicar evento
	eventData := map[string]interface{}{
		"usuario_id": subscription.UsuarioID,
		"plan_id":    subscription.PlanID.Hex(),
		"estado":     subscription.Estado,
	}
	s.eventPublisher.PublishSubscriptionEvent("create", subscription.ID.Hex(), eventData)

	// 8. Mapear a DTO de respuesta
	return s.mapSubscriptionToResponse(subscription, plan.Nombre), nil
}

// GetSubscriptionByID - Obtiene una suscripción por ID
func (s *SubscriptionService) GetSubscriptionByID(ctx context.Context, id string) (*dtos.SubscriptionResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID inválido")
	}

	subscription, err := s.subscriptionRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Enriquecer con nombre del plan
	plan, _ := s.planRepo.FindByID(ctx, subscription.PlanID)
	planNombre := ""
	if plan != nil {
		planNombre = plan.Nombre
	}

	return s.mapSubscriptionToResponse(subscription, planNombre), nil
}

// GetActiveSubscriptionByUserID - Obtiene la suscripción activa de un usuario
func (s *SubscriptionService) GetActiveSubscriptionByUserID(ctx context.Context, userID string) (*dtos.SubscriptionResponse, error) {
	subscription, err := s.subscriptionRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Enriquecer con nombre del plan
	plan, _ := s.planRepo.FindByID(ctx, subscription.PlanID)
	planNombre := ""
	if plan != nil {
		planNombre = plan.Nombre
	}

	return s.mapSubscriptionToResponse(subscription, planNombre), nil
}

// GetSubscriptionsByUserID - Obtiene todas las suscripciones de un usuario
func (s *SubscriptionService) GetSubscriptionsByUserID(ctx context.Context, userID string) ([]*dtos.SubscriptionResponse, error) {
	filters := map[string]interface{}{
		"usuario_id": userID,
	}

	fmt.Printf("🔍 [GetSubscriptionsByUserID] Buscando suscripciones para usuario: %s\n", userID)
	subscriptions, err := s.subscriptionRepo.FindAll(ctx, filters)
	if err != nil {
		fmt.Printf("❌ [GetSubscriptionsByUserID] Error: %v\n", err)
		return nil, err
	}
	fmt.Printf("✅ [GetSubscriptionsByUserID] Encontradas %d suscripciones\n", len(subscriptions))

	var responses []*dtos.SubscriptionResponse
	for _, subscription := range subscriptions {
		// Enriquecer con nombre del plan
		plan, _ := s.planRepo.FindByID(ctx, subscription.PlanID)
		planNombre := ""
		if plan != nil {
			planNombre = plan.Nombre
		}

		responses = append(responses, s.mapSubscriptionToResponse(subscription, planNombre))
	}

	return responses, nil
}

// UpdateSubscriptionStatus - Actualiza el estado de una suscripción
func (s *SubscriptionService) UpdateSubscriptionStatus(ctx context.Context, id string, req dtos.UpdateSubscriptionStatusRequest) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID inválido")
	}

	if err := s.subscriptionRepo.UpdateStatus(ctx, objID, req.Estado, req.PagoID); err != nil {
		return err
	}

	// Publicar evento
	eventData := map[string]interface{}{
		"estado":  req.Estado,
		"pago_id": req.PagoID,
	}
	s.eventPublisher.PublishSubscriptionEvent("update", id, eventData)

	return nil
}

// CancelSubscription - Cancela una suscripción
func (s *SubscriptionService) CancelSubscription(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID inválido")
	}

	if err := s.subscriptionRepo.UpdateStatus(ctx, objID, "cancelada", ""); err != nil {
		return err
	}

	// Publicar evento
	s.eventPublisher.PublishSubscriptionEvent("delete", id, nil)

	return nil
}

// mapSubscriptionToResponse - Helper para mapear entidad a DTO
func (s *SubscriptionService) mapSubscriptionToResponse(subscription *entities.Subscription, planNombre string) *dtos.SubscriptionResponse {
	var renovaciones []dtos.RenovacionResponse
	for _, r := range subscription.HistorialRenovaciones {
		renovaciones = append(renovaciones, dtos.RenovacionResponse{
			Fecha:  r.Fecha,
			PagoID: r.PagoID,
			Monto:  r.Monto,
		})
	}

	return &dtos.SubscriptionResponse{
		ID:                    subscription.ID.Hex(),
		UsuarioID:             subscription.UsuarioID,
		PlanID:                subscription.PlanID.Hex(),
		PlanNombre:            planNombre,
		SucursalOrigenID:      subscription.SucursalOrigenID,
		FechaInicio:           subscription.FechaInicio,
		FechaVencimiento:      subscription.FechaVencimiento,
		Estado:                subscription.Estado,
		PagoID:                subscription.PagoID,
		AutoRenovacion:        subscription.Metadata.AutoRenovacion,
		MetodoPagoPreferido:   subscription.Metadata.MetodoPagoPreferido,
		Notas:                 subscription.Metadata.Notas,
		HistorialRenovaciones: renovaciones,
		CreatedAt:             subscription.CreatedAt,
		UpdatedAt:             subscription.UpdatedAt,
	}
}

// ============================================================================
// MÉTODOS PARA PROCESAR EVENTOS DE PAGOS (Invocados por PaymentEventHandler)
// ============================================================================

// ActivateSubscriptionByPayment activa una suscripción cuando se completa un pago
// Llamado desde PaymentEventHandler al recibir evento payment.completed
func (s *SubscriptionService) ActivateSubscriptionByPayment(ctx context.Context, subscriptionID string, paymentID string) error {
	// Validar ID
	objID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return fmt.Errorf("ID de suscripción inválido: %w", err)
	}

	// Obtener suscripción
	subscription, err := s.subscriptionRepo.FindByID(ctx, objID)
	if err != nil {
		return fmt.Errorf("suscripción no encontrada: %w", err)
	}

	// Validar que esté en estado pendiente_pago
	if subscription.Estado != "pendiente_pago" {
		return fmt.Errorf("suscripción no está en estado pendiente_pago (actual: %s)", subscription.Estado)
	}

	// Actualizar estado a "activa" y registrar pago
	now := time.Now()
	subscription.Estado = "activa"
	subscription.PagoID = paymentID
	subscription.FechaInicio = now // Actualizar fecha inicio al momento de activación
	subscription.UpdatedAt = now

	// Recalcular fecha de vencimiento desde ahora
	plan, err := s.planRepo.FindByID(ctx, subscription.PlanID)
	if err == nil && plan != nil {
		subscription.FechaVencimiento = now.AddDate(0, 0, plan.DuracionDias)
	}

	// Guardar cambios
	if err := s.subscriptionRepo.Update(ctx, objID, subscription); err != nil {
		return fmt.Errorf("error actualizando suscripción: %w", err)
	}

	// Publicar evento de activación
	eventData := map[string]interface{}{
		"usuario_id": subscription.UsuarioID,
		"plan_id":    subscription.PlanID.Hex(),
		"payment_id": paymentID,
		"estado":     "activa",
	}
	s.eventPublisher.PublishSubscriptionEvent("activated", subscriptionID, eventData)

	return nil
}

// RegisterPaymentFailure registra un fallo de pago para una suscripción
// Llamado desde PaymentEventHandler al recibir evento payment.failed
func (s *SubscriptionService) RegisterPaymentFailure(ctx context.Context, subscriptionID string, paymentID string) error {
	// Validar ID
	objID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return fmt.Errorf("ID de suscripción inválido: %w", err)
	}

	// Obtener suscripción
	subscription, err := s.subscriptionRepo.FindByID(ctx, objID)
	if err != nil {
		return fmt.Errorf("suscripción no encontrada: %w", err)
	}

	// Registrar el fallo en metadata (opcional)
	if subscription.Metadata.Notas == "" {
		subscription.Metadata.Notas = fmt.Sprintf("Pago fallido: %s", paymentID)
	} else {
		subscription.Metadata.Notas += fmt.Sprintf("\nPago fallido: %s", paymentID)
	}

	subscription.UpdatedAt = time.Now()

	// Guardar cambios
	if err := s.subscriptionRepo.Update(ctx, objID, subscription); err != nil {
		return fmt.Errorf("error registrando fallo de pago: %w", err)
	}

	// Publicar evento de fallo
	eventData := map[string]interface{}{
		"usuario_id": subscription.UsuarioID,
		"payment_id": paymentID,
		"estado":     subscription.Estado,
	}
	s.eventPublisher.PublishSubscriptionEvent("payment_failed", subscriptionID, eventData)

	return nil
}

// CancelSubscriptionByRefund cancela una suscripción cuando se reembolsa el pago
// Llamado desde PaymentEventHandler al recibir evento payment.refunded
func (s *SubscriptionService) CancelSubscriptionByRefund(ctx context.Context, subscriptionID string, paymentID string) error {
	// Validar ID
	objID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return fmt.Errorf("ID de suscripción inválido: %w", err)
	}

	// Obtener suscripción
	subscription, err := s.subscriptionRepo.FindByID(ctx, objID)
	if err != nil {
		return fmt.Errorf("suscripción no encontrada: %w", err)
	}

	// Cancelar la suscripción
	now := time.Now()
	subscription.Estado = "cancelada"
	subscription.Metadata.Notas += fmt.Sprintf("\nReembolso procesado: %s (Fecha: %s)", paymentID, now.Format("2006-01-02"))
	subscription.UpdatedAt = now

	// Guardar cambios
	if err := s.subscriptionRepo.Update(ctx, objID, subscription); err != nil {
		return fmt.Errorf("error cancelando suscripción por reembolso: %w", err)
	}

	// Publicar evento de cancelación por reembolso
	eventData := map[string]interface{}{
		"usuario_id": subscription.UsuarioID,
		"payment_id": paymentID,
		"estado":     "cancelada",
		"motivo":     "refund",
	}
	s.eventPublisher.PublishSubscriptionEvent("cancelled_by_refund", subscriptionID, eventData)

	return nil
}

// ============================================================================
// MÉTODOS PARA EXPIRACIÓN AUTOMÁTICA DE SUSCRIPCIONES
// ============================================================================

// ExpireOverdueSubscriptions - Marca como expiradas las suscripciones vencidas
// Este método debería ejecutarse periódicamente (ej: cada hora con un cron job)
func (s *SubscriptionService) ExpireOverdueSubscriptions(ctx context.Context) (int, error) {
	// Buscar suscripciones activas con fecha de vencimiento pasada
	expiredSubscriptions, err := s.subscriptionRepo.FindExpiredSubscriptions(ctx)
	if err != nil {
		return 0, fmt.Errorf("error buscando suscripciones vencidas: %w", err)
	}

	count := 0
	for _, subscription := range expiredSubscriptions {
		// Actualizar estado a "expirada"
		err := s.subscriptionRepo.UpdateStatus(ctx, subscription.ID, "expirada", subscription.PagoID)
		if err != nil {
			fmt.Printf("⚠️ Error al expirar suscripción %s: %v\n", subscription.ID.Hex(), err)
			continue
		}

		// Publicar evento de expiración
		eventData := map[string]interface{}{
			"usuario_id":         subscription.UsuarioID,
			"plan_id":            subscription.PlanID.Hex(),
			"fecha_vencimiento":  subscription.FechaVencimiento,
		}
		s.eventPublisher.PublishSubscriptionEvent("expired", subscription.ID.Hex(), eventData)

		count++
	}

	if count > 0 {
		fmt.Printf("✅ Se expiraron %d suscripciones vencidas\n", count)
	}

	return count, nil
}
