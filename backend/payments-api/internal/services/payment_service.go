package services

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/payments-api/internal/domain/dtos"
	"github.com/yourusername/payments-api/internal/domain/entities"
	"github.com/yourusername/payments-api/internal/gateways"
	"github.com/yourusername/payments-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentService - Servicio completo que integra gateways de pago
// Orquesta: Repository (persistencia) + GatewayFactory (gateways externos) + RabbitMQ (eventos)
// Patrón: Facade + Orchestration + Event-Driven
type PaymentService struct {
	paymentRepo    repository.PaymentRepository
	gatewayFactory *gateways.GatewayFactory
	eventPublisher EventPublisher // Interface para publicar eventos (RabbitMQ, Kafka, etc.)
}

// EventPublisher - Interface para publicar eventos de pagos (Strategy Pattern)
// Permite cambiar implementación (RabbitMQ, Kafka, SNS, etc.) sin modificar el servicio
type EventPublisher interface {
	PublishPaymentCreated(ctx context.Context, payment *entities.Payment) error
	PublishPaymentCompleted(ctx context.Context, payment *entities.Payment) error
	PublishPaymentFailed(ctx context.Context, payment *entities.Payment) error
	PublishPaymentRefunded(ctx context.Context, payment *entities.Payment, refundAmount float64) error
}

// NewPaymentService - Constructor con Dependency Injection
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	gatewayFactory *gateways.GatewayFactory,
	eventPublisher EventPublisher, // Opcional: puede ser nil si no querés eventos
) *PaymentService {
	return &PaymentService{
		paymentRepo:    paymentRepo,
		gatewayFactory: gatewayFactory,
		eventPublisher: eventPublisher,
	}
}

// ProcessRecurringPayment - Crea un pago recurrente (suscripción) en el gateway
// Usa SubscriptionGateway (Preapprovals de MP) pero guarda en tabla payments normal
func (s *PaymentService) ProcessRecurringPayment(
	ctx context.Context,
	req dtos.CreatePaymentRequest,
	frequency int,
	frequencyType string,
) (dtos.PaymentResponse, error) {
	// 0. VALIDACIÓN DE IDEMPOTENCIA ⭐
	// Si el cliente envió un idempotency_key, verificar si ya existe un pago con ese key
	if req.IdempotencyKey != "" {
		existing, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			// Ya existe un pago con este idempotency key, retornar el pago original
			fmt.Printf("⚠️ Pago duplicado detectado (idempotency_key=%s), retornando pago original ID=%s\n", req.IdempotencyKey, existing.ID.Hex())
			return dtos.ToPaymentResponse(
				existing.ID,
				existing.EntityType,
				existing.EntityID,
				existing.UserID,
				existing.Amount,
				existing.Currency,
				existing.Status,
				existing.PaymentMethod,
				existing.PaymentGateway,
				existing.TransactionID,
				existing.IdempotencyKey,
				existing.Metadata,
				existing.CreatedAt,
				existing.UpdatedAt,
				existing.ProcessedAt,
			), nil
		}
	}

	// 1. Crear registro en base de datos con estado "pending"
	payment := entities.Payment{
		ID:             primitive.NewObjectID(),
		EntityType:     req.EntityType,
		EntityID:       req.EntityID,
		UserID:         req.UserID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         gateways.SubscriptionStatusPending,
		PaymentMethod:  req.PaymentMethod,
		PaymentGateway: req.PaymentGateway,
		PaymentType:    "recurring",        // Marcar como recurrente
		IdempotencyKey: req.IdempotencyKey, // ⭐ Guardar idempotency key
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, &payment); err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error creando registro de pago: %w", err)
	}

	// 1.5. PUBLICAR EVENTO payment.created ⭐
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishPaymentCreated(ctx, &payment); err != nil {
			fmt.Printf("⚠️ Error publicando evento payment.created: %v\n", err)
		}
	}

	// 2. Obtener el subscription gateway
	subscriptionGateway, err := s.gatewayFactory.CreateSubscriptionGateway(req.PaymentGateway)
	if err != nil {
		s.paymentRepo.UpdateStatus(ctx, payment.ID, gateways.StatusFailed, "")
		return dtos.PaymentResponse{}, fmt.Errorf("error creando subscription gateway: %w", err)
	}

	// 3. Preparar request para el gateway
	gatewayRequest := gateways.SubscriptionRequest{
		Reason:        fmt.Sprintf("%s - %s", req.EntityType, req.EntityID),
		Amount:        req.Amount,
		Currency:      req.Currency,
		Frequency:     frequency,
		FrequencyType: frequencyType,
		CustomerEmail: extractCustomerEmail(req.Metadata),
		CustomerName:  extractCustomerName(req.Metadata),
		CustomerID:    req.UserID,
		ExternalID:    payment.ID.Hex(),
		CallbackURL:   "", // TODO: Configurar callback URL
		WebhookURL:    "", // TODO: Configurar webhook URL
		Metadata:      req.Metadata,
	}

	// 4. Crear suscripción en el gateway externo
	result, err := subscriptionGateway.CreateSubscription(ctx, gatewayRequest)
	if err != nil {
		s.paymentRepo.UpdateStatus(ctx, payment.ID, gateways.StatusFailed, "")
		return dtos.PaymentResponse{}, fmt.Errorf("error creando suscripción en gateway: %w", err)
	}

	// 5. Actualizar pago con resultado del gateway
	payment.Status = result.Status
	payment.TransactionID = result.SubscriptionID // Guardar subscription_id como transaction_id
	payment.UpdatedAt = time.Now()

	// Guardar subscription ID
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, result.Status, result.SubscriptionID); err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error actualizando estado del pago: %w", err)
	}

	// 6. Agregar init_point al metadata de respuesta
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["init_point"] = result.InitPoint
	req.Metadata["gateway_message"] = result.Message
	req.Metadata["subscription_id"] = result.SubscriptionID
	req.Metadata["frequency"] = frequency
	req.Metadata["frequency_type"] = frequencyType

	// 7. Retornar respuesta
	return dtos.ToPaymentResponse(
		payment.ID,
		payment.EntityType,
		payment.EntityID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.TransactionID, payment.IdempotencyKey, req.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.ProcessedAt,
	), nil
}

// ProcessPaymentWithGateway - Crea un pago Y lo procesa en el gateway externo (Checkout Pro)
// Este es el método principal que debes usar para pagos únicos
func (s *PaymentService) ProcessPaymentWithGateway(
	ctx context.Context,
	req dtos.CreatePaymentRequest,
) (dtos.PaymentResponse, error) {
	// 0. VALIDACIÓN DE IDEMPOTENCIA ⭐
	// Si el cliente envió un idempotency_key, verificar si ya existe un pago con ese key
	if req.IdempotencyKey != "" {
		existing, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			// Ya existe un pago con este idempotency key, retornar el pago original
			fmt.Printf("⚠️ Pago duplicado detectado (idempotency_key=%s), retornando pago original ID=%s\n", req.IdempotencyKey, existing.ID.Hex())
			return dtos.ToPaymentResponse(
				existing.ID,
				existing.EntityType,
				existing.EntityID,
				existing.UserID,
				existing.Amount,
				existing.Currency,
				existing.Status,
				existing.PaymentMethod,
				existing.PaymentGateway,
				existing.TransactionID,
				existing.IdempotencyKey,
				existing.Metadata,
				existing.CreatedAt,
				existing.UpdatedAt,
				existing.ProcessedAt,
			), nil
		}
	}

	// 1. Crear registro en base de datos con estado "pending"
	payment := entities.Payment{
		ID:             primitive.NewObjectID(),
		EntityType:     req.EntityType,
		EntityID:       req.EntityID,
		UserID:         req.UserID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         gateways.StatusPending,
		PaymentMethod:  req.PaymentMethod,
		PaymentGateway: req.PaymentGateway,
		PaymentType:    "one_time",         // Marcar como pago único
		IdempotencyKey: req.IdempotencyKey, // ⭐ Guardar idempotency key
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, &payment); err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error creando registro de pago: %w", err)
	}

	// 1.5. PUBLICAR EVENTO payment.created ⭐
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishPaymentCreated(ctx, &payment); err != nil {
			// Log el error pero NO fallamos el request (el pago ya se creó)
			fmt.Printf("⚠️ Error publicando evento payment.created: %v\n", err)
		}
	}

	// 2. Obtener el gateway apropiado desde el factory
	gateway, err := s.gatewayFactory.CreateGateway(req.PaymentGateway)
	if err != nil {
		// Actualizar pago como fallido
		s.paymentRepo.UpdateStatus(ctx, payment.ID, gateways.StatusFailed, "")
		return dtos.PaymentResponse{}, fmt.Errorf("error creando gateway: %w", err)
	}

	// 3. Preparar request para el gateway
	gatewayRequest := gateways.PaymentRequest{
		Amount:        req.Amount,
		Currency:      req.Currency,
		Description:   fmt.Sprintf("Pago %s #%s", req.EntityType, req.EntityID),
		CustomerEmail: extractCustomerEmail(req.Metadata),
		CustomerName:  extractCustomerName(req.Metadata),
		PaymentMethod: req.PaymentMethod,
		Metadata:      req.Metadata,
		CallbackURL:   req.CallbackURL,
		WebhookURL:    req.WebhookURL,
		ExternalID:    payment.ID.Hex(),
		CustomerID:    req.UserID,
	}

	// 4. Procesar pago en el gateway externo
	result, err := gateway.CreatePayment(ctx, gatewayRequest)
	if err != nil {
		// Actualizar pago como fallido
		s.paymentRepo.UpdateStatus(ctx, payment.ID, gateways.StatusFailed, "")
		return dtos.PaymentResponse{}, fmt.Errorf("error procesando pago en gateway: %w", err)
	}

	// 5. Actualizar pago con resultado del gateway
	payment.Status = result.Status
	payment.TransactionID = result.TransactionID
	payment.UpdatedAt = time.Now()

	if result.Status == gateways.StatusCompleted {
		now := time.Now()
		payment.ProcessedAt = &now
	}

	// Guardar transaction ID y estado
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, result.Status, result.TransactionID); err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error actualizando estado del pago: %w", err)
	}

	// 5.5. PUBLICAR EVENTO EN RABBITMQ ⭐
	if s.eventPublisher != nil {
		if result.Status == gateways.StatusCompleted {
			// Evento: Pago completado exitosamente
			if err := s.eventPublisher.PublishPaymentCompleted(ctx, &payment); err != nil {
				// Log el error pero NO fallamos el request (el pago ya se procesó)
				fmt.Printf("⚠️ Error publicando evento payment.completed: %v\n", err)
			}
		} else if result.Status == gateways.StatusFailed {
			// Evento: Pago fallido
			if err := s.eventPublisher.PublishPaymentFailed(ctx, &payment); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.failed: %v\n", err)
			}
		}
	}

	// 6. Agregar metadata del gateway a la respuesta
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["payment_url"] = result.PaymentURL
	req.Metadata["gateway_message"] = result.Message

	// 7. Retornar respuesta
	return dtos.ToPaymentResponse(
		payment.ID,
		payment.EntityType,
		payment.EntityID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.TransactionID, payment.IdempotencyKey, req.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.ProcessedAt,
	), nil
}

// SyncPaymentStatus - Sincroniza el estado de un pago con el gateway
// Útil para actualizar pagos pendientes
func (s *PaymentService) SyncPaymentStatus(ctx context.Context, paymentID string) (dtos.PaymentResponse, error) {
	// 1. Obtener pago de la BD
	objID, err := primitive.ObjectIDFromHex(paymentID)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("ID de pago inválido")
	}

	payment, err := s.paymentRepo.FindByID(ctx, objID)
	if err != nil {
		return dtos.PaymentResponse{}, err
	}

	// 2. Validar que tenga transaction ID
	if payment.TransactionID == "" {
		return dtos.PaymentResponse{}, fmt.Errorf("pago sin transaction ID, no se puede sincronizar")
	}

	// 3. Obtener gateway
	gateway, err := s.gatewayFactory.CreateGateway(payment.PaymentGateway)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error creando gateway: %w", err)
	}

	// 4. Consultar estado en el gateway
	status, err := gateway.GetPaymentStatus(ctx, payment.TransactionID)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error consultando estado en gateway: %w", err)
	}

	// 5. Actualizar en BD si cambió el estado
	if status.Status != payment.Status {
		if err := s.paymentRepo.UpdateStatus(ctx, objID, status.Status, payment.TransactionID); err != nil {
			return dtos.PaymentResponse{}, fmt.Errorf("error actualizando estado: %w", err)
		}
		payment.Status = status.Status
		payment.UpdatedAt = time.Now()

		if status.Status == gateways.StatusCompleted && payment.ProcessedAt == nil {
			now := time.Now()
			payment.ProcessedAt = &now
		}
	}

	// 6. Retornar respuesta actualizada
	return dtos.ToPaymentResponse(
		payment.ID,
		payment.EntityType,
		payment.EntityID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.ProcessedAt,
	), nil
}

// RefundPayment - Procesa un reembolso a través del gateway
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount float64) error {
	// 1. Obtener pago
	objID, err := primitive.ObjectIDFromHex(paymentID)
	if err != nil {
		return fmt.Errorf("ID de pago inválido")
	}

	payment, err := s.paymentRepo.FindByID(ctx, objID)
	if err != nil {
		return err
	}

	// 2. Validaciones
	if payment.Status != gateways.StatusCompleted {
		return fmt.Errorf("solo se pueden reembolsar pagos completados")
	}

	if payment.TransactionID == "" {
		return fmt.Errorf("pago sin transaction ID")
	}

	if amount > payment.Amount {
		return fmt.Errorf("monto de reembolso mayor al monto del pago")
	}

	// 3. Obtener gateway
	gateway, err := s.gatewayFactory.CreateGateway(payment.PaymentGateway)
	if err != nil {
		return fmt.Errorf("error creando gateway: %w", err)
	}

	// 4. Procesar reembolso en el gateway
	_, err = gateway.RefundPayment(ctx, payment.TransactionID, amount)
	if err != nil {
		return fmt.Errorf("error procesando reembolso en gateway: %w", err)
	}

	// 5. Actualizar estado en BD
	if err := s.paymentRepo.UpdateStatus(ctx, objID, gateways.StatusRefunded, payment.TransactionID); err != nil {
		return fmt.Errorf("error actualizando estado a refunded: %w", err)
	}

	// 6. Actualizar payment en memoria para el evento
	payment.Status = gateways.StatusRefunded
	payment.UpdatedAt = time.Now()

	// 7. PUBLICAR EVENTO EN RABBITMQ ⭐
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishPaymentRefunded(ctx, payment, amount); err != nil {
			// Log el error pero NO fallamos el request (el reembolso ya se procesó)
			fmt.Printf("⚠️ Error publicando evento payment.refunded: %v\n", err)
		}
	}

	return nil
}

// GetPaymentByID - Obtiene un pago por ID (mismo que servicio básico)
func (s *PaymentService) GetPaymentByID(ctx context.Context, paymentID string) (dtos.PaymentResponse, error) {
	objID, err := primitive.ObjectIDFromHex(paymentID)
	if err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("ID de pago inválido")
	}

	payment, err := s.paymentRepo.FindByID(ctx, objID)
	if err != nil {
		return dtos.PaymentResponse{}, err
	}

	return dtos.ToPaymentResponse(
		payment.ID,
		payment.EntityType,
		payment.EntityID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.ProcessedAt,
	), nil
}

// GetAllPayments - Obtiene todos los pagos
func (s *PaymentService) GetAllPayments(ctx context.Context) ([]dtos.PaymentResponse, error) {
	payments, err := s.paymentRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = dtos.ToPaymentResponse(
			payment.ID,
			payment.EntityType,
			payment.EntityID,
			payment.UserID,
			payment.Amount,
			payment.Currency,
			payment.Status,
			payment.PaymentMethod,
			payment.PaymentGateway,
			payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
			payment.CreatedAt,
			payment.UpdatedAt,
			payment.ProcessedAt,
		)
	}

	return responses, nil
}

// GetPaymentsByUser - Obtiene todos los pagos de un usuario
func (s *PaymentService) GetPaymentsByUser(ctx context.Context, userID string) ([]dtos.PaymentResponse, error) {
	payments, err := s.paymentRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = dtos.ToPaymentResponse(
			payment.ID,
			payment.EntityType,
			payment.EntityID,
			payment.UserID,
			payment.Amount,
			payment.Currency,
			payment.Status,
			payment.PaymentMethod,
			payment.PaymentGateway,
			payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
			payment.CreatedAt,
			payment.UpdatedAt,
			payment.ProcessedAt,
		)
	}

	return responses, nil
}

// GetPaymentsByEntity - Obtiene pagos asociados a una entidad
func (s *PaymentService) GetPaymentsByEntity(ctx context.Context, entityType, entityID string) ([]dtos.PaymentResponse, error) {
	payments, err := s.paymentRepo.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = dtos.ToPaymentResponse(
			payment.ID,
			payment.EntityType,
			payment.EntityID,
			payment.UserID,
			payment.Amount,
			payment.Currency,
			payment.Status,
			payment.PaymentMethod,
			payment.PaymentGateway,
			payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
			payment.CreatedAt,
			payment.UpdatedAt,
			payment.ProcessedAt,
		)
	}

	return responses, nil
}

// GetPaymentsByStatus - Obtiene pagos por estado
func (s *PaymentService) GetPaymentsByStatus(ctx context.Context, status string) ([]dtos.PaymentResponse, error) {
	payments, err := s.paymentRepo.FindByStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = dtos.ToPaymentResponse(
			payment.ID,
			payment.EntityType,
			payment.EntityID,
			payment.UserID,
			payment.Amount,
			payment.Currency,
			payment.Status,
			payment.PaymentMethod,
			payment.PaymentGateway,
			payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
			payment.CreatedAt,
			payment.UpdatedAt,
			payment.ProcessedAt,
		)
	}

	return responses, nil
}

// CreatePayment - Crea un pago simple sin gateway (para compatibilidad)
func (s *PaymentService) CreatePayment(ctx context.Context, req dtos.CreatePaymentRequest) (dtos.PaymentResponse, error) {
	// VALIDACIÓN DE IDEMPOTENCIA ⭐
	if req.IdempotencyKey != "" {
		existing, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			fmt.Printf("⚠️ Pago duplicado detectado (idempotency_key=%s), retornando pago original ID=%s\n", req.IdempotencyKey, existing.ID.Hex())
			return dtos.ToPaymentResponse(
				existing.ID,
				existing.EntityType,
				existing.EntityID,
				existing.UserID,
				existing.Amount,
				existing.Currency,
				existing.Status,
				existing.PaymentMethod,
				existing.PaymentGateway,
				existing.TransactionID,
				existing.IdempotencyKey,
				existing.Metadata,
				existing.CreatedAt,
				existing.UpdatedAt,
				existing.ProcessedAt,
			), nil
		}
	}

	payment := entities.Payment{
		ID:             primitive.NewObjectID(),
		EntityType:     req.EntityType,
		EntityID:       req.EntityID,
		UserID:         req.UserID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         "pending",
		PaymentMethod:  req.PaymentMethod,
		PaymentGateway: req.PaymentGateway,
		IdempotencyKey: req.IdempotencyKey, // ⭐ Guardar idempotency key
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, &payment); err != nil {
		return dtos.PaymentResponse{}, fmt.Errorf("error al crear pago: %w", err)
	}

	// PUBLICAR EVENTO payment.created ⭐
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishPaymentCreated(ctx, &payment); err != nil {
			fmt.Printf("⚠️ Error publicando evento payment.created: %v\n", err)
		}
	}

	return dtos.ToPaymentResponse(
		payment.ID,
		payment.EntityType,
		payment.EntityID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.TransactionID, payment.IdempotencyKey, payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.ProcessedAt,
	), nil
}

// UpdatePaymentStatus - Actualiza el estado de un pago manualmente Y PUBLICA EVENTOS
func (s *PaymentService) UpdatePaymentStatus(ctx context.Context, paymentID string, req dtos.UpdatePaymentStatusRequest) error {
	objID, err := primitive.ObjectIDFromHex(paymentID)
	if err != nil {
		return fmt.Errorf("ID de pago inválido")
	}

	// 1. Obtener pago actual para comparar estados y publicar eventos
	payment, err := s.paymentRepo.FindByID(ctx, objID)
	if err != nil {
		return fmt.Errorf("pago no encontrado: %w", err)
	}

	oldStatus := payment.Status

	// 2. Actualizar en BD
	if err := s.paymentRepo.UpdateStatus(ctx, objID, req.Status, req.TransactionID); err != nil {
		return fmt.Errorf("error actualizando estado: %w", err)
	}

	// 3. Actualizar payment en memoria
	payment.Status = req.Status
	payment.TransactionID = req.TransactionID
	payment.UpdatedAt = time.Now()

	if req.Status == gateways.StatusCompleted {
		now := time.Now()
		payment.ProcessedAt = &now
	}

	// 4. PUBLICAR EVENTO SOLO SI CAMBIÓ EL ESTADO
	if oldStatus != req.Status && s.eventPublisher != nil {
		switch req.Status {
		case gateways.StatusCompleted:
			if err := s.eventPublisher.PublishPaymentCompleted(ctx, payment); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.completed: %v\n", err)
			}
		case gateways.StatusFailed:
			if err := s.eventPublisher.PublishPaymentFailed(ctx, payment); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.failed: %v\n", err)
			}
		case gateways.StatusRefunded:
			if err := s.eventPublisher.PublishPaymentRefunded(ctx, payment, payment.Amount); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.refunded: %v\n", err)
			}
		}
	}

	return nil
}

// ProcessPayment - Simula el procesamiento de un pago (para testing)
func (s *PaymentService) ProcessPayment(ctx context.Context, paymentID string) error {
	// Simular procesamiento
	time.Sleep(100 * time.Millisecond)

	// Actualizar a completado
	req := dtos.UpdatePaymentStatusRequest{
		Status:        "completed",
		TransactionID: fmt.Sprintf("TXN-%s", paymentID),
	}

	return s.UpdatePaymentStatus(ctx, paymentID, req)
}

// UpdatePaymentStatusFromWebhook - Actualiza estado de pago desde webhook Y publica eventos
// Usado por WebhookController cuando recibe notificaciones de gateways
func (s *PaymentService) UpdatePaymentStatusFromWebhook(ctx context.Context, paymentID primitive.ObjectID, newStatus, transactionID string) error {
	// 1. Obtener pago actual
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("pago no encontrado: %w", err)
	}

	oldStatus := payment.Status

	// 2. Actualizar en BD
	if err := s.paymentRepo.UpdateStatus(ctx, paymentID, newStatus, transactionID); err != nil {
		return fmt.Errorf("error actualizando estado: %w", err)
	}

	// 3. Actualizar payment en memoria
	payment.Status = newStatus
	payment.TransactionID = transactionID
	payment.UpdatedAt = time.Now()

	if newStatus == gateways.StatusCompleted {
		now := time.Now()
		payment.ProcessedAt = &now
	}

	// 4. PUBLICAR EVENTO SOLO SI CAMBIÓ EL ESTADO ⭐
	if oldStatus != newStatus && s.eventPublisher != nil {
		switch newStatus {
		case gateways.StatusCompleted:
			if err := s.eventPublisher.PublishPaymentCompleted(ctx, payment); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.completed: %v\n", err)
			}
		case gateways.StatusFailed:
			if err := s.eventPublisher.PublishPaymentFailed(ctx, payment); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.failed: %v\n", err)
			}
		case gateways.StatusRefunded:
			if err := s.eventPublisher.PublishPaymentRefunded(ctx, payment, payment.Amount); err != nil {
				fmt.Printf("⚠️ Error publicando evento payment.refunded: %v\n", err)
			}
		}
	}

	return nil
}

// Helper functions

func extractCustomerEmail(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	if email, ok := metadata["customer_email"].(string); ok {
		return email
	}
	return ""
}

func extractCustomerName(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	if name, ok := metadata["customer_name"].(string); ok {
		return name
	}
	return ""
}
