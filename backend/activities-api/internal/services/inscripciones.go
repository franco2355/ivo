package services

import (
	"activities-api/internal/domain"
	"activities-api/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// InscripcionesService define la interfaz del servicio de inscripciones
type InscripcionesService interface {
	ListByUser(ctx context.Context, usuarioID uint) ([]domain.InscripcionResponse, error)
	Create(ctx context.Context, usuarioID, actividadID uint, authToken string) (domain.InscripcionResponse, error)
	Deactivate(ctx context.Context, usuarioID, actividadID uint) error
}

// InscripcionesServiceImpl implementa InscripcionesService
// Migrado de backend/services/inscripcion_service.go con dependency injection
type InscripcionesServiceImpl struct {
	inscripcionesRepo repository.InscripcionesRepository
	actividadesRepo   repository.ActividadesRepository
	eventPublisher    EventPublisher
}

// NewInscripcionesService crea una nueva instancia del servicio
func NewInscripcionesService(inscripcionesRepo repository.InscripcionesRepository, actividadesRepo repository.ActividadesRepository, eventPublisher EventPublisher) *InscripcionesServiceImpl {
	return &InscripcionesServiceImpl{
		inscripcionesRepo: inscripcionesRepo,
		actividadesRepo:   actividadesRepo,
		eventPublisher:    eventPublisher,
	}
}

// ListByUser obtiene todas las inscripciones de un usuario
// Migrado de backend/services/inscripcion_service.go:24
func (s *InscripcionesServiceImpl) ListByUser(ctx context.Context, usuarioID uint) ([]domain.InscripcionResponse, error) {
	inscripciones, err := s.inscripcionesRepo.ListByUser(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error listing inscripciones: %w", err)
	}

	// Convertir a Response DTO
	responses := make([]domain.InscripcionResponse, len(inscripciones))
	for i, insc := range inscripciones {
		responses[i] = insc.ToResponse()
	}

	return responses, nil
}

// ResultadoValidacion representa el resultado de una validación concurrente
type ResultadoValidacion struct {
	Nombre  string
	Exitoso bool
	Error   error
	Datos   interface{}
}

// Create inscribe a un usuario en una actividad
// Migrado de backend/services/inscripcion_service.go:44
// IMPLEMENTA PROCESAMIENTO CONCURRENTE con Go Routines, Channels y WaitGroup
func (s *InscripcionesServiceImpl) Create(ctx context.Context, usuarioID, actividadID uint, authToken string) (domain.InscripcionResponse, error) {
	// PROCESAMIENTO CONCURRENTE - Subdividir validaciones en Go Routines
	// Crear canal para recibir resultados de las validaciones
	canalResultados := make(chan ResultadoValidacion, 3)

	// WaitGroup para sincronizar las goroutines
	var wg sync.WaitGroup

	// Goroutine 1: Validar que la actividad existe
	wg.Add(1)
	go func() {
		defer wg.Done()
		actividad, err := s.actividadesRepo.GetByID(ctx, actividadID)
		if err != nil {
			canalResultados <- ResultadoValidacion{
				Nombre:  "actividad",
				Exitoso: false,
				Error:   fmt.Errorf("actividad no encontrada: %w", err),
			}
			return
		}
		canalResultados <- ResultadoValidacion{
			Nombre:  "actividad",
			Exitoso: true,
			Datos:   actividad,
		}
	}()

	// Goroutine 2: Validar que el usuario no tenga inscripciones duplicadas
	wg.Add(1)
	go func() {
		defer wg.Done()
		inscripciones, err := s.inscripcionesRepo.ListByUser(ctx, usuarioID)
		if err != nil {
			canalResultados <- ResultadoValidacion{
				Nombre:  "duplicados",
				Exitoso: false,
				Error:   fmt.Errorf("error verificando inscripciones: %w", err),
			}
			return
		}

		// Verificar si ya está inscripto
		for _, insc := range inscripciones {
			if insc.ActividadID == actividadID && insc.IsActiva {
				canalResultados <- ResultadoValidacion{
					Nombre:  "duplicados",
					Exitoso: false,
					Error:   fmt.Errorf("el usuario ya está inscripto a esta actividad"),
				}
				return
			}
		}

		canalResultados <- ResultadoValidacion{
			Nombre:  "duplicados",
			Exitoso: true,
		}
	}()

	// Goroutine 3: Calcular disponibilidad de cupos
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Esta validación se podría expandir para verificar cupos en tiempo real
		// o hacer cálculos más complejos
		inscripcionesActivas, err := s.inscripcionesRepo.ListByUser(ctx, usuarioID)
		if err != nil {
			canalResultados <- ResultadoValidacion{
				Nombre:  "disponibilidad",
				Exitoso: false,
				Error:   fmt.Errorf("error verificando disponibilidad: %w", err),
			}
			return
		}

		// Simular cálculo de disponibilidad (cantidad de inscripciones activas del usuario)
		conteoActivas := 0
		for _, insc := range inscripcionesActivas {
			if insc.IsActiva {
				conteoActivas++
			}
		}

		canalResultados <- ResultadoValidacion{
			Nombre:  "disponibilidad",
			Exitoso: true,
			Datos:   conteoActivas,
		}
	}()

	// Cerrar el canal cuando todas las goroutines terminen
	go func() {
		wg.Wait()
		close(canalResultados)
	}()

	// COMUNICACIÓN mediante CHANNEL - Recolectar resultados
	var actividadValidada *domain.Actividad
	erroresValidacion := make(map[string]error)

	// Leer resultados del canal
	for resultado := range canalResultados {
		if !resultado.Exitoso {
			erroresValidacion[resultado.Nombre] = resultado.Error
		} else {
			// Guardar datos importantes
			if resultado.Nombre == "actividad" && resultado.Datos != nil {
				actividad := resultado.Datos.(domain.Actividad)
				actividadValidada = &actividad
			}
		}
	}

	// Verificar si hubo errores en las validaciones
	if len(erroresValidacion) > 0 {
		// Retornar el primer error encontrado
		for _, err := range erroresValidacion {
			return domain.InscripcionResponse{}, err
		}
	}

	// Validación adicional: verificar que obtuvimos la actividad
	if actividadValidada == nil {
		return domain.InscripcionResponse{}, fmt.Errorf("error en validación de actividad")
	}

	// Validación HTTP - Validar suscripción activa (HTTP call a subscriptions-api)
	activeSub, err := s.getActiveSubscription(ctx, usuarioID, authToken)
	if err != nil {
		return domain.InscripcionResponse{}, fmt.Errorf("no tiene suscripción activa: %w", err)
	}

	// Crear inscripción
	inscripcion := domain.Inscripcion{
		UsuarioID:      usuarioID,
		ActividadID:    actividadID,
		IsActiva:       true,
		SuscripcionID:  &activeSub.ID,
	}

	createdInscripcion, err := s.inscripcionesRepo.Create(ctx, inscripcion)
	if err != nil {
		return domain.InscripcionResponse{}, fmt.Errorf("error creating inscripcion: %w", err)
	}

	// Publicar evento a RabbitMQ
	eventData := map[string]interface{}{
		"usuario_id":   createdInscripcion.UsuarioID,
		"actividad_id": createdInscripcion.ActividadID,
		"is_activa":    createdInscripcion.IsActiva,
	}
	if err := s.eventPublisher.PublishInscriptionEvent("create", fmt.Sprintf("%d", createdInscripcion.ID), eventData); err != nil {
		// Log el error pero NO fallamos la inscripción (ya está creada)
		fmt.Printf("⚠️  Error publicando evento inscription.create: %v\n", err)
	}

	return createdInscripcion.ToResponse(), nil
}

// Deactivate desinscribe a un usuario de una actividad
// Migrado de backend/services/inscripcion_service.go:48
func (s *InscripcionesServiceImpl) Deactivate(ctx context.Context, usuarioID, actividadID uint) error {
	if err := s.inscripcionesRepo.Deactivate(ctx, usuarioID, actividadID); err != nil {
		return fmt.Errorf("error deactivating inscripcion: %w", err)
	}

	// Publicar evento a RabbitMQ
	eventData := map[string]interface{}{
		"usuario_id":   usuarioID,
		"actividad_id": actividadID,
	}
	inscripcionID := fmt.Sprintf("%d_%d", usuarioID, actividadID)
	if err := s.eventPublisher.PublishInscriptionEvent("delete", inscripcionID, eventData); err != nil {
		// Log el error pero NO fallamos la operación (ya está desactivada)
		fmt.Printf("⚠️  Error publicando evento inscription.delete: %v\n", err)
	}

	return nil
}

// Subscription representa una suscripción activa del usuario
type Subscription struct {
	ID     string `json:"id"`
	UserID uint   `json:"usuario_id"`
	PlanID string `json:"plan_id"`
	Status string `json:"estado"`
}

// getActiveSubscription valida que el usuario tenga una suscripción activa
func (s *InscripcionesServiceImpl) getActiveSubscription(ctx context.Context, userID uint, authToken string) (Subscription, error) {
	// Crear cliente HTTP con timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Construir URL (usar localhost en desarrollo, cambiar en producción)
	url := fmt.Sprintf("http://localhost:8081/subscriptions/active/%d", userID)

	// Crear request con contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Subscription{}, fmt.Errorf("error creando request: %w", err)
	}

	// Agregar header de autorización
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	// Ejecutar request
	resp, err := client.Do(req)
	if err != nil {
		return Subscription{}, fmt.Errorf("error llamando a subscriptions-api: %w", err)
	}
	defer resp.Body.Close()

	// Manejar diferentes códigos de estado
	if resp.StatusCode == 404 {
		return Subscription{}, fmt.Errorf("no se encontró suscripción activa")
	}
	if resp.StatusCode != 200 {
		return Subscription{}, fmt.Errorf("error obteniendo suscripción activa (status: %d)", resp.StatusCode)
	}

	// Decodificar respuesta
	var subscription Subscription
	if err := json.NewDecoder(resp.Body).Decode(&subscription); err != nil {
		return Subscription{}, fmt.Errorf("error decodificando suscripción: %w", err)
	}

	// Validar que la suscripción esté activa
	if subscription.Status != "activa" {
		return Subscription{}, fmt.Errorf("la suscripción del usuario no está activa (estado: %s)", subscription.Status)
	}

	return subscription, nil
}
