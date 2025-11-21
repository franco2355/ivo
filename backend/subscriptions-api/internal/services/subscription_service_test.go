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

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	t.Run("Crear suscripción exitosamente", func(t *testing.T) {
		// Arrange
		planID := primitive.NewObjectID()
		mockPlan := &entities.Plan{
			ID:            planID,
			Nombre:        "Plan Premium",
			PrecioMensual: 100.0,
			DuracionDias:  30,
			Activo:        true,
		}

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			CreateFunc: func(ctx context.Context, subscription *entities.Subscription) error {
				subscription.ID = primitive.NewObjectID()
				return nil
			},
		}

		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				if id == planID {
					return mockPlan, nil
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

		req := dtos.CreateSubscriptionRequest{
			UsuarioID:        "user123",
			PlanID:           planID.Hex(),
			SucursalOrigenID: "sucursal1",
			MetodoPago:       "credit_card",
			AutoRenovacion:   true,
		}

		// Act
		result, err := service.CreateSubscription(context.Background(), req)

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if result == nil {
			t.Fatal("Se esperaba un resultado, pero fue nil")
		}
		if result.UsuarioID != req.UsuarioID {
			t.Errorf("UsuarioID esperado %s, obtenido %s", req.UsuarioID, result.UsuarioID)
		}
		if result.Estado != "pendiente_pago" {
			t.Errorf("Estado esperado 'pendiente_pago', obtenido '%s'", result.Estado)
		}
		if result.PlanNombre != mockPlan.Nombre {
			t.Errorf("Nombre del plan esperado %s, obtenido %s", mockPlan.Nombre, result.PlanNombre)
		}
	})

	t.Run("Error cuando usuario no es válido", func(t *testing.T) {
		// Arrange
		mockSubRepo := &repoMocks.MockSubscriptionRepository{}
		mockPlanRepo := &repoMocks.MockPlanRepository{}
		mockUserValidator := &serviceMocks.MockUserValidator{
			ValidateUserFunc: func(ctx context.Context, userID string) (bool, error) {
				return false, errors.New("usuario no encontrado")
			},
		}
		mockEventPublisher := &serviceMocks.MockEventPublisher{}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		req := dtos.CreateSubscriptionRequest{
			UsuarioID:  "invalid_user",
			PlanID:     primitive.NewObjectID().Hex(),
			MetodoPago: "credit_card",
		}

		// Act
		result, err := service.CreateSubscription(context.Background(), req)

		// Assert
		if err == nil {
			t.Error("Se esperaba un error por usuario inválido")
		}
		if result != nil {
			t.Error("No se esperaba resultado con usuario inválido")
		}
	})

	t.Run("Error cuando plan no existe", func(t *testing.T) {
		// Arrange
		mockSubRepo := &repoMocks.MockSubscriptionRepository{}
		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				return nil, errors.New("plan no encontrado")
			},
		}
		mockUserValidator := &serviceMocks.MockUserValidator{
			ValidateUserFunc: func(ctx context.Context, userID string) (bool, error) {
				return true, nil
			},
		}
		mockEventPublisher := &serviceMocks.MockEventPublisher{}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		req := dtos.CreateSubscriptionRequest{
			UsuarioID:  "user123",
			PlanID:     primitive.NewObjectID().Hex(),
			MetodoPago: "credit_card",
		}

		// Act
		result, err := service.CreateSubscription(context.Background(), req)

		// Assert
		if err == nil {
			t.Error("Se esperaba un error por plan no encontrado")
		}
		if result != nil {
			t.Error("No se esperaba resultado con plan inexistente")
		}
	})

	t.Run("Error cuando plan no está activo", func(t *testing.T) {
		// Arrange
		planID := primitive.NewObjectID()
		mockPlan := &entities.Plan{
			ID:           planID,
			Nombre:       "Plan Inactivo",
			DuracionDias: 30,
			Activo:       false, // Plan inactivo
		}

		mockSubRepo := &repoMocks.MockSubscriptionRepository{}
		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				return mockPlan, nil
			},
		}
		mockUserValidator := &serviceMocks.MockUserValidator{
			ValidateUserFunc: func(ctx context.Context, userID string) (bool, error) {
				return true, nil
			},
		}
		mockEventPublisher := &serviceMocks.MockEventPublisher{}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		req := dtos.CreateSubscriptionRequest{
			UsuarioID:  "user123",
			PlanID:     planID.Hex(),
			MetodoPago: "credit_card",
		}

		// Act
		result, err := service.CreateSubscription(context.Background(), req)

		// Assert
		if err == nil {
			t.Error("Se esperaba un error por plan inactivo")
		}
		if result != nil {
			t.Error("No se esperaba resultado con plan inactivo")
		}
	})
}

func TestSubscriptionService_GetActiveSubscriptionByUserID(t *testing.T) {
	t.Run("Obtener suscripción activa exitosamente", func(t *testing.T) {
		// Arrange
		planID := primitive.NewObjectID()
		subscriptionID := primitive.NewObjectID()

		mockSubscription := &entities.Subscription{
			ID:               subscriptionID,
			UsuarioID:        "user123",
			PlanID:           planID,
			Estado:           "activa",
			FechaInicio:      time.Now(),
			FechaVencimiento: time.Now().Add(30 * 24 * time.Hour),
		}

		mockPlan := &entities.Plan{
			ID:     planID,
			Nombre: "Plan Premium",
		}

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			FindActiveByUserIDFunc: func(ctx context.Context, userID string) (*entities.Subscription, error) {
				if userID == "user123" {
					return mockSubscription, nil
				}
				return nil, errors.New("no hay suscripción activa")
			},
		}

		mockPlanRepo := &repoMocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				return mockPlan, nil
			},
		}

		mockUserValidator := &serviceMocks.MockUserValidator{}
		mockEventPublisher := &serviceMocks.MockEventPublisher{}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		// Act
		result, err := service.GetActiveSubscriptionByUserID(context.Background(), "user123")

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if result == nil {
			t.Fatal("Se esperaba un resultado, pero fue nil")
		}
		if result.Estado != "activa" {
			t.Errorf("Estado esperado 'activa', obtenido '%s'", result.Estado)
		}
		if result.PlanNombre != mockPlan.Nombre {
			t.Errorf("Nombre del plan esperado %s, obtenido %s", mockPlan.Nombre, result.PlanNombre)
		}
	})

	t.Run("No hay suscripción activa", func(t *testing.T) {
		// Arrange
		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			FindActiveByUserIDFunc: func(ctx context.Context, userID string) (*entities.Subscription, error) {
				return nil, errors.New("no hay suscripción activa")
			},
		}

		mockPlanRepo := &repoMocks.MockPlanRepository{}
		mockUserValidator := &serviceMocks.MockUserValidator{}
		mockEventPublisher := &serviceMocks.MockEventPublisher{}

		service := NewSubscriptionService(mockSubRepo, mockPlanRepo, mockUserValidator, mockEventPublisher)

		// Act
		result, err := service.GetActiveSubscriptionByUserID(context.Background(), "user_sin_suscripcion")

		// Assert
		if err == nil {
			t.Error("Se esperaba un error")
		}
		if result != nil {
			t.Error("No se esperaba resultado")
		}
	})
}

func TestSubscriptionService_UpdateSubscriptionStatus(t *testing.T) {
	t.Run("Actualizar estado exitosamente", func(t *testing.T) {
		// Arrange
		subscriptionID := primitive.NewObjectID()

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			UpdateStatusFunc: func(ctx context.Context, id primitive.ObjectID, status, pagoID string) error {
				if id == subscriptionID && status == "activa" {
					return nil
				}
				return errors.New("error al actualizar")
			},
		}

		eventCalled := false
		mockEventPublisher := &serviceMocks.MockEventPublisher{
			PublishSubscriptionEventFunc: func(action, subscriptionID string, data map[string]interface{}) error {
				eventCalled = true
				return nil
			},
		}

		service := NewSubscriptionService(mockSubRepo, nil, nil, mockEventPublisher)

		req := dtos.UpdateSubscriptionStatusRequest{
			Estado: "activa",
			PagoID: "payment123",
		}

		// Act
		err := service.UpdateSubscriptionStatus(context.Background(), subscriptionID.Hex(), req)

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if !eventCalled {
			t.Error("Se esperaba que se publicara un evento")
		}
	})

	t.Run("Error con ID inválido", func(t *testing.T) {
		// Arrange
		service := NewSubscriptionService(nil, nil, nil, nil)

		req := dtos.UpdateSubscriptionStatusRequest{
			Estado: "activa",
		}

		// Act
		err := service.UpdateSubscriptionStatus(context.Background(), "invalid-id", req)

		// Assert
		if err == nil {
			t.Error("Se esperaba un error por ID inválido")
		}
	})
}

func TestSubscriptionService_CancelSubscription(t *testing.T) {
	t.Run("Cancelar suscripción exitosamente", func(t *testing.T) {
		// Arrange
		subscriptionID := primitive.NewObjectID()

		mockSubRepo := &repoMocks.MockSubscriptionRepository{
			UpdateStatusFunc: func(ctx context.Context, id primitive.ObjectID, status, pagoID string) error {
				if id == subscriptionID && status == "cancelada" {
					return nil
				}
				return errors.New("error al cancelar")
			},
		}

		eventCalled := false
		mockEventPublisher := &serviceMocks.MockEventPublisher{
			PublishSubscriptionEventFunc: func(action, subscriptionID string, data map[string]interface{}) error {
				if action == "delete" {
					eventCalled = true
				}
				return nil
			},
		}

		service := NewSubscriptionService(mockSubRepo, nil, nil, mockEventPublisher)

		// Act
		err := service.CancelSubscription(context.Background(), subscriptionID.Hex())

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if !eventCalled {
			t.Error("Se esperaba que se publicara un evento de tipo 'delete'")
		}
	})
}
