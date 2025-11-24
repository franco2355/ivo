package services

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"github.com/yourusername/gym-management/subscriptions-api/internal/repository/mocks"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPlanService_CreatePlan(t *testing.T) {
	t.Run("Crear plan exitosamente", func(t *testing.T) {
		// Arrange
		mockRepo := &mocks.MockPlanRepository{
			CreateFunc: func(ctx context.Context, plan *entities.Plan) error {
				plan.ID = primitive.NewObjectID()
				return nil
			},
		}
		service := NewPlanService(mockRepo)

		req := dtos.CreatePlanRequest{
			Nombre:                "Plan Test",
			Descripcion:           "Descripción test",
			PrecioMensual:         100.0,
			TipoAcceso:            "completo",
			DuracionDias:          30,
			Activo:                true,
			ActividadesPermitidas: []string{"gym", "pool"},
		}

		// Act
		result, err := service.CreatePlan(context.Background(), req)

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if result == nil {
			t.Error("Se esperaba un resultado, pero fue nil")
		}
		if result.Nombre != req.Nombre {
			t.Errorf("Nombre esperado %s, obtenido %s", req.Nombre, result.Nombre)
		}
		if result.PrecioMensual != req.PrecioMensual {
			t.Errorf("Precio esperado %.2f, obtenido %.2f", req.PrecioMensual, result.PrecioMensual)
		}
	})

	t.Run("Error al crear plan en repositorio", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("error de base de datos")
		mockRepo := &mocks.MockPlanRepository{
			CreateFunc: func(ctx context.Context, plan *entities.Plan) error {
				return expectedError
			},
		}
		service := NewPlanService(mockRepo)

		req := dtos.CreatePlanRequest{
			Nombre:        "Plan Test",
			PrecioMensual: 100.0,
			TipoAcceso:    "completo",
			DuracionDias:  30,
		}

		// Act
		result, err := service.CreatePlan(context.Background(), req)

		// Assert
		if err == nil {
			t.Error("Se esperaba un error, pero no se obtuvo ninguno")
		}
		if result != nil {
			t.Error("No se esperaba resultado, pero se obtuvo uno")
		}
	})
}

func TestPlanService_GetPlanByID(t *testing.T) {
	t.Run("Obtener plan existente", func(t *testing.T) {
		// Arrange
		planID := primitive.NewObjectID()
		expectedPlan := &entities.Plan{
			ID:            planID,
			Nombre:        "Plan Premium",
			PrecioMensual: 150.0,
			TipoAcceso:    "completo",
			DuracionDias:  30,
			Activo:        true,
		}

		mockRepo := &mocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				if id == planID {
					return expectedPlan, nil
				}
				return nil, errors.New("plan no encontrado")
			},
		}
		service := NewPlanService(mockRepo)

		// Act
		result, err := service.GetPlanByID(context.Background(), planID.Hex())

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if result == nil {
			t.Fatal("Se esperaba un resultado, pero fue nil")
		}
		if result.ID != planID.Hex() {
			t.Errorf("ID esperado %s, obtenido %s", planID.Hex(), result.ID)
		}
		if result.Nombre != expectedPlan.Nombre {
			t.Errorf("Nombre esperado %s, obtenido %s", expectedPlan.Nombre, result.Nombre)
		}
	})

	t.Run("Error con ID inválido", func(t *testing.T) {
		// Arrange
		mockRepo := &mocks.MockPlanRepository{}
		service := NewPlanService(mockRepo)

		// Act
		result, err := service.GetPlanByID(context.Background(), "invalid-id")

		// Assert
		if err == nil {
			t.Error("Se esperaba un error por ID inválido")
		}
		if result != nil {
			t.Error("No se esperaba resultado con ID inválido")
		}
	})

	t.Run("Plan no encontrado", func(t *testing.T) {
		// Arrange
		mockRepo := &mocks.MockPlanRepository{
			FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Plan, error) {
				return nil, errors.New("plan no encontrado")
			},
		}
		service := NewPlanService(mockRepo)

		// Act
		result, err := service.GetPlanByID(context.Background(), primitive.NewObjectID().Hex())

		// Assert
		if err == nil {
			t.Error("Se esperaba un error")
		}
		if result != nil {
			t.Error("No se esperaba resultado")
		}
	})
}

func TestPlanService_ListPlans(t *testing.T) {
	t.Run("Listar planes exitosamente", func(t *testing.T) {
		// Arrange
		mockPlans := []*entities.Plan{
			{
				ID:            primitive.NewObjectID(),
				Nombre:        "Plan Basic",
				PrecioMensual: 50.0,
				Activo:        true,
			},
			{
				ID:            primitive.NewObjectID(),
				Nombre:        "Plan Premium",
				PrecioMensual: 100.0,
				Activo:        true,
			},
		}

		mockRepo := &mocks.MockPlanRepository{
			FindAllPaginatedFunc: func(ctx context.Context, filters map[string]interface{}, page, pageSize int64, sortBy string, sortDesc bool) ([]*entities.Plan, error) {
				return mockPlans, nil
			},
			CountFunc: func(ctx context.Context, filters map[string]interface{}) (int64, error) {
				return int64(len(mockPlans)), nil
			},
		}
		service := NewPlanService(mockRepo)

		activo := true
		query := dtos.ListPlansQuery{
			Activo:   &activo,
			Page:     1,
			PageSize: 10,
		}

		// Act
		result, err := service.ListPlans(context.Background(), query)

		// Assert
		if err != nil {
			t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
		}
		if result == nil {
			t.Fatal("Se esperaba un resultado, pero fue nil")
		}
		if len(result.Plans) != len(mockPlans) {
			t.Errorf("Se esperaban %d planes, pero se obtuvieron %d", len(mockPlans), len(result.Plans))
		}
		if result.Total != len(mockPlans) {
			t.Errorf("Total esperado %d, obtenido %d", len(mockPlans), result.Total)
		}
	})
}
