package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	repoMocks "github.com/yourusername/gym-management/subscriptions-api/internal/repository/mocks"
	serviceMocks "github.com/yourusername/gym-management/subscriptions-api/internal/services/mocks"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestPlanUpgradeFlow - Prueba el flujo completo de upgrade/cambio de plan
func TestPlanUpgradeFlow(t *testing.T) {
	t.Run("Upgrade de plan básico a premium exitoso", func(t *testing.T) {
		// Arrange
		userID := "user123"
		basicPlanID := primitive.NewObjectID()
		premiumPlanID := primitive.NewObjectID()
		currentSubscriptionID := primitive.NewObjectID()

		// Plan actual (Básico)
		basicPlan := &entities.Plan{
			ID:            basicPlanID,
			Nombre:        "Plan Básico",
			PrecioMensual: 15000.0,
			DuracionDias:  30,
			Activo:        true,
			TipoAcceso:    "limitado",
		}

		// Plan nuevo (Premium)
		premiumPlan := &entities.Plan{
			ID:            premiumPlanID,
			Nombre:        "Plan Premium",
			PrecioMensual: 25000.0,
			DuracionDias:  30,
			Activo:        true,
			TipoAcceso:    "completo",
		}

		// Suscripción actual
		currentSubscription := &entities.Subscription{
			ID:               currentSubscriptionID,
			UsuarioID:        userID,
			PlanID:           basicPlanID,
			Estado:           "activa",
			FechaInicio:      time.Now().Add(-10 * 24 * time.Hour), // 10 días atrás
			FechaVencimiento: time.Now().Add(20 * 24 * time.Hour),  // 20 días restantes
		}

		cancelCalled := false
		createCalled := false
		eventsPublished := []string{}

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			// Encontrar suscripción activa
			FindActiveByUserIDFunc: func(ctx context.Context, userID string) (*entities.Subscription, error) {
				if cancelCalled {
					// Después de cancelar, no hay suscripción activa
					return nil, errors.New("no hay suscripción activa")
				}
				return currentSubscription, nil
			},
			// Cancelar suscripción actual
			UpdateStatusFunc: func(ctx context.Context, id primitive.ObjectID, status, pagoID string) error {
				if id == currentSubscriptionID && status == "cancelada" {
					cancelCalled = true
					currentSubscription.Estado = "cancelada"
					return nil
				}
				return errors.New("error al actualizar estado")
			},
			// Crear nueva suscripción
			CreateFunc: func(ctx context.Context, subscription *entities.Subscription) error {
				if subscription.PlanID == premiumPlanID {
					createCalled = true
					subscription.ID = primitive.NewObjectID()
					return nil
				}
				return errors.New("plan incorrecto")
			},
		}

		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				if id == basicPlanID {
					return basicPlan, nil
				}
				if id == premiumPlanID {
					return premiumPlan, nil
				}
				return nil, errors.New("plan no encontrado")
			},
		}

		mockUserValidator := &serviceMocks.MockUserValidator{
			ValidateUserFunc: func(ctx context.Context, userID string) (bool, error) {
				return true, nil
			},
		}

		mockEventPublisher := &serviceMocks.MockEventPublisher{
			PublishSubscriptionEventFunc: func(action, subscriptionID string, data map[string]interface{}) error {
				eventsPublished = append(eventsPublished, action)
				return nil
			},
		}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		// Act - PASO 1: Verificar que tiene suscripción activa
		activeSub, err := service.GetActiveSubscriptionByUserID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Error al obtener suscripción activa: %v", err)
		}
		if activeSub.PlanID != basicPlanID.Hex() {
			t.Errorf("Plan actual esperado %s, obtenido %s", basicPlanID.Hex(), activeSub.PlanID)
		}

		// Act - PASO 2: Cancelar suscripción actual
		err = service.CancelSubscription(context.Background(), currentSubscriptionID.Hex())
		if err != nil {
			t.Fatalf("Error al cancelar suscripción: %v", err)
		}

		// Assert - Verificar que se canceló
		if !cancelCalled {
			t.Error("Se esperaba que se cancelara la suscripción")
		}

		// Act - PASO 3: Crear nueva suscripción con plan premium
		newSubReq := dtos.CreateSubscriptionRequest{
			UsuarioID:        userID,
			PlanID:           premiumPlanID.Hex(),
			SucursalOrigenID: "sucursal1",
			MetodoPago:       "cash",
			AutoRenovacion:   false,
		}

		newSub, err := service.CreateSubscription(context.Background(), newSubReq)

		// Assert - Verificar nueva suscripción
		if err != nil {
			t.Fatalf("Error al crear nueva suscripción: %v", err)
		}
		if !createCalled {
			t.Error("Se esperaba que se creara una nueva suscripción")
		}
		if newSub.PlanID != premiumPlanID.Hex() {
			t.Errorf("Plan de nueva suscripción esperado %s, obtenido %s", premiumPlanID.Hex(), newSub.PlanID)
		}
		if newSub.PlanNombre != "Plan Premium" {
			t.Errorf("Nombre del plan esperado 'Plan Premium', obtenido '%s'", newSub.PlanNombre)
		}

		// Assert - Verificar eventos publicados
		if len(eventsPublished) < 2 {
			t.Errorf("Se esperaban al menos 2 eventos, se publicaron %d", len(eventsPublished))
		}

		// Verificar que se publicó evento de cancelación
		foundDeleteEvent := false
		for _, event := range eventsPublished {
			if event == "delete" {
				foundDeleteEvent = true
				break
			}
		}
		if !foundDeleteEvent {
			t.Error("Se esperaba evento 'delete' para la cancelación")
		}
	})

	t.Run("Error al intentar upgrade sin cancelar suscripción anterior", func(t *testing.T) {
		// Arrange
		userID := "user123"
		basicPlanID := primitive.NewObjectID()
		premiumPlanID := primitive.NewObjectID()

		basicPlan := &entities.Plan{
			ID:     basicPlanID,
			Nombre: "Plan Básico",
			Activo: true,
		}

		premiumPlan := &entities.Plan{
			ID:     premiumPlanID,
			Nombre: "Plan Premium",
			Activo: true,
		}

		// Suscripción activa existente
		activeSubscription := &entities.Subscription{
			ID:        primitive.NewObjectID(),
			UsuarioID: userID,
			PlanID:    basicPlanID,
			Estado:    "activa",
		}

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			FindActiveByUserIDFunc: func(ctx context.Context, userID string) (*entities.Subscription, error) {
				return activeSubscription, nil
			},
			CreateFunc: func(ctx context.Context, subscription *entities.Subscription) error {
				// No debería llegar aquí si hay validación
				return errors.New("ya tienes una suscripción activa")
			},
		}

		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				if id == basicPlanID {
					return basicPlan, nil
				}
				if id == premiumPlanID {
					return premiumPlan, nil
				}
				return nil, errors.New("plan no encontrado")
			},
		}

		mockUserValidator := &serviceMocks.MockUserValidator{
			ValidateUserFunc: func(ctx context.Context, userID string) (bool, error) {
				return true, nil
			},
		}

		mockEventPublisher := &serviceMocks.MockEventPublisher{
			PublishSubscriptionEventFunc: func(action, subscriptionID string, data map[string]interface{}) error {
				return nil
			},
		}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		// Act - Intentar crear nueva suscripción sin cancelar la anterior
		newSubReq := dtos.CreateSubscriptionRequest{
			UsuarioID:  userID,
			PlanID:     premiumPlanID.Hex(),
			MetodoPago: "cash",
		}

		_, err := service.CreateSubscription(context.Background(), newSubReq)

		// Assert - Debe fallar porque ya tiene suscripción activa
		if err == nil {
			t.Error("Se esperaba un error al intentar crear suscripción con una activa existente")
		}
	})

	t.Run("Downgrade de plan premium a básico", func(t *testing.T) {
		// Arrange
		userID := "user456"
		basicPlanID := primitive.NewObjectID()
		premiumPlanID := primitive.NewObjectID()
		currentSubscriptionID := primitive.NewObjectID()

		basicPlan := &entities.Plan{
			ID:            basicPlanID,
			Nombre:        "Plan Básico",
			PrecioMensual: 15000.0,
			Activo:        true,
		}

		premiumPlan := &entities.Plan{
			ID:            premiumPlanID,
			Nombre:        "Plan Premium",
			PrecioMensual: 25000.0,
			Activo:        true,
		}

		currentSubscription := &entities.Subscription{
			ID:        currentSubscriptionID,
			UsuarioID: userID,
			PlanID:    premiumPlanID,
			Estado:    "activa",
		}

		cancelled := false
		created := false

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			FindActiveByUserIDFunc: func(ctx context.Context, userID string) (*entities.Subscription, error) {
				if cancelled {
					return nil, errors.New("no hay suscripción activa")
				}
				return currentSubscription, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id primitive.ObjectID, status, pagoID string) error {
				if status == "cancelada" {
					cancelled = true
					return nil
				}
				return errors.New("estado inválido")
			},
			CreateFunc: func(ctx context.Context, subscription *entities.Subscription) error {
				if subscription.PlanID == basicPlanID {
					created = true
					subscription.ID = primitive.NewObjectID()
					return nil
				}
				return errors.New("plan incorrecto")
			},
		}

		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				if id == basicPlanID {
					return basicPlan, nil
				}
				if id == premiumPlanID {
					return premiumPlan, nil
				}
				return nil, errors.New("plan no encontrado")
			},
		}

		mockUserValidator := &serviceMocks.MockUserValidator{
			ValidateUserFunc: func(ctx context.Context, userID string) (bool, error) {
				return true, nil
			},
		}

		mockEventPublisher := &serviceMocks.MockEventPublisher{
			PublishSubscriptionEventFunc: func(action, subscriptionID string, data map[string]interface{}) error {
				return nil
			},
		}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		// Act - Cancelar premium
		err := service.CancelSubscription(context.Background(), currentSubscriptionID.Hex())
		if err != nil {
			t.Fatalf("Error al cancelar: %v", err)
		}

		// Act - Crear básico
		newSubReq := dtos.CreateSubscriptionRequest{
			UsuarioID:  userID,
			PlanID:     basicPlanID.Hex(),
			MetodoPago: "cash",
		}

		newSub, err := service.CreateSubscription(context.Background(), newSubReq)

		// Assert
		if err != nil {
			t.Fatalf("Error al crear nueva suscripción: %v", err)
		}
		if !cancelled || !created {
			t.Error("No se completó el proceso de downgrade")
		}
		if newSub.PlanNombre != "Plan Básico" {
			t.Errorf("Se esperaba 'Plan Básico', obtenido '%s'", newSub.PlanNombre)
		}
	})
}
