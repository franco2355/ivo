package services

import (
	"activities-api/internal/domain"
	"activities-api/internal/repository"
	"context"
	"fmt"
	"time"
)

// ActividadesService define la interfaz del servicio de actividades
type ActividadesService interface {
	List(ctx context.Context) ([]domain.ActividadResponse, error)
	GetByID(ctx context.Context, id uint) (domain.ActividadResponse, error)
	Search(ctx context.Context, params map[string]interface{}) ([]domain.ActividadResponse, error)
	Create(ctx context.Context, actividadCreate domain.ActividadCreate) (domain.ActividadResponse, error)
	Update(ctx context.Context, id uint, actividadUpdate domain.ActividadUpdate) (domain.ActividadResponse, error)
	Delete(ctx context.Context, id uint) error
}

// ActividadesServiceImpl implementa ActividadesService
// Migrado de backend/services/actividad_service.go con dependency injection
type ActividadesServiceImpl struct {
	repository     repository.ActividadesRepository
	eventPublisher EventPublisher
}

// NewActividadesService crea una nueva instancia del servicio
func NewActividadesService(repo repository.ActividadesRepository, eventPublisher EventPublisher) *ActividadesServiceImpl {
	return &ActividadesServiceImpl{
		repository:     repo,
		eventPublisher: eventPublisher,
	}
}

// List obtiene todas las actividades
// Migrado de backend/services/actividad_service.go:92
func (s *ActividadesServiceImpl) List(ctx context.Context) ([]domain.ActividadResponse, error) {
	actividades, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing actividades: %w", err)
	}

	// Convertir a Response DTO
	responses := make([]domain.ActividadResponse, len(actividades))
	for i, act := range actividades {
		responses[i] = act.ToResponse()
	}

	return responses, nil
}

// GetByID obtiene una actividad por ID
// Migrado de backend/services/actividad_service.go:114
func (s *ActividadesServiceImpl) GetByID(ctx context.Context, id uint) (domain.ActividadResponse, error) {
	actividad, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.ActividadResponse{}, fmt.Errorf("actividad con ID %d no encontrada: %w", id, err)
	}

	return actividad.ToResponse(), nil
}

// Search busca actividades por parámetros
// Migrado de backend/services/actividad_service.go:103
func (s *ActividadesServiceImpl) Search(ctx context.Context, params map[string]interface{}) ([]domain.ActividadResponse, error) {
	actividades, err := s.repository.GetByParams(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error searching actividades: %w", err)
	}

	// Convertir a Response DTO
	responses := make([]domain.ActividadResponse, len(actividades))
	for i, act := range actividades {
		responses[i] = act.ToResponse()
	}

	return responses, nil
}

// Create crea una nueva actividad
// Migrado de backend/services/actividad_service.go:123
func (s *ActividadesServiceImpl) Create(ctx context.Context, actividadCreate domain.ActividadCreate) (domain.ActividadResponse, error) {
	// Validar campos básicos
	if err := s.validateBasicFields(actividadCreate); err != nil {
		return domain.ActividadResponse{}, err
	}

	// Parsear horarios
	horaInicio, horaFin, err := s.parseHorarios(actividadCreate.HorarioInicio, actividadCreate.HorarioFinal)
	if err != nil {
		return domain.ActividadResponse{}, err
	}

	// Crear dominio
	actividad := domain.Actividad{
		Titulo:        actividadCreate.Titulo,
		Descripcion:   actividadCreate.Descripcion,
		Cupo:          actividadCreate.Cupo,
		Dia:           actividadCreate.Dia,
		HorarioInicio: actividadCreate.HorarioInicio,
		HorarioFinal:  actividadCreate.HorarioFinal,
		FotoUrl:       actividadCreate.FotoUrl,
		Instructor:    actividadCreate.Instructor,
		Categoria:     actividadCreate.Categoria,
		SucursalID:    actividadCreate.SucursalID,
	}

	createdActividad, err := s.repository.Create(ctx, actividad, horaInicio, horaFin)
	if err != nil {
		return domain.ActividadResponse{}, fmt.Errorf("error creating actividad: %w", err)
	}

	// Publicar evento a RabbitMQ
	eventData := map[string]interface{}{
		"titulo":      createdActividad.Titulo,
		"descripcion": createdActividad.Descripcion,
		"categoria":   createdActividad.Categoria,
		"dia":         createdActividad.Dia,
		"instructor":  createdActividad.Instructor,
	}
	if err := s.eventPublisher.PublishActivityEvent("create", fmt.Sprintf("%d", createdActividad.ID), eventData); err != nil {
		// Log el error pero NO fallamos la creación (ya está creada)
		fmt.Printf("⚠️  Error publicando evento activity.create: %v\n", err)
	}

	return createdActividad.ToResponse(), nil
}

// Update actualiza una actividad existente
// Migrado de backend/services/actividad_service.go:151
func (s *ActividadesServiceImpl) Update(ctx context.Context, id uint, actividadUpdate domain.ActividadUpdate) (domain.ActividadResponse, error) {
	// Validar campos básicos
	if err := s.validateBasicFieldsUpdate(actividadUpdate); err != nil {
		return domain.ActividadResponse{}, err
	}

	// Parsear horarios
	horaInicio, horaFin, err := s.parseHorarios(actividadUpdate.HorarioInicio, actividadUpdate.HorarioFinal)
	if err != nil {
		return domain.ActividadResponse{}, err
	}

	// Crear dominio
	actividad := domain.Actividad{
		Titulo:        actividadUpdate.Titulo,
		Descripcion:   actividadUpdate.Descripcion,
		Cupo:          actividadUpdate.Cupo,
		Dia:           actividadUpdate.Dia,
		HorarioInicio: actividadUpdate.HorarioInicio,
		HorarioFinal:  actividadUpdate.HorarioFinal,
		FotoUrl:       actividadUpdate.FotoUrl,
		Instructor:    actividadUpdate.Instructor,
		Categoria:     actividadUpdate.Categoria,
		SucursalID:    actividadUpdate.SucursalID,
	}

	updatedActividad, err := s.repository.Update(ctx, id, actividad, horaInicio, horaFin)
	if err != nil {
		return domain.ActividadResponse{}, fmt.Errorf("error updating actividad: %w", err)
	}

	// Publicar evento a RabbitMQ
	eventData := map[string]interface{}{
		"titulo":      updatedActividad.Titulo,
		"descripcion": updatedActividad.Descripcion,
		"categoria":   updatedActividad.Categoria,
		"dia":         updatedActividad.Dia,
		"instructor":  updatedActividad.Instructor,
	}
	if err := s.eventPublisher.PublishActivityEvent("update", fmt.Sprintf("%d", updatedActividad.ID), eventData); err != nil {
		// Log el error pero NO fallamos la actualización (ya está actualizada)
		fmt.Printf("⚠️  Error publicando evento activity.update: %v\n", err)
	}

	return updatedActividad.ToResponse(), nil
}

// Delete elimina una actividad
// Migrado de backend/services/actividad_service.go:184
func (s *ActividadesServiceImpl) Delete(ctx context.Context, id uint) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting actividad: %w", err)
	}

	// Publicar evento a RabbitMQ
	eventData := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := s.eventPublisher.PublishActivityEvent("delete", fmt.Sprintf("%d", id), eventData); err != nil {
		// Log el error pero NO fallamos la eliminación (ya está eliminada)
		fmt.Printf("⚠️  Error publicando evento activity.delete: %v\n", err)
	}

	return nil
}

// validateBasicFields valida los campos básicos para crear
// Migrado de backend/services/actividad_service.go:32
func (s *ActividadesServiceImpl) validateBasicFields(actividadCreate domain.ActividadCreate) error {
	if actividadCreate.Cupo == 0 {
		return fmt.Errorf("el cupo debe ser mayor a 0")
	}

	if actividadCreate.Titulo == "" {
		return fmt.Errorf("el título no puede estar vacío")
	}

	if actividadCreate.Dia == "" {
		return fmt.Errorf("el día no puede estar vacío")
	}

	return nil
}

// validateBasicFieldsUpdate valida los campos básicos para actualizar
func (s *ActividadesServiceImpl) validateBasicFieldsUpdate(actividadUpdate domain.ActividadUpdate) error {
	if actividadUpdate.Cupo == 0 {
		return fmt.Errorf("el cupo debe ser mayor a 0")
	}

	if actividadUpdate.Titulo == "" {
		return fmt.Errorf("el título no puede estar vacío")
	}

	if actividadUpdate.Dia == "" {
		return fmt.Errorf("el día no puede estar vacío")
	}

	return nil
}

// parseHorarios parsea horarios en formato "HH:MM" a time.Time
// Migrado de backend/services/actividad_service.go:49
func (s *ActividadesServiceImpl) parseHorarios(horaInicio, horaFin string) (time.Time, time.Time, error) {
	// Obtener la zona horaria local
	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		loc = time.Local // Si no se puede cargar, usar la zona horaria local del sistema
	}

	// Usar una fecha base (2024-01-01) para parsear las horas
	fechaBase := "2024-01-01"
	inicio, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", fechaBase, horaInicio), loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("formato de hora inicio inválido (debe ser HH:MM): %v", err)
	}

	fin, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", fechaBase, horaFin), loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("formato de hora fin inválido (debe ser HH:MM): %v", err)
	}

	// Validar que hora fin sea después de hora inicio
	if fin.Before(inicio) {
		return time.Time{}, time.Time{}, fmt.Errorf("la hora de fin debe ser posterior a la hora de inicio")
	}

	return inicio, fin, nil
}

// TODO: Los compañeros deben agregar:
// - Publicación de eventos a RabbitMQ cuando se crea/actualiza/elimina actividad
// - Validación de sucursal_id existe (HTTP call a activities-api sucursales endpoint cuando esté listo)
