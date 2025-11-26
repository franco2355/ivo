package services

import (
	"testing"
	"time"
)

// MockCacheService - Cache service para tests sin dependencias externas
type MockCacheService struct {
	cache map[string]*mockCacheEntry
	ttl   time.Duration
}

type mockCacheEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewMockCacheService(ttl time.Duration) *MockCacheService {
	return &MockCacheService{
		cache: make(map[string]*mockCacheEntry),
		ttl:   ttl,
	}
}

func (m *MockCacheService) Get(key string) ([]byte, bool) {
	entry, exists := m.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		delete(m.cache, key)
		return nil, false
	}

	return entry.data, true
}

func (m *MockCacheService) Set(key string, value []byte) {
	m.cache[key] = &mockCacheEntry{
		data:      value,
		expiresAt: time.Now().Add(m.ttl),
	}
}

func (m *MockCacheService) Delete(key string) {
	delete(m.cache, key)
}

func (m *MockCacheService) CleanExpired() {
	now := time.Now()
	for key, entry := range m.cache {
		if now.After(entry.expiresAt) {
			delete(m.cache, key)
		}
	}
}

func (m *MockCacheService) Size() int {
	return len(m.cache)
}

// ========== TESTS ==========

func TestCacheService_SetAndGet(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	key := "test_key"
	value := []byte("test_value")

	cache.Set(key, value)

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("El valor debería encontrarse en el cache")
	}

	if string(retrieved) != string(value) {
		t.Errorf("Valor esperado '%s', obtenido '%s'", value, retrieved)
	}
}

func TestCacheService_GetNonExistent(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	_, found := cache.Get("non_existent_key")
	if found {
		t.Error("No debería encontrarse una clave inexistente")
	}
}

func TestCacheService_Delete(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	key := "test_key"
	cache.Set(key, []byte("value"))

	// Verificar que existe
	_, found := cache.Get(key)
	if !found {
		t.Error("El valor debería existir antes de eliminar")
	}

	// Eliminar
	cache.Delete(key)

	// Verificar que ya no existe
	_, found = cache.Get(key)
	if found {
		t.Error("El valor no debería existir después de eliminar")
	}
}

func TestCacheService_DeleteNonExistent(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	// No debería causar error
	cache.Delete("non_existent_key")
}

func TestCacheService_Expiration(t *testing.T) {
	// Cache con TTL muy corto
	cache := NewMockCacheService(50 * time.Millisecond)

	key := "expiring_key"
	cache.Set(key, []byte("value"))

	// Verificar que existe inmediatamente
	_, found := cache.Get(key)
	if !found {
		t.Error("El valor debería existir inmediatamente")
	}

	// Esperar a que expire
	time.Sleep(100 * time.Millisecond)

	// Verificar que expiró
	_, found = cache.Get(key)
	if found {
		t.Error("El valor debería haber expirado")
	}
}

func TestCacheService_CleanExpired(t *testing.T) {
	cache := NewMockCacheService(50 * time.Millisecond)

	// Agregar varias entradas
	cache.Set("key1", []byte("value1"))
	cache.Set("key2", []byte("value2"))
	cache.Set("key3", []byte("value3"))

	if cache.Size() != 3 {
		t.Errorf("Se esperaban 3 entradas, hay %d", cache.Size())
	}

	// Esperar a que expiren
	time.Sleep(100 * time.Millisecond)

	// Limpiar expirados
	cache.CleanExpired()

	if cache.Size() != 0 {
		t.Errorf("Se esperaban 0 entradas después de limpiar, hay %d", cache.Size())
	}
}

func TestCacheService_OverwriteKey(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	key := "test_key"
	cache.Set(key, []byte("original_value"))
	cache.Set(key, []byte("new_value"))

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("El valor debería existir")
	}

	if string(retrieved) != "new_value" {
		t.Errorf("Valor esperado 'new_value', obtenido '%s'", retrieved)
	}

	if cache.Size() != 1 {
		t.Errorf("Debería haber solo 1 entrada, hay %d", cache.Size())
	}
}

func TestCacheService_MultipleKeys(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	cache.Set("key1", []byte("value1"))
	cache.Set("key2", []byte("value2"))
	cache.Set("key3", []byte("value3"))

	if cache.Size() != 3 {
		t.Errorf("Se esperaban 3 entradas, hay %d", cache.Size())
	}

	val1, _ := cache.Get("key1")
	val2, _ := cache.Get("key2")
	val3, _ := cache.Get("key3")

	if string(val1) != "value1" {
		t.Errorf("value1 incorrecto: %s", val1)
	}
	if string(val2) != "value2" {
		t.Errorf("value2 incorrecto: %s", val2)
	}
	if string(val3) != "value3" {
		t.Errorf("value3 incorrecto: %s", val3)
	}
}

func TestCacheService_EmptyValue(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	cache.Set("empty_key", []byte{})

	retrieved, found := cache.Get("empty_key")
	if !found {
		t.Error("Valor vacío debería encontrarse")
	}

	if len(retrieved) != 0 {
		t.Error("El valor debería estar vacío")
	}
}

func TestCacheService_LargeValue(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	// Crear un valor grande (1MB)
	largeValue := make([]byte, 1024*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	cache.Set("large_key", largeValue)

	retrieved, found := cache.Get("large_key")
	if !found {
		t.Error("Valor grande debería encontrarse")
	}

	if len(retrieved) != len(largeValue) {
		t.Errorf("Tamaño esperado %d, obtenido %d", len(largeValue), len(retrieved))
	}
}

func TestCacheService_SpecialCharactersInKey(t *testing.T) {
	cache := NewMockCacheService(time.Minute)

	specialKeys := []string{
		"key:with:colons",
		"key/with/slashes",
		"key-with-dashes",
		"key_with_underscores",
		"key.with.dots",
		"key with spaces",
	}

	for _, key := range specialKeys {
		cache.Set(key, []byte("value"))

		_, found := cache.Get(key)
		if !found {
			t.Errorf("Clave '%s' debería encontrarse", key)
		}
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		obj    interface{}
	}{
		{"simple string", "search", "yoga"},
		{"map", "search", map[string]string{"query": "yoga", "categoria": "fitness"}},
		{"struct", "search", struct{ Query string }{Query: "yoga"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GenerateCacheKey(tt.prefix, tt.obj)

			if key == "" {
				t.Error("La clave generada no debería estar vacía")
			}

			// Verificar que el prefijo está presente
			if len(key) < len(tt.prefix) || key[:len(tt.prefix)] != tt.prefix {
				t.Errorf("La clave debería comenzar con '%s', obtenida '%s'", tt.prefix, key)
			}

			// Verificar que la clave es consistente
			key2 := GenerateCacheKey(tt.prefix, tt.obj)
			if key != key2 {
				t.Error("La misma entrada debería generar la misma clave")
			}
		})
	}
}

func TestGenerateCacheKey_DifferentInputs(t *testing.T) {
	key1 := GenerateCacheKey("search", map[string]string{"query": "yoga"})
	key2 := GenerateCacheKey("search", map[string]string{"query": "spinning"})

	if key1 == key2 {
		t.Error("Diferentes inputs deberían generar diferentes claves")
	}
}

func TestGenerateCacheKey_DifferentPrefixes(t *testing.T) {
	obj := map[string]string{"query": "yoga"}

	key1 := GenerateCacheKey("search", obj)
	key2 := GenerateCacheKey("activities", obj)

	if key1 == key2 {
		t.Error("Diferentes prefijos deberían generar diferentes claves")
	}
}
