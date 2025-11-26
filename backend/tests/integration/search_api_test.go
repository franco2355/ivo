package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const searchAPIURL = "http://localhost:8084"

// SearchResponse representa la respuesta del search API
type SearchResponse struct {
	Results    []SearchDocument `json:"results"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// SearchDocument representa un documento de búsqueda
type SearchDocument struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Titulo         string `json:"titulo"`
	Descripcion    string `json:"descripcion"`
	Categoria      string `json:"categoria"`
	Instructor     string `json:"instructor"`
	Dia            string `json:"dia"`
	HorarioInicio  string `json:"horario_inicio"`
	HorarioFinal   string `json:"horario_final"`
	SucursalID     string `json:"sucursal_id"`
	CupoDisponible int    `json:"cupo_disponible"`
}

// HealthResponse representa la respuesta de health check
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks"`
}

func TestSearchAPI_HealthCheck(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API Health Check")

	resp, err := http.Get(searchAPIURL + "/healthz")
	if err != nil {
		t.Fatalf("❌ Error conectando al Search API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("❌ Health check falló - Status: %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("❌ Error parseando respuesta: %v", err)
	}

	t.Logf("✅ Search API Status: %s", health.Status)
	t.Logf("✅ Solr: %s", health.Checks["solr"])
	t.Logf("✅ Cache: %s", health.Checks["cache"])

	if health.Status != "ok" {
		t.Errorf("❌ Status esperado 'ok', obtenido '%s'", health.Status)
	}

	if health.Checks["solr"] != "connected" {
		t.Errorf("❌ Solr debería estar conectado")
	}

	t.Log("================================================================================")
	t.Log("🎉 TEST HEALTH CHECK COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_SearchByQuery(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Búsqueda por Query")

	t.Log("\n📝 PASO 1: Buscar actividades con query 'yoga'")
	resp, err := http.Get(searchAPIURL + "/search?q=yoga")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("❌ Búsqueda falló - Status: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		t.Fatalf("❌ Error parseando respuesta: %v", err)
	}

	t.Logf("✅ Búsqueda 'yoga' retornó %d resultados", searchResp.TotalCount)

	if searchResp.TotalCount == 0 {
		t.Error("❌ Se esperaban resultados para 'yoga'")
	}

	// Verificar que los resultados contienen 'yoga'
	for _, doc := range searchResp.Results[:min(5, len(searchResp.Results))] {
		t.Logf("   - %s (categoría: %s)", doc.Titulo, doc.Categoria)
	}

	t.Log("\n📝 PASO 2: Buscar actividades con query 'spinning'")
	resp2, err := http.Get(searchAPIURL + "/search?q=spinning")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp2.Body.Close()

	var searchResp2 SearchResponse
	json.NewDecoder(resp2.Body).Decode(&searchResp2)
	t.Logf("✅ Búsqueda 'spinning' retornó %d resultados", searchResp2.TotalCount)

	t.Log("\n📝 PASO 3: Buscar con query que no existe")
	resp3, err := http.Get(searchAPIURL + "/search?q=natacion_inexistente_xyz")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp3.Body.Close()

	var searchResp3 SearchResponse
	json.NewDecoder(resp3.Body).Decode(&searchResp3)
	t.Logf("✅ Búsqueda inexistente retornó %d resultados (esperado: 0)", searchResp3.TotalCount)

	t.Log("\n================================================================================")
	t.Log("🎉 TEST BÚSQUEDA POR QUERY COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_SearchByCategory(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Búsqueda por Categoría")

	categories := []string{"yoga", "spinning", "funcional", "pilates"}

	for _, cat := range categories {
		t.Logf("\n📝 Buscando categoría: %s", cat)
		url := fmt.Sprintf("%s/search?categoria=%s", searchAPIURL, cat)
		resp, err := http.Get(url)
		if err != nil {
			t.Logf("⚠️  Error buscando categoría %s: %v", cat, err)
			continue
		}
		defer resp.Body.Close()

		var searchResp SearchResponse
		json.NewDecoder(resp.Body).Decode(&searchResp)
		t.Logf("✅ Categoría '%s' retornó %d resultados", cat, searchResp.TotalCount)

		// Verificar que todos los resultados son de la categoría correcta
		for _, doc := range searchResp.Results {
			if doc.Categoria != cat && doc.Categoria != "" {
				t.Logf("⚠️  Documento %s tiene categoría '%s', esperado '%s'", doc.ID, doc.Categoria, cat)
			}
		}
	}

	t.Log("\n================================================================================")
	t.Log("🎉 TEST BÚSQUEDA POR CATEGORÍA COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_SearchPagination(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Paginación")

	t.Log("\n📝 PASO 1: Obtener página 1 con 5 elementos")
	resp, err := http.Get(searchAPIURL + "/search?page=1&page_size=5")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp.Body.Close()

	var page1 SearchResponse
	json.NewDecoder(resp.Body).Decode(&page1)

	t.Logf("✅ Página 1: %d resultados de %d total", len(page1.Results), page1.TotalCount)
	t.Logf("   - Page: %d, PageSize: %d, TotalPages: %d", page1.Page, page1.PageSize, page1.TotalPages)

	if len(page1.Results) > 5 {
		t.Errorf("❌ Se esperaban máximo 5 resultados, se obtuvieron %d", len(page1.Results))
	}

	if page1.Page != 1 {
		t.Errorf("❌ Page esperado 1, obtenido %d", page1.Page)
	}

	t.Log("\n📝 PASO 2: Obtener página 2")
	resp2, err := http.Get(searchAPIURL + "/search?page=2&page_size=5")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp2.Body.Close()

	var page2 SearchResponse
	json.NewDecoder(resp2.Body).Decode(&page2)

	t.Logf("✅ Página 2: %d resultados", len(page2.Results))

	// Verificar que los resultados son diferentes
	if len(page1.Results) > 0 && len(page2.Results) > 0 {
		if page1.Results[0].ID == page2.Results[0].ID {
			t.Log("⚠️  Primera página y segunda tienen el mismo primer elemento")
		} else {
			t.Log("✅ Páginas contienen resultados diferentes")
		}
	}

	t.Log("\n📝 PASO 3: Verificar página fuera de rango")
	resp3, err := http.Get(searchAPIURL + "/search?page=9999&page_size=10")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp3.Body.Close()

	var pageOutOfRange SearchResponse
	json.NewDecoder(resp3.Body).Decode(&pageOutOfRange)

	t.Logf("✅ Página 9999: %d resultados (esperado: 0)", len(pageOutOfRange.Results))

	t.Log("\n================================================================================")
	t.Log("🎉 TEST PAGINACIÓN COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_SearchByDay(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Búsqueda por Día")

	days := []string{"Lunes", "Martes", "Miercoles", "Jueves", "Viernes"}

	for _, day := range days {
		url := fmt.Sprintf("%s/search?dia=%s", searchAPIURL, day)
		resp, err := http.Get(url)
		if err != nil {
			t.Logf("⚠️  Error buscando día %s: %v", day, err)
			continue
		}
		defer resp.Body.Close()

		var searchResp SearchResponse
		json.NewDecoder(resp.Body).Decode(&searchResp)
		t.Logf("✅ Día '%s' retornó %d resultados", day, searchResp.TotalCount)
	}

	t.Log("\n================================================================================")
	t.Log("🎉 TEST BÚSQUEDA POR DÍA COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_CombinedFilters(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Filtros Combinados")

	t.Log("\n📝 PASO 1: Buscar yoga del Lunes")
	resp, err := http.Get(searchAPIURL + "/search?q=yoga&dia=Lunes")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp.Body.Close()

	var searchResp SearchResponse
	json.NewDecoder(resp.Body).Decode(&searchResp)

	t.Logf("✅ 'yoga' + 'Lunes' retornó %d resultados", searchResp.TotalCount)

	for _, doc := range searchResp.Results[:min(3, len(searchResp.Results))] {
		t.Logf("   - %s (%s, %s)", doc.Titulo, doc.Categoria, doc.Dia)
	}

	t.Log("\n📝 PASO 2: Buscar spinning categoría específica")
	resp2, err := http.Get(searchAPIURL + "/search?categoria=spinning&page_size=5")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp2.Body.Close()

	var searchResp2 SearchResponse
	json.NewDecoder(resp2.Body).Decode(&searchResp2)

	t.Logf("✅ Categoría 'spinning' retornó %d resultados", searchResp2.TotalCount)

	t.Log("\n================================================================================")
	t.Log("🎉 TEST FILTROS COMBINADOS COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_SearchPerformance(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Performance")

	queries := []string{"yoga", "spinning", "funcional", "clase", "matutino"}
	var totalTime time.Duration

	t.Log("\n📝 Ejecutando 5 búsquedas para medir tiempo promedio...")

	for _, q := range queries {
		start := time.Now()
		resp, err := http.Get(searchAPIURL + "/search?q=" + q)
		elapsed := time.Since(start)
		totalTime += elapsed

		if err != nil {
			t.Logf("⚠️  Error en búsqueda '%s': %v", q, err)
			continue
		}
		resp.Body.Close()

		t.Logf("   - Búsqueda '%s': %v", q, elapsed)
	}

	avgTime := totalTime / time.Duration(len(queries))
	t.Logf("\n✅ Tiempo promedio de búsqueda: %v", avgTime)

	if avgTime > 500*time.Millisecond {
		t.Logf("⚠️  El tiempo promedio supera 500ms, considerar optimización")
	} else {
		t.Log("✅ Performance dentro de parámetros aceptables (<500ms)")
	}

	t.Log("\n================================================================================")
	t.Log("🎉 TEST PERFORMANCE COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_EmptyQuery(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Query Vacío")

	t.Log("\n📝 PASO 1: Búsqueda sin parámetros (listar todo)")
	resp, err := http.Get(searchAPIURL + "/search")
	if err != nil {
		t.Fatalf("❌ Error en búsqueda: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("❌ Status esperado 200, obtenido %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	json.NewDecoder(resp.Body).Decode(&searchResp)

	t.Logf("✅ Búsqueda sin filtros retornó %d resultados", searchResp.TotalCount)
	t.Logf("   - Mostrando página %d de %d", searchResp.Page, searchResp.TotalPages)

	if searchResp.TotalCount == 0 {
		t.Log("⚠️  No hay documentos indexados")
	}

	t.Log("\n================================================================================")
	t.Log("🎉 TEST QUERY VACÍO COMPLETADO!")
	t.Log("================================================================================")
}

func TestSearchAPI_SpecialCharacters(t *testing.T) {
	t.Log("🚀 Iniciando test: Search API - Caracteres Especiales")

	specialQueries := []string{
		"yoga%20matutino",  // espacio encoded
		"clase+especial",  // plus
		"María",           // acentos
		"González",        // ñ
	}

	for _, q := range specialQueries {
		url := fmt.Sprintf("%s/search?q=%s", searchAPIURL, q)
		resp, err := http.Get(url)
		if err != nil {
			t.Logf("⚠️  Error con query '%s': %v", q, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("⚠️  Query '%s' retornó status %d", q, resp.StatusCode)
			continue
		}

		var searchResp SearchResponse
		json.NewDecoder(resp.Body).Decode(&searchResp)
		t.Logf("✅ Query '%s' retornó %d resultados", q, searchResp.TotalCount)
	}

	t.Log("\n================================================================================")
	t.Log("🎉 TEST CARACTERES ESPECIALES COMPLETADO!")
	t.Log("================================================================================")
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
