package services

import (
	"activities-api/internal/domain"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// ActividadesCacheService - Servicio con cache in-memory para actividades
// Cachea la lista completa de actividades que es la query más frecuente
type ActividadesCacheService struct {
	baseService *ActividadesServiceImpl
	cache       map[string]*actividadCacheEntry
	mu          sync.RWMutex
	ttl         time.Duration
}

type actividadCacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// NewActividadesCacheService - Constructor con cache habilitado
// TTL: 5 minutos (las actividades cambian ocasionalmente)
func NewActividadesCacheService(baseService *ActividadesServiceImpl) *ActividadesCacheService {
	service := &ActividadesCacheService{
		baseService: baseService,
		cache:       make(map[string]*actividadCacheEntry),
		ttl:         5 * time.Minute, // TTL de 5 minutos
	}

	// Iniciar rutina de limpieza periódica cada 5 minutos
	go service.startCleanupRoutine(5 * time.Minute)

	log.Println("✅ Cache de actividades inicializado (TTL: 5 minutos)")
	return service
}

// List - Lista todas las actividades CON CACHE
func (s *ActividadesCacheService) List(ctx context.Context) ([]domain.ActividadResponse, error) {
	cacheKey := "actividades:all"

	// Intentar obtener del cache
	if cached, found := s.getFromCache(cacheKey); found {
		var response []domain.ActividadResponse
		if err := json.Unmarshal(cached, &response); err == nil {
			log.Println("✅ CACHE HIT: lista de actividades")
			return response, nil
		}
	}

	// CACHE MISS - obtener de la base de datos
	log.Println("⚠️  CACHE MISS: lista de actividades - consultando BD")
	response, err := s.baseService.List(ctx)
	if err != nil {
		return nil, err
	}

	// Guardar en cache
	if data, err := json.Marshal(response); err == nil {
		s.setInCache(cacheKey, data)
		log.Printf("💾 Lista de actividades guardada en cache (%d actividades)", len(response))
	}

	return response, nil
}

// GetByID - Obtiene una actividad por ID (sin cache para queries específicas)
func (s *ActividadesCacheService) GetByID(ctx context.Context, id uint) (domain.ActividadResponse, error) {
	return s.baseService.GetByID(ctx, id)
}

// Search - Busca actividades por parámetros (sin cache, queries muy variables)
func (s *ActividadesCacheService) Search(ctx context.Context, params map[string]interface{}) ([]domain.ActividadResponse, error) {
	return s.baseService.Search(ctx, params)
}

// Create - Crea una actividad e invalida el cache
func (s *ActividadesCacheService) Create(ctx context.Context, actividadCreate domain.ActividadCreate) (domain.ActividadResponse, error) {
	result, err := s.baseService.Create(ctx, actividadCreate)
	if err == nil {
		s.invalidateCache()
		log.Println("✅ Cache de actividades invalidado (nueva actividad creada)")
	}
	return result, err
}

// Update - Actualiza una actividad e invalida el cache
func (s *ActividadesCacheService) Update(ctx context.Context, id uint, actividadUpdate domain.ActividadUpdate) (domain.ActividadResponse, error) {
	result, err := s.baseService.Update(ctx, id, actividadUpdate)
	if err == nil {
		s.invalidateCache()
		log.Println("✅ Cache de actividades invalidado (actividad actualizada)")
	}
	return result, err
}

// Delete - Elimina una actividad e invalida el cache
func (s *ActividadesCacheService) Delete(ctx context.Context, id uint) error {
	err := s.baseService.Delete(ctx, id)
	if err == nil {
		s.invalidateCache()
		log.Println("✅ Cache de actividades invalidado (actividad eliminada)")
	}
	return err
}

// --- Métodos privados de cache ---

// getFromCache - Obtiene un valor del cache si existe y no ha expirado
func (s *ActividadesCacheService) getFromCache(key string) ([]byte, bool) {
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
func (s *ActividadesCacheService) setInCache(key string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[key] = &actividadCacheEntry{
		data:      data,
		expiresAt: time.Now().Add(s.ttl),
	}
}

// invalidateCache - Invalida todo el cache de actividades
func (s *ActividadesCacheService) invalidateCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Limpiar todo el cache
	s.cache = make(map[string]*actividadCacheEntry)
}

// cleanExpired - Limpia entradas expiradas del cache
func (s *ActividadesCacheService) cleanExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0
	for key, entry := range s.cache {
		if now.After(entry.expiresAt) {
			delete(s.cache, key)
			count++
		}
	}

	if count > 0 {
		log.Printf("🧹 Cache de actividades: %d entradas expiradas eliminadas", count)
	}
}

// startCleanupRoutine - Rutina de limpieza periódica
func (s *ActividadesCacheService) startCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanExpired()
	}
}

// GetCacheStats - Obtiene estadísticas del cache (útil para debugging)
func (s *ActividadesCacheService) GetCacheStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"entries":     len(s.cache),
		"ttl_seconds": s.ttl.Seconds(),
	}
}
