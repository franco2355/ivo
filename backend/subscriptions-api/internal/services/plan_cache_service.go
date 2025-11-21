package services

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/yourusername/gym-management/subscriptions-api/internal/domain/dtos"
)

// PlanCacheService - Servicio con cache in-memory para planes
// Implementa el mismo interface que PlanService pero con cache
type PlanCacheService struct {
	baseService *PlanService
	cache       map[string]*cacheEntry
	mu          sync.RWMutex
	ttl         time.Duration
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// NewPlanCacheService - Constructor con cache habilitado
// TTL por defecto: 1 hora (los planes cambian muy poco)
func NewPlanCacheService(baseService *PlanService) *PlanCacheService {
	service := &PlanCacheService{
		baseService: baseService,
		cache:       make(map[string]*cacheEntry),
		ttl:         1 * time.Hour, // TTL de 1 hora
	}

	// Iniciar rutina de limpieza periódica cada 10 minutos
	go service.startCleanupRoutine(10 * time.Minute)

	return service
}

// CreatePlan - Crea un plan e invalida el cache
func (s *PlanCacheService) CreatePlan(ctx context.Context, req dtos.CreatePlanRequest) (*dtos.PlanResponse, error) {
	result, err := s.baseService.CreatePlan(ctx, req)
	if err == nil {
		s.invalidateCache() // Invalidar todo el cache de planes
		log.Println("✅ Cache de planes invalidado (nuevo plan creado)")
	}
	return result, err
}

// GetPlanByID - Obtiene un plan por ID (sin cache para queries específicas)
func (s *PlanCacheService) GetPlanByID(ctx context.Context, id string) (*dtos.PlanResponse, error) {
	return s.baseService.GetPlanByID(ctx, id)
}

// ListPlans - Lista planes CON CACHE
// Solo cachea la consulta de planes activos sin filtros adicionales
func (s *PlanCacheService) ListPlans(ctx context.Context, query dtos.ListPlansQuery) (*dtos.PaginatedPlansResponse, error) {
	// Solo cachear la consulta de planes activos (la más común)
	if query.Activo != nil && *query.Activo && query.Page == 1 && query.PageSize == 0 {
		cacheKey := "plans:active:all"

		// Intentar obtener del cache
		if cached, found := s.getFromCache(cacheKey); found {
			var response dtos.PaginatedPlansResponse
			if err := json.Unmarshal(cached, &response); err == nil {
				log.Println("✅ CACHE HIT: planes activos")
				return &response, nil
			}
		}

		// CACHE MISS - obtener de la base de datos
		log.Println("⚠️  CACHE MISS: planes activos - consultando BD")
		response, err := s.baseService.ListPlans(ctx, query)
		if err != nil {
			return nil, err
		}

		// Guardar en cache
		if data, err := json.Marshal(response); err == nil {
			s.setInCache(cacheKey, data)
		}

		return response, nil
	}

	// Para otras queries, no usar cache
	return s.baseService.ListPlans(ctx, query)
}

// UpdatePlan - Actualiza un plan e invalida el cache
func (s *PlanCacheService) UpdatePlan(ctx context.Context, id string, req dtos.UpdatePlanRequest) (*dtos.PlanResponse, error) {
	result, err := s.baseService.UpdatePlan(ctx, id, req)
	if err == nil {
		s.invalidateCache()
		log.Println("✅ Cache de planes invalidado (plan actualizado)")
	}
	return result, err
}

// DeletePlan - Elimina un plan e invalida el cache
func (s *PlanCacheService) DeletePlan(ctx context.Context, id string) error {
	err := s.baseService.DeletePlan(ctx, id)
	if err == nil {
		s.invalidateCache()
		log.Println("✅ Cache de planes invalidado (plan eliminado)")
	}
	return err
}

// TogglePlanStatus - Activa/desactiva un plan e invalida el cache
func (s *PlanCacheService) TogglePlanStatus(ctx context.Context, id string, activo bool) (*dtos.PlanResponse, error) {
	result, err := s.baseService.TogglePlanStatus(ctx, id, activo)
	if err == nil {
		s.invalidateCache()
		log.Println("✅ Cache de planes invalidado (estado cambiado)")
	}
	return result, err
}

// --- Métodos privados de cache ---

// getFromCache - Obtiene un valor del cache si existe y no ha expirado
func (s *PlanCacheService) getFromCache(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil, false
	}

	// Verificar si expiró
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.data, true
}

// setInCache - Guarda un valor en el cache con TTL
func (s *PlanCacheService) setInCache(key string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(s.ttl),
	}
}

// invalidateCache - Invalida todo el cache de planes
func (s *PlanCacheService) invalidateCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Limpiar todo el cache
	s.cache = make(map[string]*cacheEntry)
}

// cleanExpired - Limpia entradas expiradas del cache
func (s *PlanCacheService) cleanExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.cache {
		if now.After(entry.expiresAt) {
			delete(s.cache, key)
		}
	}
}

// startCleanupRoutine - Rutina de limpieza periódica
func (s *PlanCacheService) startCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanExpired()
		log.Println("🧹 Cache de planes: limpieza periódica ejecutada")
	}
}

// GetCacheStats - Obtiene estadísticas del cache (útil para debugging)
func (s *PlanCacheService) GetCacheStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"entries": len(s.cache),
		"ttl_seconds": s.ttl.Seconds(),
	}
}
