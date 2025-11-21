package services

import (
	"activities-api/internal/domain"
	"context"
	"errors"
	"testing"
	"time"
)

// --- Manual Mocks ---

type MockActividadesRepository struct {
	ListFunc        func(ctx context.Context) ([]domain.Actividad, error)
	GetByIDFunc     func(ctx context.Context, id uint) (domain.Actividad, error)
	GetByParamsFunc func(ctx context.Context, params map[string]interface{}) ([]domain.Actividad, error)
	CreateFunc      func(ctx context.Context, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error)
	UpdateFunc      func(ctx context.Context, id uint, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error)
	DeleteFunc      func(ctx context.Context, id uint) error
}

func (m *MockActividadesRepository) List(ctx context.Context) ([]domain.Actividad, error) {
	return m.ListFunc(ctx)
}
func (m *MockActividadesRepository) GetByID(ctx context.Context, id uint) (domain.Actividad, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockActividadesRepository) GetByParams(ctx context.Context, params map[string]interface{}) ([]domain.Actividad, error) {
	return m.GetByParamsFunc(ctx, params)
}
func (m *MockActividadesRepository) Create(ctx context.Context, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error) {
	return m.CreateFunc(ctx, actividad, horaInicio, horaFin)
}
func (m *MockActividadesRepository) Update(ctx context.Context, id uint, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error) {
	return m.UpdateFunc(ctx, id, actividad, horaInicio, horaFin)
}
func (m *MockActividadesRepository) Delete(ctx context.Context, id uint) error {
	return m.DeleteFunc(ctx, id)
}

type MockEventPublisher struct {
	PublishActivityEventFunc    func(action, activityID string, data map[string]interface{}) error
	PublishInscriptionEventFunc func(action, inscriptionID string, data map[string]interface{}) error
}

func (m *MockEventPublisher) PublishActivityEvent(action, activityID string, data map[string]interface{}) error {
	if m.PublishActivityEventFunc != nil {
		return m.PublishActivityEventFunc(action, activityID, data)
	}
	return nil
}
func (m *MockEventPublisher) PublishInscriptionEvent(action, inscriptionID string, data map[string]interface{}) error {
	if m.PublishInscriptionEventFunc != nil {
		return m.PublishInscriptionEventFunc(action, inscriptionID, data)
	}
	return nil
}

// --- Tests ---

func TestCreateActividad_Success(t *testing.T) {
	// Setup
	mockRepo := &MockActividadesRepository{}
	mockPublisher := &MockEventPublisher{}
	service := NewActividadesService(mockRepo, mockPublisher)

	input := domain.ActividadCreate{
		Titulo:        "Yoga",
		Descripcion:   "Clase de Yoga",
		Cupo:          20,
		Dia:           "Lunes",
		HorarioInicio: "10:00",
		HorarioFinal:  "11:00",
		FotoUrl:       "http://example.com/yoga.jpg",
		Instructor:    "Juan Perez",
		Categoria:     "Relax",
	}

	expectedID := uint(1)

	// Mock behaviors
	mockRepo.CreateFunc = func(ctx context.Context, actividad domain.Actividad, horaInicio, horaFin time.Time) (domain.Actividad, error) {
		// Verify basic mapping
		if actividad.Titulo != input.Titulo {
			t.Errorf("Expected Titulo %s, got %s", input.Titulo, actividad.Titulo)
		}
		// Return success
		actividad.ID = expectedID
		return actividad, nil
	}

	mockPublisher.PublishActivityEventFunc = func(action, activityID string, data map[string]interface{}) error {
		if action != "create" {
			t.Errorf("Expected action 'create', got %s", action)
		}
		if activityID != "1" {
			t.Errorf("Expected activityID '1', got %s", activityID)
		}
		return nil
	}

	// Execute
	result, err := service.Create(context.Background(), input)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != expectedID {
		t.Errorf("Expected ID %d, got %d", expectedID, result.ID)
	}
	if result.Titulo != input.Titulo {
		t.Errorf("Expected Titulo %s, got %s", input.Titulo, result.Titulo)
	}
}

func TestCreateActividad_ValidationError(t *testing.T) {
	service := NewActividadesService(&MockActividadesRepository{}, &MockEventPublisher{})

	// Case 1: Empty Title
	input := domain.ActividadCreate{
		Titulo:        "",
		Cupo:          10,
		Dia:           "Lunes",
		HorarioInicio: "10:00",
		HorarioFinal:  "11:00",
	}
	_, err := service.Create(context.Background(), input)
	if err == nil {
		t.Error("Expected error for empty title, got nil")
	}

	// Case 2: Invalid Time Range
	input2 := domain.ActividadCreate{
		Titulo:        "Test",
		Cupo:          10,
		Dia:           "Lunes",
		HorarioInicio: "11:00",
		HorarioFinal:  "10:00", // End before start
	}
	_, err = service.Create(context.Background(), input2)
	if err == nil {
		t.Error("Expected error for invalid time range, got nil")
	}
}

func TestGetActividad_Found(t *testing.T) {
	mockRepo := &MockActividadesRepository{}
	service := NewActividadesService(mockRepo, &MockEventPublisher{})

	expectedID := uint(1)
	expectedTitle := "Yoga"

	mockRepo.GetByIDFunc = func(ctx context.Context, id uint) (domain.Actividad, error) {
		if id != expectedID {
			t.Errorf("Expected ID %d, got %d", expectedID, id)
		}
		return domain.Actividad{
			ID:     expectedID,
			Titulo: expectedTitle,
		}, nil
	}

	result, err := service.GetByID(context.Background(), expectedID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != expectedID {
		t.Errorf("Expected ID %d, got %d", expectedID, result.ID)
	}
	if result.Titulo != expectedTitle {
		t.Errorf("Expected Titulo %s, got %s", expectedTitle, result.Titulo)
	}
}

func TestGetActividad_NotFound(t *testing.T) {
	mockRepo := &MockActividadesRepository{}
	service := NewActividadesService(mockRepo, &MockEventPublisher{})

	mockRepo.GetByIDFunc = func(ctx context.Context, id uint) (domain.Actividad, error) {
		return domain.Actividad{}, errors.New("actividad not found")
	}

	_, err := service.GetByID(context.Background(), 999)

	if err == nil {
		t.Error("Expected error for not found, got nil")
	}
}
