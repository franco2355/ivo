package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// CacheService maneja el sistema de caché de dos niveles
type CacheService struct {
	memcached     *memcache.Client
	localCache    map[string]*cacheEntry
	localMu       sync.RWMutex
	memcachedTTL  int32
	localCacheTTL time.Duration
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewCacheService(servers []string, memcachedTTL, localCacheTTL int) *CacheService {
	mc := memcache.New(servers...)

	return &CacheService{
		memcached:     mc,
		localCache:    make(map[string]*cacheEntry),
		memcachedTTL:  int32(memcachedTTL),
		localCacheTTL: time.Duration(localCacheTTL) * time.Second,
	}
}

// Get intenta obtener del caché local primero, luego memcached
func (c *CacheService) Get(key string) ([]byte, bool) {
	// 1. Intentar caché local
	c.localMu.RLock()
	entry, exists := c.localCache[key]
	c.localMu.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		return entry.data, true
	}

	// 2. Si no está en local, intentar memcached
	item, err := c.memcached.Get(key)
	if err == nil {
		// Guardar en caché local
		c.setLocal(key, item.Value)
		return item.Value, true
	}

	return nil, false
}

// Set guarda en ambos niveles de caché
func (c *CacheService) Set(key string, value []byte) error {
	// Guardar en memcached
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: c.memcachedTTL,
	}

	if err := c.memcached.Set(item); err != nil {
		// Log the error but continue with local cache
		fmt.Printf("⚠️  Memcached Set failed: %v (using local cache only)\n", err)
	} else {
		fmt.Printf("✅ Memcached Set successful for key: %s\n", key)
	}

	// Guardar en caché local
	c.setLocal(key, value)

	return nil
}

func (c *CacheService) setLocal(key string, value []byte) {
	c.localMu.Lock()
	defer c.localMu.Unlock()

	c.localCache[key] = &cacheEntry{
		data:      value,
		expiresAt: time.Now().Add(c.localCacheTTL),
	}
}

// Delete elimina de ambos niveles de caché
func (c *CacheService) Delete(key string) {
	c.memcached.Delete(key)

	c.localMu.Lock()
	delete(c.localCache, key)
	c.localMu.Unlock()
}

// InvalidatePattern invalida todas las claves que coinciden con un patrón
func (c *CacheService) InvalidatePattern(pattern string) {
	c.localMu.Lock()
	defer c.localMu.Unlock()

	for key := range c.localCache {
		if contains(key, pattern) {
			delete(c.localCache, key)
			c.memcached.Delete(key)
		}
	}
}

// FlushAll invalida TODO el caché (local y memcached)
func (c *CacheService) FlushAll() {
	c.localMu.Lock()
	defer c.localMu.Unlock()

	// Borrar todas las claves conocidas de memcached
	for key := range c.localCache {
		c.memcached.Delete(key)
	}

	// Limpiar caché local
	c.localCache = make(map[string]*cacheEntry)

	// También intentar flush de memcached
	c.memcached.FlushAll()

	fmt.Println("🗑️  Cache flushed (local + memcached)")
}

// GenerateCacheKey genera una clave de caché desde un objeto
func GenerateCacheKey(prefix string, obj interface{}) string {
	data, _ := json.Marshal(obj)
	hash := md5.Sum(data)
	return fmt.Sprintf("%s:%x", prefix, hash)
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && s[:len(substr)] == substr
}

// CleanExpired limpia entradas expiradas del caché local
func (c *CacheService) CleanExpired() {
	c.localMu.Lock()
	defer c.localMu.Unlock()

	now := time.Now()
	for key, entry := range c.localCache {
		if now.After(entry.expiresAt) {
			delete(c.localCache, key)
		}
	}
}

// StartCleanupRoutine inicia una rutina de limpieza periódica
func (c *CacheService) StartCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.CleanExpired()
		}
	}()
}
