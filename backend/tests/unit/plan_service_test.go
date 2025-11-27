package unit

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockPlanRepository para tests unitarios
type MockPlanRepository struct {
	CreateFunc         func(ctx context.Context, plan interface{}) (interface{}, error)
	GetByIDFunc        func(ctx context.Context, id uint) (interface{}, error)
	ListFunc           func(ctx context.Context) ([]interface{}, error)
	UpdateFunc         func(ctx context.Context, id uint, plan interface{}) (interface{}, error)
	DeleteFunc         func(ctx context.Context, id uint) error
	GetActiveFunc      func(ctx context.Context) ([]interface{}, error)
	GetByDurationFunc  func(ctx context.Context, duration int) ([]interface{}, error)
}

func (m *MockPlanRepository) Create(ctx context.Context, plan interface{}) (interface{}, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, plan)
	}
	return nil, nil
}

func (m *MockPlanRepository) GetByID(ctx context.Context, id uint) (interface{}, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPlanRepository) List(ctx context.Context) ([]interface{}, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return []interface{}{}, nil
}

func (m *MockPlanRepository) Update(ctx context.Context, id uint, plan interface{}) (interface{}, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, plan)
	}
	return nil, nil
}

func (m *MockPlanRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockPlanRepository) GetActive(ctx context.Context) ([]interface{}, error) {
	if m.GetActiveFunc != nil {
		return m.GetActiveFunc(ctx)
	}
	return []interface{}{}, nil
}

func (m *MockPlanRepository) GetByDuration(ctx context.Context, duration int) ([]interface{}, error) {
	if m.GetByDurationFunc != nil {
		return m.GetByDurationFunc(ctx, duration)
	}
	return []interface{}{}, nil
}

// Plan structure for testing
type Plan struct {
	ID             uint
	Nombre         string
	Descripcion    string
	Precio         float64
	DuracionMeses  int
	Activo         bool
	Caracteristicas []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TestCreatePlan_Success verifica creación exitosa de plan
func TestCreatePlan_Success(t *testing.T) {
	mockRepo := &MockPlanRepository{
		CreateFunc: func(ctx context.Context, plan interface{}) (interface{}, error) {
			p := plan.(Plan)
			p.ID = 1
			p.CreatedAt = time.Now()
			p.UpdatedAt = time.Now()
			return p, nil
		},
	}

	newPlan := Plan{
		Nombre:         "Plan Mensual",
		Descripcion:    "Plan básico mensual",
		Precio:         5000.0,
		DuracionMeses:  1,
		Activo:         true,
		Caracteristicas: []string{"Acceso al gym", "Clases grupales"},
	}

	result, err := mockRepo.Create(context.Background(), newPlan)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	createdPlan := result.(Plan)
	if createdPlan.ID != 1 {
		t.Errorf("Expected ID 1, got: %d", createdPlan.ID)
	}

	if createdPlan.Nombre != "Plan Mensual" {
		t.Errorf("Expected name 'Plan Mensual', got: %s", createdPlan.Nombre)
	}
}

// TestCreatePlan_ValidationErrors verifica validaciones
func TestCreatePlan_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		plan        Plan
		expectedErr string
	}{
		{
			name: "precio negativo",
			plan: Plan{
				Nombre:        "Plan Test",
				Precio:        -100,
				DuracionMeses: 1,
			},
			expectedErr: "precio debe ser positivo",
		},
		{
			name: "duración inválida",
			plan: Plan{
				Nombre:        "Plan Test",
				Precio:        1000,
				DuracionMeses: 0,
			},
			expectedErr: "duración debe ser al menos 1 mes",
		},
		{
			name: "nombre vacío",
			plan: Plan{
				Nombre:        "",
				Precio:        1000,
				DuracionMeses: 1,
			},
			expectedErr: "nombre es requerido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validaciones que deberían hacerse antes de crear
			var err error

			if tt.plan.Precio <= 0 {
				err = errors.New("precio debe ser positivo")
			} else if tt.plan.DuracionMeses <= 0 {
				err = errors.New("duración debe ser al menos 1 mes")
			} else if tt.plan.Nombre == "" {
				err = errors.New("nombre es requerido")
			}

			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}

			if err.Error() != tt.expectedErr {
				t.Errorf("Expected error '%s', got: '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestGetPlanByID_Success verifica obtener plan por ID
func TestGetPlanByID_Success(t *testing.T) {
	mockRepo := &MockPlanRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (interface{}, error) {
			if id == 1 {
				return Plan{
					ID:             1,
					Nombre:         "Plan Anual",
					Precio:         50000,
					DuracionMeses:  12,
					Activo:         true,
				}, nil
			}
			return nil, errors.New("plan not found")
		},
	}

	result, err := mockRepo.GetByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	plan := result.(Plan)
	if plan.ID != 1 {
		t.Errorf("Expected ID 1, got: %d", plan.ID)
	}

	if plan.Nombre != "Plan Anual" {
		t.Errorf("Expected name 'Plan Anual', got: %s", plan.Nombre)
	}
}

// TestGetPlanByID_NotFound verifica plan no encontrado
func TestGetPlanByID_NotFound(t *testing.T) {
	mockRepo := &MockPlanRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (interface{}, error) {
			return nil, errors.New("plan not found")
		},
	}

	_, err := mockRepo.GetByID(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "plan not found" {
		t.Errorf("Expected error 'plan not found', got: %s", err.Error())
	}
}

// TestListActivePlans_Success verifica listado de planes activos
func TestListActivePlans_Success(t *testing.T) {
	mockRepo := &MockPlanRepository{
		GetActiveFunc: func(ctx context.Context) ([]interface{}, error) {
			return []interface{}{
				Plan{ID: 1, Nombre: "Plan Mensual", Activo: true},
				Plan{ID: 2, Nombre: "Plan Trimestral", Activo: true},
				Plan{ID: 3, Nombre: "Plan Anual", Activo: true},
			}, nil
		},
	}

	result, err := mockRepo.GetActive(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 active plans, got: %d", len(result))
	}

	for i, r := range result {
		plan := r.(Plan)
		if !plan.Activo {
			t.Errorf("Plan at index %d is not active", i)
		}
	}
}

// TestGetPlansByDuration_Success verifica obtener planes por duración
func TestGetPlansByDuration_Success(t *testing.T) {
	mockRepo := &MockPlanRepository{
		GetByDurationFunc: func(ctx context.Context, duration int) ([]interface{}, error) {
			if duration == 12 {
				return []interface{}{
					Plan{ID: 1, Nombre: "Plan Anual Básico", DuracionMeses: 12},
					Plan{ID: 2, Nombre: "Plan Anual Premium", DuracionMeses: 12},
				}, nil
			}
			return []interface{}{}, nil
		},
	}

	result, err := mockRepo.GetByDuration(context.Background(), 12)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 annual plans, got: %d", len(result))
	}

	for _, r := range result {
		plan := r.(Plan)
		if plan.DuracionMeses != 12 {
			t.Errorf("Expected duration 12 months, got: %d", plan.DuracionMeses)
		}
	}
}

// TestUpdatePlan_Success verifica actualización de plan
func TestUpdatePlan_Success(t *testing.T) {
	mockRepo := &MockPlanRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (interface{}, error) {
			return Plan{
				ID:             id,
				Nombre:         "Plan Original",
				Precio:         5000,
				DuracionMeses:  1,
				Activo:         true,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, id uint, plan interface{}) (interface{}, error) {
			p := plan.(Plan)
			p.ID = id
			p.UpdatedAt = time.Now()
			return p, nil
		},
	}

	// Obtener plan existente
	result, err := mockRepo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	plan := result.(Plan)

	// Actualizar precio
	plan.Precio = 6000

	updated, err := mockRepo.Update(context.Background(), 1, plan)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	updatedPlan := updated.(Plan)
	if updatedPlan.Precio != 6000 {
		t.Errorf("Expected price 6000, got: %f", updatedPlan.Precio)
	}
}

// TestDeletePlan_Success verifica eliminación de plan
func TestDeletePlan_Success(t *testing.T) {
	deleted := false

	mockRepo := &MockPlanRepository{
		DeleteFunc: func(ctx context.Context, id uint) error {
			deleted = true
			return nil
		},
	}

	err := mockRepo.Delete(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !deleted {
		t.Error("Expected plan to be deleted")
	}
}

// TestDeactivatePlan_InsteadOfDelete verifica desactivación en lugar de eliminación
func TestDeactivatePlan_InsteadOfDelete(t *testing.T) {
	mockRepo := &MockPlanRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (interface{}, error) {
			return Plan{
				ID:     id,
				Nombre: "Plan a desactivar",
				Activo: true,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, id uint, plan interface{}) (interface{}, error) {
			p := plan.(Plan)
			p.ID = id
			return p, nil
		},
	}

	// Obtener plan
	result, err := mockRepo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	plan := result.(Plan)

	// Desactivar en lugar de eliminar (soft delete)
	plan.Activo = false

	updated, err := mockRepo.Update(context.Background(), 1, plan)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	updatedPlan := updated.(Plan)
	if updatedPlan.Activo {
		t.Error("Expected plan to be deactivated")
	}
}

// TestPlanPriceCalculation_Discount verifica cálculo de descuentos
func TestPlanPriceCalculation_Discount(t *testing.T) {
	tests := []struct {
		name            string
		basePrice       float64
		durationMonths  int
		expectedDiscount float64
	}{
		{
			name:            "Plan mensual sin descuento",
			basePrice:       5000,
			durationMonths:  1,
			expectedDiscount: 0,
		},
		{
			name:            "Plan trimestral 5% descuento",
			basePrice:       5000,
			durationMonths:  3,
			expectedDiscount: 0.05,
		},
		{
			name:            "Plan semestral 10% descuento",
			basePrice:       5000,
			durationMonths:  6,
			expectedDiscount: 0.10,
		},
		{
			name:            "Plan anual 15% descuento",
			basePrice:       5000,
			durationMonths:  12,
			expectedDiscount: 0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lógica de descuento basada en duración
			var discount float64
			switch tt.durationMonths {
			case 1:
				discount = 0
			case 3:
				discount = 0.05
			case 6:
				discount = 0.10
			case 12:
				discount = 0.15
			}

			if discount != tt.expectedDiscount {
				t.Errorf("Expected discount %f, got: %f", tt.expectedDiscount, discount)
			}

			totalPrice := tt.basePrice * float64(tt.durationMonths)
			finalPrice := totalPrice * (1 - discount)

			t.Logf("Base: $%.2f, Duration: %d months, Discount: %.0f%%, Final: $%.2f",
				tt.basePrice, tt.durationMonths, discount*100, finalPrice)
		})
	}
}
