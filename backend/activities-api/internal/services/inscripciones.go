package services

import (
	"activities-api/internal/domain"
	"activities-api/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// InscripcionesService define la interfaz del servicio de inscripciones
type InscripcionesService interface {
	ListByUser(ctx context.Context, usuarioID uint) ([]domain.InscripcionResponse, error)
	Create(ctx context.Context, usuarioID, actividadID uint, authToken string) (domain.InscripcionResponse, error)
	Deactivate(ctx context.Context, usuarioID, actividadID uint) error
	DeactivateAllByUser(ctx context.Context, usuarioID uint) (int, error)
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
// IMPLEMENTA PROCESAMIENTO CONCURRENTE con Go Routines, Errgroup y Context
func (s *InscripcionesServiceImpl) Create(ctx context.Context, usuarioID, actividadID uint, authToken string) (domain.InscripcionResponse, error) {
	// Crear contexto con timeout de 10 segundos para todas las validaciones
	validationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// ERRGROUP - Manejo coordinado de goroutines con cancelación automática
	// Si una goroutine falla, las demás son canceladas automáticamente
	g, gCtx := errgroup.WithContext(validationCtx)

	// Variables compartidas para almacenar resultados (protegidas por errgroup)
	var actividadValidada *domain.Actividad

	// Goroutine 1: Validar que la actividad existe
	g.Go(func() error {
		actividad, err := s.actividadesRepo.GetByID(gCtx, actividadID)
		if err != nil {
			return fmt.Errorf("actividad no encontrada: %w", err)
		}
		actividadValidada = &actividad
		return nil
	})

	// Goroutine 2: Validar que el usuario no tenga inscripciones duplicadas
	g.Go(func() error {
		inscripciones, err := s.inscripcionesRepo.ListByUser(gCtx, usuarioID)
		if err != nil {
			return fmt.Errorf("error verificando inscripciones: %w", err)
		}

		// Verificar si ya está inscripto
		for _, insc := range inscripciones {
			if insc.ActividadID == actividadID && insc.IsActiva {
				return fmt.Errorf("el usuario ya está inscripto a esta actividad")
			}
		}
		return nil
	})

	// Goroutine 3: Validar disponibilidad de cupos del usuario
	g.Go(func() error {
		// Esta validación se podría expandir para verificar cupos en tiempo real
		// o hacer cálculos más complejos
		inscripcionesActivas, err := s.inscripcionesRepo.ListByUser(gCtx, usuarioID)
		if err != nil {
			return fmt.Errorf("error verificando disponibilidad: %w", err)
		}

		// Simular cálculo de disponibilidad (cantidad de inscripciones activas del usuario)
		conteoActivas := 0
		for _, insc := range inscripcionesActivas {
			if insc.IsActiva {
				conteoActivas++
			}
		}

		// Ejemplo: podríamos validar un límite máximo aquí
		// if conteoActivas >= MAX_INSCRIPCIONES {
		//     return fmt.Errorf("límite de inscripciones alcanzado")
		// }
		return nil
	})

	// Esperar a que todas las validaciones terminen
	// Si alguna falla, Wait() retorna el primer error y cancela las demás
	if err := g.Wait(); err != nil {
		// Manejar error de timeout específicamente
		if validationCtx.Err() == context.DeadlineExceeded {
			return domain.InscripcionResponse{}, fmt.Errorf("timeout en validaciones: las validaciones tardaron más de 10 segundos")
		}
		return domain.InscripcionResponse{}, err
	}

	// Validación adicional: verificar que obtuvimos la actividad
	if actividadValidada == nil {
		return domain.InscripcionResponse{}, fmt.Errorf("error en validación de actividad")
	}

	// Validación HTTP - Validar suscripción activa (HTTP call to subscriptions-api)
	// Crear contexto con timeout específico para llamadas HTTP (5 segundos)
	httpCtx, httpCancel := context.WithTimeout(ctx, 5*time.Second)
	defer httpCancel()

	activeSub, err := s.getActiveSubscription(httpCtx, usuarioID, authToken)
	if err != nil {
		// Manejar timeout específicamente
		if httpCtx.Err() == context.DeadlineExceeded {
			return domain.InscripcionResponse{}, fmt.Errorf("timeout validando suscripción: el servicio tardó más de 5 segundos en responder")
		}
		return domain.InscripcionResponse{}, fmt.Errorf("debe tener un plan para inscribirse a esta actividad")
	}

	// Validar restricciones del plan - Verificar si la actividad está permitida
	if err := s.validatePlanRestrictions(activeSub, actividadValidada); err != nil {
		return domain.InscripcionResponse{}, err
	}

	// Validar límite de actividades semanales del plan
	if err := s.validateWeeklyActivityLimit(ctx, usuarioID, activeSub); err != nil {
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

	// Invalidar cache de actividades para que se reflejen los nuevos cupos
	if s.actividadesRepo != nil {
		s.actividadesRepo.InvalidateCache()
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

	// Invalidar cache de actividades para reflejar cupos liberados
	if s.actividadesRepo != nil {
		s.actividadesRepo.InvalidateCache()
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
	ActividadesPorSemana  int      `json:"actividades_por_semana"` // Límite de actividades por semana (0 = ilimitado)
}

// getActiveSubscription valida que el usuario tenga una suscripción activa
func (s *InscripcionesServiceImpl) getActiveSubscription(ctx context.Context, userID uint, authToken string) (Subscription, error) {
	// Crear cliente HTTP sin timeout hardcoded (usa el contexto)
	client := &http.Client{}

	// Construir URL usando el service name de Docker
	url := fmt.Sprintf("http://subscriptions-api:8081/subscriptions/active/%d", userID)

	// Crear request con contexto (respeta timeout del contexto padre)
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
		return Subscription{}, fmt.Errorf("debe tener un plan para inscribirse a esta actividad")
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
	// Crear cliente HTTP sin timeout hardcoded (usa el contexto)
	client := &http.Client{}

	url := fmt.Sprintf("http://subscriptions-api:8081/plans/%s", planID)

	// Crear request con contexto (respeta timeout del contexto padre)
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

// validateWeeklyActivityLimit valida que el usuario no exceda el límite de actividades por semana de su plan
func (s *InscripcionesServiceImpl) validateWeeklyActivityLimit(ctx context.Context, usuarioID uint, subscription Subscription) error {
	// Si el plan no tiene límite (0 o tipo_acceso completo), permitir
	if subscription.PlanInfo.ActividadesPorSemana == 0 || subscription.PlanInfo.TipoAcceso == "completo" {
		fmt.Printf("✅ [validateWeeklyActivityLimit] Sin límite semanal (ActividadesPorSemana: %d, TipoAcceso: %s)\n", subscription.PlanInfo.ActividadesPorSemana, subscription.PlanInfo.TipoAcceso)
		return nil
	}

	// Obtener inscripciones activas del usuario
	inscripcionesActivas, err := s.inscripcionesRepo.ListByUser(ctx, usuarioID)
	if err != nil {
		return fmt.Errorf("error verificando inscripciones activas: %w", err)
	}

	// Calcular el inicio de la semana actual (Lunes a las 00:00:00)
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // Domingo
		weekday = 7
	}
	// Restar días para llegar al lunes
	startOfWeek := now.AddDate(0, 0, -(weekday - 1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	fmt.Printf("📅 [validateWeeklyActivityLimit] Inicio de semana: %s\n", startOfWeek.Format("2006-01-02 15:04:05"))

	// Contar inscripciones activas de esta semana
	inscripcionesEstaSemana := 0
	for _, insc := range inscripcionesActivas {
		if insc.IsActiva && insc.FechaInscripcion.After(startOfWeek) {
			inscripcionesEstaSemana++
			fmt.Printf("📝 [validateWeeklyActivityLimit] Inscripción #%d (actividad: %d) en esta semana: %s\n",
				insc.ID, insc.ActividadID, insc.FechaInscripcion.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Printf("📊 [validateWeeklyActivityLimit] Inscripciones esta semana: %d, Límite del plan: %d\n",
		inscripcionesEstaSemana, subscription.PlanInfo.ActividadesPorSemana)

	// Verificar si ya alcanzó el límite
	if inscripcionesEstaSemana >= subscription.PlanInfo.ActividadesPorSemana {
		return fmt.Errorf(
			"has alcanzado el límite de actividades semanales de tu plan '%s' (%d/%d actividades). Espera a la próxima semana o mejora tu plan",
			subscription.PlanInfo.Nombre,
			inscripcionesEstaSemana,
			subscription.PlanInfo.ActividadesPorSemana,
		)
	}

	fmt.Printf("✅ [validateWeeklyActivityLimit] Dentro del límite (%d/%d)\n",
		inscripcionesEstaSemana+1, subscription.PlanInfo.ActividadesPorSemana)
	return nil
}

// DeactivateAllByUser desactiva todas las inscripciones de un usuario
// Se llama cuando se cancela la suscripción del usuario
func (s *InscripcionesServiceImpl) DeactivateAllByUser(ctx context.Context, usuarioID uint) (int, error) {
	fmt.Printf("🔄 [DeactivateAllByUser] Desactivando todas las inscripciones del usuario %d\n", usuarioID)

	// Obtener todas las inscripciones activas del usuario
	inscripciones, err := s.inscripcionesRepo.ListByUser(ctx, usuarioID)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo inscripciones: %w", err)
	}

	// Contar inscripciones activas
	count := 0
	for _, insc := range inscripciones {
		if insc.IsActiva {
			// Desactivar cada inscripción
			if err := s.inscripcionesRepo.Deactivate(ctx, usuarioID, insc.ActividadID); err != nil {
				fmt.Printf("⚠️ [DeactivateAllByUser] Error desactivando inscripción actividad %d: %v\n", insc.ActividadID, err)
				continue
			}
			count++
			fmt.Printf("✅ [DeactivateAllByUser] Inscripción desactivada - Actividad ID: %d\n", insc.ActividadID)

			// Publicar evento para cada desinscripción
			eventData := map[string]interface{}{
				"usuario_id":   usuarioID,
				"actividad_id": insc.ActividadID,
				"reason":       "subscription_cancelled",
			}
			inscripcionID := fmt.Sprintf("%d_%d", usuarioID, insc.ActividadID)
			if err := s.eventPublisher.PublishInscriptionEvent("delete", inscripcionID, eventData); err != nil {
				fmt.Printf("⚠️ [DeactivateAllByUser] Error publicando evento: %v\n", err)
			}
		}
	}

	// Invalidar cache de actividades para reflejar cupos liberados
	if s.actividadesRepo != nil && count > 0 {
		s.actividadesRepo.InvalidateCache()
		fmt.Printf("🔄 [DeactivateAllByUser] Cache invalidado\n")
	}

	fmt.Printf("✅ [DeactivateAllByUser] Total inscripciones desactivadas para usuario %d: %d\n", usuarioID, count)
	return count, nil
}
