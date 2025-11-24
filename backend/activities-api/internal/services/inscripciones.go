package services

import (
	"activities-api/internal/domain"
	"activities-api/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	// Validación HTTP - Validar suscripción activa (HTTP call to subscriptions-api)
	activeSub, err := s.getActiveSubscription(ctx, usuarioID, authToken)
	if err != nil {
		return domain.InscripcionResponse{}, fmt.Errorf("no tiene suscripción activa: %w", err)
	}

	// Validar restricciones del plan - Verificar si la actividad está permitida
	if err := s.validatePlanRestrictions(activeSub, actividadValidada); err != nil {
		return domain.InscripcionResponse{}, err
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
	ID       string `json:"id"`
	UserID   string `json:"usuario_id"` // Cambiado de uint a string para coincidir con la respuesta de subscriptions-api
	PlanID   string `json:"plan_id"`
	Status   string `json:"estado"`
	PlanInfo Plan   `json:"plan_info,omitempty"` // Info del plan expandida
}

// Plan representa la información del plan de suscripción
type Plan struct {
	ID                    string   `json:"id"`
	Nombre                string   `json:"nombre"`
	TipoAcceso            string   `json:"tipo_acceso"` // "limitado" | "completo"
	ActividadesPermitidas []string `json:"actividades_permitidas"`
}

// getActiveSubscription valida que el usuario tenga una suscripción activa
func (s *InscripcionesServiceImpl) getActiveSubscription(ctx context.Context, userID uint, authToken string) (Subscription, error) {
	// Crear cliente HTTP con timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Construir URL usando el service name de Docker
	url := fmt.Sprintf("http://subscriptions-api:8081/subscriptions/active/%d", userID)

	// Crear request con contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Subscription{}, fmt.Errorf("error creando request: %w", err)
	}

	// Agregar header de autorización
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
		fmt.Printf("🔐 [getActiveSubscription] Usando token de autorización para usuario %d\n", userID)
	} else {
		fmt.Printf("⚠️  [getActiveSubscription] NO se proporcionó token de autorización para usuario %d\n", userID)
	}

	// Ejecutar request
	fmt.Printf("🌐 [getActiveSubscription] Llamando a: %s\n", url)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ [getActiveSubscription] Error ejecutando request: %v\n", err)
		return Subscription{}, fmt.Errorf("error llamando a subscriptions-api: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📊 [getActiveSubscription] Status Code recibido: %d\n", resp.StatusCode)

	// Manejar diferentes códigos de estado
	if resp.StatusCode == 401 {
		fmt.Printf("❌ [getActiveSubscription] Error 401 - No autorizado\n")
		return Subscription{}, fmt.Errorf("no autorizado para consultar suscripción (falta token de autenticación)")
	}
	if resp.StatusCode == 404 {
		fmt.Printf("❌ [getActiveSubscription] Error 404 - Suscripción no encontrada\n")
		return Subscription{}, fmt.Errorf("no se encontró suscripción activa")
	}
	if resp.StatusCode != 200 {
		// Leer el body para obtener más detalles del error
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ [getActiveSubscription] Error HTTP %d desde subscriptions-api: %s\n", resp.StatusCode, string(bodyBytes))
		return Subscription{}, fmt.Errorf("error obteniendo suscripción activa (status: %d)", resp.StatusCode)
	}

	// Decodificar respuesta
	var subscription Subscription
	if err := json.NewDecoder(resp.Body).Decode(&subscription); err != nil {
		fmt.Printf("❌ [getActiveSubscription] Error decodificando JSON: %v\n", err)
		return Subscription{}, fmt.Errorf("error decodificando suscripción: %w", err)
	}

	fmt.Printf("📦 [getActiveSubscription] Suscripción decodificada - ID: %s, UserID: %s, PlanID: %s, Status: %s\n",
		subscription.ID, subscription.UserID, subscription.PlanID, subscription.Status)

	// Validar que la suscripción esté activa
	if subscription.Status != "activa" {
		fmt.Printf("❌ [getActiveSubscription] Suscripción no activa - Estado: %s\n", subscription.Status)
		return Subscription{}, fmt.Errorf("la suscripción del usuario no está activa (estado: %s)", subscription.Status)
	}

	// Obtener información del plan
	fmt.Printf("📋 [getActiveSubscription] Obteniendo info del plan: %s\n", subscription.PlanID)
	planInfo, err := s.getPlanInfo(ctx, subscription.PlanID, authToken)
	if err != nil {
		// Log el error pero no fallamos - el plan puede no estar disponible temporalmente
		fmt.Printf("⚠️  [getActiveSubscription] No se pudo obtener info del plan %s: %v\n", subscription.PlanID, err)
	} else {
		subscription.PlanInfo = planInfo
		fmt.Printf("✅ [getActiveSubscription] Plan cargado: %s (tipo_acceso: %s, actividades: %v)\n", planInfo.Nombre, planInfo.TipoAcceso, planInfo.ActividadesPermitidas)
	}

	fmt.Printf("✅ [getActiveSubscription] Suscripción obtenida exitosamente para usuario %d (Estado: %s)\n", userID, subscription.Status)
	return subscription, nil
}

// getPlanInfo obtiene la información del plan desde subscriptions-api
func (s *InscripcionesServiceImpl) getPlanInfo(ctx context.Context, planID string, authToken string) (Plan, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://subscriptions-api:8081/plans/%s", planID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Plan{}, fmt.Errorf("error creando request: %w", err)
	}

	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Plan{}, fmt.Errorf("error llamando a subscriptions-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Plan{}, fmt.Errorf("error obteniendo plan (status: %d)", resp.StatusCode)
	}

	var plan Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return Plan{}, fmt.Errorf("error decodificando plan: %w", err)
	}

	return plan, nil
}

// validatePlanRestrictions valida que la actividad esté permitida por el plan del usuario
func (s *InscripcionesServiceImpl) validatePlanRestrictions(subscription Subscription, actividad *domain.Actividad) error {
	fmt.Printf("🔍 [validatePlanRestrictions] Validando restricciones para actividad '%s' (categoría: %s)\n", actividad.Titulo, actividad.Categoria)
	fmt.Printf("🔍 [validatePlanRestrictions] Plan: %s, TipoAcceso: %s, ActividadesPermitidas: %v\n", subscription.PlanInfo.Nombre, subscription.PlanInfo.TipoAcceso, subscription.PlanInfo.ActividadesPermitidas)

	// Si el plan tiene tipo_acceso "completo", permitir cualquier actividad
	if subscription.PlanInfo.TipoAcceso == "completo" {
		fmt.Printf("✅ [validatePlanRestrictions] Plan completo - actividad permitida\n")
		return nil
	}

	// Si el plan tiene tipo_acceso "limitado", validar actividades permitidas
	if subscription.PlanInfo.TipoAcceso == "limitado" {
		// Verificar si la categoría de la actividad está en las actividades permitidas
		actividadPermitida := false
		for _, categoriaPermitida := range subscription.PlanInfo.ActividadesPermitidas {
			if categoriaPermitida == actividad.Categoria {
				actividadPermitida = true
				break
			}
		}

		if !actividadPermitida {
			return fmt.Errorf(
				"tu plan '%s' no incluye la categoría '%s'. Actualiza tu plan para acceder a esta actividad",
				subscription.PlanInfo.Nombre,
				actividad.Categoria,
			)
		}
	}

	return nil
}
