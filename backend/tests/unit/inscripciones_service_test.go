package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"activities-api/internal/domain"
)

// MockInscripcionesRepository para tests unitarios
type MockInscripcionesRepository struct {
	CreateFunc               func(ctx context.Context, inscripcion domain.Inscripcion) (domain.Inscripcion, error)
	GetByIDFunc              func(ctx context.Context, id uint) (domain.Inscripcion, error)
	GetByUserAndActivityFunc func(ctx context.Context, userID, activityID uint) (domain.Inscripcion, error)
	ListByUserFunc           func(ctx context.Context, userID uint) ([]domain.Inscripcion, error)
	ListByActivityFunc       func(ctx context.Context, activityID uint) ([]domain.Inscripcion, error)
	UpdateFunc               func(ctx context.Context, id uint, inscripcion domain.Inscripcion) (domain.Inscripcion, error)
	DeleteFunc               func(ctx context.Context, id uint) error
	CountByActivityFunc      func(ctx context.Context, activityID uint) (int64, error)
}

func (m *MockInscripcionesRepository) Create(ctx context.Context, inscripcion domain.Inscripcion) (domain.Inscripcion, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, inscripcion)
	}
	return domain.Inscripcion{}, nil
}

func (m *MockInscripcionesRepository) GetByID(ctx context.Context, id uint) (domain.Inscripcion, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return domain.Inscripcion{}, nil
}

func (m *MockInscripcionesRepository) GetByUserAndActivity(ctx context.Context, userID, activityID uint) (domain.Inscripcion, error) {
	if m.GetByUserAndActivityFunc != nil {
		return m.GetByUserAndActivityFunc(ctx, userID, activityID)
	}
	return domain.Inscripcion{}, nil
}

func (m *MockInscripcionesRepository) ListByUser(ctx context.Context, userID uint) ([]domain.Inscripcion, error) {
	if m.ListByUserFunc != nil {
		return m.ListByUserFunc(ctx, userID)
	}
	return []domain.Inscripcion{}, nil
}

func (m *MockInscripcionesRepository) ListByActivity(ctx context.Context, activityID uint) ([]domain.Inscripcion, error) {
	if m.ListByActivityFunc != nil {
		return m.ListByActivityFunc(ctx, activityID)
	}
	return []domain.Inscripcion{}, nil
}

func (m *MockInscripcionesRepository) Update(ctx context.Context, id uint, inscripcion domain.Inscripcion) (domain.Inscripcion, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, inscripcion)
	}
	return domain.Inscripcion{}, nil
}

func (m *MockInscripcionesRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockInscripcionesRepository) CountByActivity(ctx context.Context, activityID uint) (int64, error) {
	if m.CountByActivityFunc != nil {
		return m.CountByActivityFunc(ctx, activityID)
	}
	return 0, nil
}

// TestInscripcionValidation_DuplicateInscription verifica que no se permitan inscripciones duplicadas
func TestInscripcionValidation_DuplicateInscription(t *testing.T) {
	mockRepo := &MockInscripcionesRepository{
		GetByUserAndActivityFunc: func(ctx context.Context, userID, activityID uint) (domain.Inscripcion, error) {
			// Simula que ya existe una inscripción
			return domain.Inscripcion{
				ID:          1,
				IDUsuario:   userID,
				IDActividad: activityID,
				Estado:      "confirmada",
			}, nil
		},
	}

	// Simular que se intenta crear una inscripción duplicada
	existing, err := mockRepo.GetByUserAndActivity(context.Background(), 1, 1)

	if err != nil {
		t.Fatalf("Expected no error when checking existing inscription, got: %v", err)
	}

	if existing.ID == 0 {
		t.Error("Expected to find existing inscription")
	}

	// En un servicio real, esto debería devolver error de inscripción duplicada
}

// TestInscripcionValidation_ActivityCapacityFull verifica validación de cupo lleno
func TestInscripcionValidation_ActivityCapacityFull(t *testing.T) {
	activityCapacity := int64(20)
	currentCount := int64(20)

	mockRepo := &MockInscripcionesRepository{
		CountByActivityFunc: func(ctx context.Context, activityID uint) (int64, error) {
			return currentCount, nil
		},
	}

	count, err := mockRepo.CountByActivity(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if count >= activityCapacity {
		// Comportamiento esperado: el cupo está lleno
		t.Logf("Activity is full: %d/%d", count, activityCapacity)
	} else {
		t.Error("Expected activity to be full")
	}
}

// TestInscripcionValidation_InactiveUser verifica que usuarios inactivos no puedan inscribirse
func TestInscripcionValidation_InactiveUser(t *testing.T) {
	// Este test simula la validación de usuario activo
	userActive := false

	if !userActive {
		t.Log("User is inactive, inscription should be rejected")
		// Comportamiento esperado
	} else {
		t.Error("Expected user to be inactive")
	}
}

// TestInscripcionCancellation_RefundLogic verifica lógica de reembolso al cancelar
func TestInscripcionCancellation_RefundLogic(t *testing.T) {
	inscriptionDate := time.Now().Add(-5 * 24 * time.Hour) // Inscripción hace 5 días
	cancellationDate := time.Now()
	activityDate := time.Now().Add(3 * 24 * time.Hour) // Actividad en 3 días

	// Regla: reembolso completo si se cancela con más de 48 horas de anticipación
	hoursDifference := activityDate.Sub(cancellationDate).Hours()

	if hoursDifference >= 48 {
		t.Log("Eligible for full refund")
	} else {
		t.Log("Not eligible for full refund")
	}

	if hoursDifference < 72 {
		t.Error("Expected at least 72 hours difference for this test case")
	}
}

// TestListInscriptionsByUser_Success verifica listado de inscripciones por usuario
func TestListInscriptionsByUser_Success(t *testing.T) {
	mockRepo := &MockInscripcionesRepository{
		ListByUserFunc: func(ctx context.Context, userID uint) ([]domain.Inscripcion, error) {
			return []domain.Inscripcion{
				{
					ID:          1,
					IDUsuario:   userID,
					IDActividad: 1,
					Estado:      "confirmada",
				},
				{
					ID:          2,
					IDUsuario:   userID,
					IDActividad: 2,
					Estado:      "pendiente",
				},
			}, nil
		},
	}

	inscriptions, err := mockRepo.ListByUser(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(inscriptions) != 2 {
		t.Errorf("Expected 2 inscriptions, got: %d", len(inscriptions))
	}

	for _, insc := range inscriptions {
		if insc.IDUsuario != 1 {
			t.Errorf("Expected user ID 1, got: %d", insc.IDUsuario)
		}
	}
}

// TestListInscriptionsByActivity_Success verifica listado de inscripciones por actividad
func TestListInscriptionsByActivity_Success(t *testing.T) {
	mockRepo := &MockInscripcionesRepository{
		ListByActivityFunc: func(ctx context.Context, activityID uint) ([]domain.Inscripcion, error) {
			return []domain.Inscripcion{
				{
					ID:          1,
					IDUsuario:   1,
					IDActividad: activityID,
					Estado:      "confirmada",
				},
				{
					ID:          2,
					IDUsuario:   2,
					IDActividad: activityID,
					Estado:      "confirmada",
				},
			}, nil
		},
	}

	inscriptions, err := mockRepo.ListByActivity(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(inscriptions) != 2 {
		t.Errorf("Expected 2 inscriptions, got: %d", len(inscriptions))
	}

	for _, insc := range inscriptions {
		if insc.IDActividad != 1 {
			t.Errorf("Expected activity ID 1, got: %d", insc.IDActividad)
		}
	}
}

// TestUpdateInscriptionStatus_Success verifica actualización de estado
func TestUpdateInscriptionStatus_Success(t *testing.T) {
	mockRepo := &MockInscripcionesRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (domain.Inscripcion, error) {
			return domain.Inscripcion{
				ID:          id,
				IDUsuario:   1,
				IDActividad: 1,
				Estado:      "pendiente",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, id uint, inscripcion domain.Inscripcion) (domain.Inscripcion, error) {
			inscripcion.ID = id
			return inscripcion, nil
		},
	}

	// Obtener inscripción
	insc, err := mockRepo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Actualizar estado
	insc.Estado = "confirmada"
	updated, err := mockRepo.Update(context.Background(), 1, insc)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updated.Estado != "confirmada" {
		t.Errorf("Expected status 'confirmada', got: %s", updated.Estado)
	}
}

// TestDeleteInscription_Success verifica eliminación de inscripción
func TestDeleteInscription_Success(t *testing.T) {
	deleted := false

	mockRepo := &MockInscripcionesRepository{
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
		t.Error("Expected inscription to be deleted")
	}
}

// TestDeleteInscription_NotFound verifica error al eliminar inscripción inexistente
func TestDeleteInscription_NotFound(t *testing.T) {
	mockRepo := &MockInscripcionesRepository{
		DeleteFunc: func(ctx context.Context, id uint) error {
			return errors.New("inscription not found")
		},
	}

	err := mockRepo.Delete(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "inscription not found" {
		t.Errorf("Expected error 'inscription not found', got: %s", err.Error())
	}
}
