package services

import (
	"context"
	"time"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/entities"
	"github.com/yourusername/gym-management/subscriptions-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlanService - Servicio de lógica de negocio para planes
type PlanService struct {
	planRepo repository.PlanRepository // Inyección de Dependencias (Interface)
}

// NewPlanService - Constructor con DI
func NewPlanService(planRepo repository.PlanRepository) *PlanService {
	return &PlanService{
		planRepo: planRepo,
	}
}

// CreatePlan - Crea un nuevo plan
func (s *PlanService) CreatePlan(ctx context.Context, req dtos.CreatePlanRequest) (*dtos.PlanResponse, error) {
	// Mapear DTO a entidad
	plan := &entities.Plan{
		ID:                    primitive.NewObjectID(),
		Nombre:                req.Nombre,
		Descripcion:           req.Descripcion,
		PrecioMensual:         req.PrecioMensual,
		TipoAcceso:            req.TipoAcceso,
		DuracionDias:          req.DuracionDias,
		Activo:                req.Activo,
		ActividadesPermitidas: req.ActividadesPermitidas,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Guardar en repositorio
	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}

	// Mapear entidad a DTO de respuesta
	return s.mapPlanToResponse(plan), nil
}

// GetPlanByID - Obtiene un plan por ID
func (s *PlanService) GetPlanByID(ctx context.Context, id string) (*dtos.PlanResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	plan, err := s.planRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	return s.mapPlanToResponse(plan), nil
}

// ListPlans - Lista planes con filtros y paginación real
func (s *PlanService) ListPlans(ctx context.Context, query dtos.ListPlansQuery) (*dtos.PaginatedPlansResponse, error) {
	// Construir filtros
	filters := make(map[string]interface{})
	if query.Activo != nil {
		filters["activo"] = *query.Activo
	}

	// Valores por defecto para paginación
	page := int64(query.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int64(query.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Límite máximo
	}

	// Valores por defecto para ordenamiento
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at" // Ordenar por fecha de creación por defecto
	}

	// Obtener total de registros (ANTES de paginar)
	total, err := s.planRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Obtener planes paginados con PAGINACIÓN REAL
	plansList, err := s.planRepo.FindAllPaginated(ctx, filters, page, pageSize, sortBy, query.SortDesc)
	if err != nil {
		return nil, err
	}

	// Mapear a DTOs
	var plans []dtos.PlanResponse
	for _, plan := range plansList {
		plans = append(plans, *s.mapPlanToResponse(plan))
	}

	// Calcular total de páginas
	totalPages := int(total) / int(pageSize)
	if int(total)%int(pageSize) > 0 {
		totalPages++
	}

	return &dtos.PaginatedPlansResponse{
		Plans:      plans,
		Total:      int(total),
		Page:       int(page),
		PageSize:   int(pageSize),
		TotalPages: totalPages,
	}, nil
}

// UpdatePlan - Actualiza un plan existente
func (s *PlanService) UpdatePlan(ctx context.Context, id string, req dtos.UpdatePlanRequest) (*dtos.PlanResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Obtener plan existente
	plan, err := s.planRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si están presentes
	if req.Nombre != nil {
		plan.Nombre = *req.Nombre
	}
	if req.Descripcion != nil {
		plan.Descripcion = *req.Descripcion
	}
	if req.PrecioMensual != nil {
		plan.PrecioMensual = *req.PrecioMensual
	}
	if req.TipoAcceso != nil {
		plan.TipoAcceso = *req.TipoAcceso
	}
	if req.DuracionDias != nil {
		plan.DuracionDias = *req.DuracionDias
	}
	if req.Activo != nil {
		plan.Activo = *req.Activo
	}
	if req.ActividadesPermitidas != nil {
		plan.ActividadesPermitidas = *req.ActividadesPermitidas
	}

	plan.UpdatedAt = time.Now()

	// Guardar cambios
	if err := s.planRepo.Update(ctx, objID, plan); err != nil {
		return nil, err
	}

	return s.mapPlanToResponse(plan), nil
}

// DeletePlan - Elimina un plan
func (s *PlanService) DeletePlan(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	return s.planRepo.Delete(ctx, objID)
}

// TogglePlanStatus - Activa o desactiva un plan
func (s *PlanService) TogglePlanStatus(ctx context.Context, id string, activo bool) (*dtos.PlanResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Obtener plan existente
	plan, err := s.planRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Actualizar estado
	plan.Activo = activo
	plan.UpdatedAt = time.Now()

	// Guardar cambios
	if err := s.planRepo.Update(ctx, objID, plan); err != nil {
		return nil, err
	}

	return s.mapPlanToResponse(plan), nil
}

// mapPlanToResponse - Helper para mapear entidad a DTO
func (s *PlanService) mapPlanToResponse(plan *entities.Plan) *dtos.PlanResponse {
	return &dtos.PlanResponse{
		ID:                    plan.ID.Hex(),
		Nombre:                plan.Nombre,
		Descripcion:           plan.Descripcion,
		PrecioMensual:         plan.PrecioMensual,
		TipoAcceso:            plan.TipoAcceso,
		DuracionDias:          plan.DuracionDias,
		Activo:                plan.Activo,
		ActividadesPermitidas: plan.ActividadesPermitidas,
		CreatedAt:             plan.CreatedAt,
		UpdatedAt:             plan.UpdatedAt,
	}
}
