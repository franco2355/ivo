package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/payments-api/internal/domain/dtos"
	"github.com/yourusername/payments-api/internal/domain/entities"
	"github.com/yourusername/payments-api/internal/gateways"
	"github.com/yourusername/payments-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockPaymentRepository implementa PaymentRepository para tests
type MockPaymentRepository struct {
	CreateFunc       func(ctx context.Context, payment *entities.Payment) error
	FindByIDFunc     func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error)
	FindAllFunc      func(ctx context.Context) ([]*entities.Payment, error)
	FindByUserFunc   func(ctx context.Context, userID string) ([]*entities.Payment, error)
	FindByEntityFunc func(ctx context.Context, entityType, entityID string) ([]*entities.Payment, error)
	FindByStatusFunc func(ctx context.Context, status string) ([]*entities.Payment, error)
	UpdateStatusFunc func(ctx context.Context, id primitive.ObjectID, status, transactionID string) error
	CountFunc        func(ctx context.Context, filters map[string]interface{}) (int64, error)
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *entities.Payment) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, payment)
	}
	return nil
}

func (m *MockPaymentRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPaymentRepository) FindAll(ctx context.Context) ([]*entities.Payment, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx)
	}
	return []*entities.Payment{}, nil
}

func (m *MockPaymentRepository) FindByUser(ctx context.Context, userID string) ([]*entities.Payment, error) {
	if m.FindByUserFunc != nil {
		return m.FindByUserFunc(ctx, userID)
	}
	return []*entities.Payment{}, nil
}

func (m *MockPaymentRepository) FindByEntity(ctx context.Context, entityType, entityID string) ([]*entities.Payment, error) {
	if m.FindByEntityFunc != nil {
		return m.FindByEntityFunc(ctx, entityType, entityID)
	}
	return []*entities.Payment{}, nil
}

func (m *MockPaymentRepository) FindByStatus(ctx context.Context, status string) ([]*entities.Payment, error) {
	if m.FindByStatusFunc != nil {
		return m.FindByStatusFunc(ctx, status)
	}
	return []*entities.Payment{}, nil
}

func (m *MockPaymentRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status, transactionID string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status, transactionID)
	}
	return nil
}

func (m *MockPaymentRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, filters)
	}
	return 0, nil
}

// Verificar que MockPaymentRepository implementa la interfaz
var _ repository.PaymentRepository = (*MockPaymentRepository)(nil)

// MockEventPublisher implementa EventPublisher para tests
type MockEventPublisher struct {
	PublishPaymentCreatedFunc   func(ctx context.Context, payment *entities.Payment) error
	PublishPaymentCompletedFunc func(ctx context.Context, payment *entities.Payment) error
	PublishPaymentFailedFunc    func(ctx context.Context, payment *entities.Payment) error
	PublishPaymentRefundedFunc  func(ctx context.Context, payment *entities.Payment, refundAmount float64) error
}

func (m *MockEventPublisher) PublishPaymentCreated(ctx context.Context, payment *entities.Payment) error {
	if m.PublishPaymentCreatedFunc != nil {
		return m.PublishPaymentCreatedFunc(ctx, payment)
	}
	return nil
}

func (m *MockEventPublisher) PublishPaymentCompleted(ctx context.Context, payment *entities.Payment) error {
	if m.PublishPaymentCompletedFunc != nil {
		return m.PublishPaymentCompletedFunc(ctx, payment)
	}
	return nil
}

func (m *MockEventPublisher) PublishPaymentFailed(ctx context.Context, payment *entities.Payment) error {
	if m.PublishPaymentFailedFunc != nil {
		return m.PublishPaymentFailedFunc(ctx, payment)
	}
	return nil
}

func (m *MockEventPublisher) PublishPaymentRefunded(ctx context.Context, payment *entities.Payment, refundAmount float64) error {
	if m.PublishPaymentRefundedFunc != nil {
		return m.PublishPaymentRefundedFunc(ctx, payment, refundAmount)
	}
	return nil
}

// Verificar que MockEventPublisher implementa la interfaz
var _ EventPublisher = (*MockEventPublisher)(nil)

// TestCreatePayment_Success prueba la creación exitosa de un pago simple
func TestCreatePayment_Success(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		CreateFunc: func(ctx context.Context, payment *entities.Payment) error {
			return nil
		},
	}

	mockPublisher := &MockEventPublisher{
		PublishPaymentCreatedFunc: func(ctx context.Context, payment *entities.Payment) error {
			return nil
		},
	}

	service := NewPaymentService(mockRepo, nil, mockPublisher)

	req := dtos.CreatePaymentRequest{
		EntityType:     "subscription",
		EntityID:       "sub123",
		UserID:         "user456",
		Amount:         1000.0,
		Currency:       "ARS",
		PaymentMethod:  "cash",
		PaymentGateway: "manual",
	}

	resp, err := service.CreatePayment(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.EntityType != "subscription" {
		t.Errorf("Expected entity_type 'subscription', got: %s", resp.EntityType)
	}

	if resp.Amount != 1000.0 {
		t.Errorf("Expected amount 1000.0, got: %f", resp.Amount)
	}

	if resp.Status != "pending" {
		t.Errorf("Expected status 'pending', got: %s", resp.Status)
	}
}

// TestCreatePayment_RepositoryError prueba el error del repositorio
func TestCreatePayment_RepositoryError(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		CreateFunc: func(ctx context.Context, payment *entities.Payment) error {
			return errors.New("database error")
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	req := dtos.CreatePaymentRequest{
		EntityType:     "subscription",
		EntityID:       "sub123",
		UserID:         "user456",
		Amount:         1000.0,
		Currency:       "ARS",
		PaymentMethod:  "cash",
		PaymentGateway: "manual",
	}

	_, err := service.CreatePayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "error al crear pago") {
		t.Errorf("Expected error to contain 'error al crear pago', got: %s", err.Error())
	}
}

// TestGetPaymentByID_Success prueba obtener pago por ID
func TestGetPaymentByID_Success(t *testing.T) {
	paymentID := primitive.NewObjectID()

	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			if id == paymentID {
				return &entities.Payment{
					ID:             paymentID,
					EntityType:     "subscription",
					EntityID:       "sub123",
					UserID:         "user456",
					Amount:         1000.0,
					Currency:       "ARS",
					Status:         "completed",
					PaymentMethod:  "cash",
					PaymentGateway: "manual",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}, nil
			}
			return nil, errors.New("payment not found")
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	resp, err := service.GetPaymentByID(context.Background(), paymentID.Hex())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.ID != paymentID.Hex() {
		t.Errorf("Expected ID %s, got: %s", paymentID.Hex(), resp.ID)
	}

	if resp.Status != "completed" {
		t.Errorf("Expected status 'completed', got: %s", resp.Status)
	}
}

// TestGetPaymentByID_NotFound prueba pago no encontrado
func TestGetPaymentByID_NotFound(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			return nil, errors.New("payment not found")
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	paymentID := primitive.NewObjectID()
	_, err := service.GetPaymentByID(context.Background(), paymentID.Hex())

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "payment not found" {
		t.Errorf("Expected error 'payment not found', got: %s", err.Error())
	}
}

// TestGetPaymentByID_InvalidID prueba ID inválido
func TestGetPaymentByID_InvalidID(t *testing.T) {
	mockRepo := &MockPaymentRepository{}
	service := NewPaymentService(mockRepo, nil, nil)

	_, err := service.GetPaymentByID(context.Background(), "invalid-id")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "ID de pago inválido") {
		t.Errorf("Expected error to contain 'ID de pago inválido', got: %s", err.Error())
	}
}

// TestGetAllPayments_Success prueba obtener todos los pagos
func TestGetAllPayments_Success(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		FindAllFunc: func(ctx context.Context) ([]*entities.Payment, error) {
			return []*entities.Payment{
				{
					ID:         primitive.NewObjectID(),
					EntityType: "subscription",
					Status:     "completed",
					Amount:     1000.0,
				},
				{
					ID:         primitive.NewObjectID(),
					EntityType: "inscription",
					Status:     "pending",
					Amount:     500.0,
				},
			}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	payments, err := service.GetAllPayments(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(payments) != 2 {
		t.Errorf("Expected 2 payments, got: %d", len(payments))
	}

	if payments[0].EntityType != "subscription" {
		t.Errorf("Expected first payment entity_type 'subscription', got: %s", payments[0].EntityType)
	}
}

// TestGetAllPayments_Empty prueba cuando no hay pagos
func TestGetAllPayments_Empty(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		FindAllFunc: func(ctx context.Context) ([]*entities.Payment, error) {
			return []*entities.Payment{}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	payments, err := service.GetAllPayments(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(payments) != 0 {
		t.Errorf("Expected 0 payments, got: %d", len(payments))
	}
}

// TestGetPaymentsByUser_Success prueba obtener pagos por usuario
func TestGetPaymentsByUser_Success(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		FindByUserFunc: func(ctx context.Context, userID string) ([]*entities.Payment, error) {
			if userID == "user123" {
				return []*entities.Payment{
					{
						ID:         primitive.NewObjectID(),
						UserID:     "user123",
						EntityType: "subscription",
						Status:     "completed",
						Amount:     1000.0,
					},
				}, nil
			}
			return []*entities.Payment{}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	payments, err := service.GetPaymentsByUser(context.Background(), "user123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(payments) != 1 {
		t.Errorf("Expected 1 payment, got: %d", len(payments))
	}

	if payments[0].UserID != "user123" {
		t.Errorf("Expected user_id 'user123', got: %s", payments[0].UserID)
	}
}

// TestGetPaymentsByEntity_Success prueba obtener pagos por entidad
func TestGetPaymentsByEntity_Success(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		FindByEntityFunc: func(ctx context.Context, entityType, entityID string) ([]*entities.Payment, error) {
			if entityType == "subscription" && entityID == "sub123" {
				return []*entities.Payment{
					{
						ID:         primitive.NewObjectID(),
						EntityType: "subscription",
						EntityID:   "sub123",
						Status:     "completed",
						Amount:     1000.0,
					},
				}, nil
			}
			return []*entities.Payment{}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	payments, err := service.GetPaymentsByEntity(context.Background(), "subscription", "sub123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(payments) != 1 {
		t.Errorf("Expected 1 payment, got: %d", len(payments))
	}

	if payments[0].EntityID != "sub123" {
		t.Errorf("Expected entity_id 'sub123', got: %s", payments[0].EntityID)
	}
}

// TestGetPaymentsByStatus_Success prueba obtener pagos por estado
func TestGetPaymentsByStatus_Success(t *testing.T) {
	mockRepo := &MockPaymentRepository{
		FindByStatusFunc: func(ctx context.Context, status string) ([]*entities.Payment, error) {
			if status == "completed" {
				return []*entities.Payment{
					{
						ID:     primitive.NewObjectID(),
						Status: "completed",
						Amount: 1000.0,
					},
					{
						ID:     primitive.NewObjectID(),
						Status: "completed",
						Amount: 2000.0,
					},
				}, nil
			}
			return []*entities.Payment{}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	payments, err := service.GetPaymentsByStatus(context.Background(), "completed")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(payments) != 2 {
		t.Errorf("Expected 2 payments, got: %d", len(payments))
	}

	for _, p := range payments {
		if p.Status != "completed" {
			t.Errorf("Expected status 'completed', got: %s", p.Status)
		}
	}
}

// TestUpdatePaymentStatus_Success prueba actualizar estado exitosamente
func TestUpdatePaymentStatus_Success(t *testing.T) {
	paymentID := primitive.NewObjectID()

	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			return &entities.Payment{
				ID:             paymentID,
				EntityType:     "subscription",
				EntityID:       "sub123",
				UserID:         "user456",
				Amount:         1000.0,
				Status:         "pending",
				PaymentMethod:  "cash",
				PaymentGateway: "manual",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id primitive.ObjectID, status, transactionID string) error {
			if id == paymentID && status == "completed" {
				return nil
			}
			return errors.New("invalid update")
		},
	}

	eventPublished := false
	mockPublisher := &MockEventPublisher{
		PublishPaymentCompletedFunc: func(ctx context.Context, payment *entities.Payment) error {
			eventPublished = true
			return nil
		},
	}

	service := NewPaymentService(mockRepo, nil, mockPublisher)

	req := dtos.UpdatePaymentStatusRequest{
		Status:        "completed",
		TransactionID: "TXN123",
	}

	err := service.UpdatePaymentStatus(context.Background(), paymentID.Hex(), req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !eventPublished {
		t.Error("Expected payment.completed event to be published")
	}
}

// TestUpdatePaymentStatus_InvalidID prueba actualizar con ID inválido
func TestUpdatePaymentStatus_InvalidID(t *testing.T) {
	mockRepo := &MockPaymentRepository{}
	service := NewPaymentService(mockRepo, nil, nil)

	req := dtos.UpdatePaymentStatusRequest{
		Status:        "completed",
		TransactionID: "TXN123",
	}

	err := service.UpdatePaymentStatus(context.Background(), "invalid-id", req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "ID de pago inválido") {
		t.Errorf("Expected error to contain 'ID de pago inválido', got: %s", err.Error())
	}
}

// TestUpdatePaymentStatus_PaymentNotFound prueba actualizar pago inexistente
func TestUpdatePaymentStatus_PaymentNotFound(t *testing.T) {
	paymentID := primitive.NewObjectID()

	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			return nil, errors.New("payment not found")
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	req := dtos.UpdatePaymentStatusRequest{
		Status:        "completed",
		TransactionID: "TXN123",
	}

	err := service.UpdatePaymentStatus(context.Background(), paymentID.Hex(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "pago no encontrado") {
		t.Errorf("Expected error to contain 'pago no encontrado', got: %s", err.Error())
	}
}

// TestUpdatePaymentStatus_EventPublishError prueba que el error de evento no falla la actualización
func TestUpdatePaymentStatus_EventPublishError(t *testing.T) {
	paymentID := primitive.NewObjectID()

	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			return &entities.Payment{
				ID:             paymentID,
				EntityType:     "subscription",
				Status:         "pending",
				PaymentMethod:  "cash",
				PaymentGateway: "manual",
				Amount:         1000.0,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id primitive.ObjectID, status, transactionID string) error {
			return nil
		},
	}

	mockPublisher := &MockEventPublisher{
		PublishPaymentCompletedFunc: func(ctx context.Context, payment *entities.Payment) error {
			return errors.New("rabbitmq connection error")
		},
	}

	service := NewPaymentService(mockRepo, nil, mockPublisher)

	req := dtos.UpdatePaymentStatusRequest{
		Status:        "completed",
		TransactionID: "TXN123",
	}

	// El error de publicación NO debe fallar la actualización
	err := service.UpdatePaymentStatus(context.Background(), paymentID.Hex(), req)

	if err != nil {
		t.Fatalf("Expected no error (event publishing error should be logged, not returned), got: %v", err)
	}
}

// TestRefundPayment_Success prueba validaciones de reembolso (sin gateway real)
// Nota: El test completo con gateway requeriría inyección de dependencias más flexible
func TestRefundPayment_Validations(t *testing.T) {
	// Este test valida las validaciones antes de llamar al gateway
	// Los tests completos con gateway mock requerirían refactorizar el servicio para
	// permitir inyección de gateway directamente
	t.Skip("Refund test requires gateway factory injection refactoring")
}

// TestRefundPayment_InvalidStatus prueba reembolso de pago no completado
func TestRefundPayment_InvalidStatus(t *testing.T) {
	paymentID := primitive.NewObjectID()

	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			return &entities.Payment{
				ID:             paymentID,
				Status:         "pending",
				TransactionID:  "TXN123",
				Amount:         1000.0,
				PaymentGateway: "mercadopago",
			}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	err := service.RefundPayment(context.Background(), paymentID.Hex(), 500.0)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "solo se pueden reembolsar pagos completados") {
		t.Errorf("Expected error about status, got: %s", err.Error())
	}
}

// TestRefundPayment_ExceedsAmount prueba reembolso mayor al monto
func TestRefundPayment_ExceedsAmount(t *testing.T) {
	paymentID := primitive.NewObjectID()

	mockRepo := &MockPaymentRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*entities.Payment, error) {
			return &entities.Payment{
				ID:             paymentID,
				Status:         "completed",
				TransactionID:  "TXN123",
				Amount:         1000.0,
				PaymentGateway: "mercadopago",
			}, nil
		},
	}

	service := NewPaymentService(mockRepo, nil, nil)

	err := service.RefundPayment(context.Background(), paymentID.Hex(), 1500.0)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "monto de reembolso mayor al monto del pago") {
		t.Errorf("Expected error about amount, got: %s", err.Error())
	}
}

// Mocks adicionales para gateway

type MockPaymentGateway struct {
	GetNameFunc              func() string
	CreatePaymentFunc        func(ctx context.Context, req gateways.PaymentRequest) (*gateways.PaymentResult, error)
	GetPaymentStatusFunc     func(ctx context.Context, transactionID string) (*gateways.PaymentStatus, error)
	RefundPaymentFunc        func(ctx context.Context, transactionID string, amount float64) (*gateways.RefundResult, error)
	CancelPaymentFunc        func(ctx context.Context, transactionID string) error
	ProcessWebhookFunc       func(ctx context.Context, payload []byte, headers map[string]string) (*gateways.WebhookEvent, error)
	ValidateCredentialsFunc  func(ctx context.Context) error
}

func (m *MockPaymentGateway) GetName() string {
	if m.GetNameFunc != nil {
		return m.GetNameFunc()
	}
	return "mock"
}

func (m *MockPaymentGateway) CreatePayment(ctx context.Context, req gateways.PaymentRequest) (*gateways.PaymentResult, error) {
	if m.CreatePaymentFunc != nil {
		return m.CreatePaymentFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockPaymentGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*gateways.PaymentStatus, error) {
	if m.GetPaymentStatusFunc != nil {
		return m.GetPaymentStatusFunc(ctx, transactionID)
	}
	return nil, nil
}

func (m *MockPaymentGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*gateways.RefundResult, error) {
	if m.RefundPaymentFunc != nil {
		return m.RefundPaymentFunc(ctx, transactionID, amount)
	}
	return nil, nil
}

func (m *MockPaymentGateway) CancelPayment(ctx context.Context, transactionID string) error {
	if m.CancelPaymentFunc != nil {
		return m.CancelPaymentFunc(ctx, transactionID)
	}
	return nil
}

func (m *MockPaymentGateway) ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*gateways.WebhookEvent, error) {
	if m.ProcessWebhookFunc != nil {
		return m.ProcessWebhookFunc(ctx, payload, headers)
	}
	return nil, nil
}

func (m *MockPaymentGateway) ValidateCredentials(ctx context.Context) error {
	if m.ValidateCredentialsFunc != nil {
		return m.ValidateCredentialsFunc(ctx)
	}
	return nil
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
