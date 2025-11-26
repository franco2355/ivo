package services

import (
	"testing"

	"github.com/yourusername/gym-management/search-api/internal/domain/dtos"
)

// MockSearchService - Servicio de búsqueda para tests (sin dependencias externas)
type MockSearchService struct {
	documents map[string]dtos.SearchDocument
}

func NewMockSearchService() *MockSearchService {
	return &MockSearchService{
		documents: make(map[string]dtos.SearchDocument),
	}
}

func (m *MockSearchService) IndexDocument(doc dtos.SearchDocument) {
	m.documents[doc.ID] = doc
}

func (m *MockSearchService) DeleteDocument(id string) {
	delete(m.documents, id)
}

func (m *MockSearchService) GetDocument(id string) (dtos.SearchDocument, bool) {
	doc, exists := m.documents[id]
	return doc, exists
}

func (m *MockSearchService) Search(req dtos.SearchRequest) *dtos.SearchResponse {
	var results []dtos.SearchDocument

	for _, doc := range m.documents {
		if matchesSearchCriteria(doc, req) {
			results = append(results, doc)
		}
	}

	// Paginación
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	totalCount := len(results)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start >= totalCount {
		results = []dtos.SearchDocument{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		results = results[start:end]
	}

	totalPages := (totalCount + req.PageSize - 1) / req.PageSize

	return &dtos.SearchResponse{
		Results:    results,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}
}

func matchesSearchCriteria(doc dtos.SearchDocument, req dtos.SearchRequest) bool {
	// Filtrar por tipo
	if req.Type != "" && doc.Type != req.Type {
		return false
	}

	// Búsqueda por query en título
	if req.Query != "" {
		found := false
		if containsSubstr(doc.Titulo, req.Query) {
			found = true
		}
		if containsSubstr(doc.Descripcion, req.Query) {
			found = true
		}
		if containsSubstr(doc.Categoria, req.Query) {
			found = true
		}
		if !found {
			return false
		}
	}

	// Filtros
	for key, value := range req.Filters {
		switch key {
		case "categoria":
			if doc.Categoria != value {
				return false
			}
		case "dia":
			if doc.Dia != value {
				return false
			}
		case "instructor":
			if doc.Instructor != value {
				return false
			}
		}
	}

	return true
}

func containsSubstr(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ========== TESTS ==========

func TestIndexDocument_Success(t *testing.T) {
	service := NewMockSearchService()

	doc := dtos.SearchDocument{
		ID:          "activity_1",
		Type:        "activity",
		Titulo:      "Yoga Matutino",
		Descripcion: "Clase de yoga para principiantes",
		Categoria:   "yoga",
		Instructor:  "María García",
		Dia:         "Lunes",
	}

	service.IndexDocument(doc)

	// Verificar que el documento fue indexado
	retrieved, exists := service.GetDocument("activity_1")
	if !exists {
		t.Error("El documento debería existir después de indexar")
	}

	if retrieved.Titulo != "Yoga Matutino" {
		t.Errorf("Título esperado 'Yoga Matutino', obtenido '%s'", retrieved.Titulo)
	}

	if retrieved.Categoria != "yoga" {
		t.Errorf("Categoría esperada 'yoga', obtenida '%s'", retrieved.Categoria)
	}
}

func TestIndexDocument_MultipleDocuments(t *testing.T) {
	service := NewMockSearchService()

	docs := []dtos.SearchDocument{
		{ID: "1", Type: "activity", Titulo: "Yoga", Categoria: "yoga"},
		{ID: "2", Type: "activity", Titulo: "Spinning", Categoria: "spinning"},
		{ID: "3", Type: "activity", Titulo: "Pilates", Categoria: "pilates"},
	}

	for _, doc := range docs {
		service.IndexDocument(doc)
	}

	// Verificar que todos los documentos fueron indexados
	for _, doc := range docs {
		_, exists := service.GetDocument(doc.ID)
		if !exists {
			t.Errorf("Documento %s debería existir", doc.ID)
		}
	}
}

func TestDeleteDocument_Success(t *testing.T) {
	service := NewMockSearchService()

	doc := dtos.SearchDocument{
		ID:     "activity_1",
		Type:   "activity",
		Titulo: "Yoga Matutino",
	}

	service.IndexDocument(doc)

	// Verificar que existe
	_, exists := service.GetDocument("activity_1")
	if !exists {
		t.Error("El documento debería existir antes de eliminar")
	}

	// Eliminar
	service.DeleteDocument("activity_1")

	// Verificar que ya no existe
	_, exists = service.GetDocument("activity_1")
	if exists {
		t.Error("El documento no debería existir después de eliminar")
	}
}

func TestDeleteDocument_NonExistent(t *testing.T) {
	service := NewMockSearchService()

	// No debería causar error
	service.DeleteDocument("non_existent_id")

	// Verificar que el servicio sigue funcionando
	_, exists := service.GetDocument("non_existent_id")
	if exists {
		t.Error("Documento inexistente no debería encontrarse")
	}
}

func TestSearch_ByQuery(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga Matutino", Categoria: "yoga"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "activity", Titulo: "Yoga Vespertino", Categoria: "yoga"})
	service.IndexDocument(dtos.SearchDocument{ID: "3", Type: "activity", Titulo: "Spinning Intenso", Categoria: "spinning"})

	// Buscar "Yoga"
	response := service.Search(dtos.SearchRequest{Query: "Yoga"})

	if response.TotalCount != 2 {
		t.Errorf("Se esperaban 2 resultados para 'Yoga', se obtuvieron %d", response.TotalCount)
	}

	for _, doc := range response.Results {
		if !containsSubstr(doc.Titulo, "Yoga") {
			t.Errorf("Resultado '%s' no contiene 'Yoga'", doc.Titulo)
		}
	}
}

func TestSearch_ByType(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "plan", PlanNombre: "Plan Premium"})
	service.IndexDocument(dtos.SearchDocument{ID: "3", Type: "activity", Titulo: "Spinning"})

	// Buscar solo actividades
	response := service.Search(dtos.SearchRequest{Type: "activity"})

	if response.TotalCount != 2 {
		t.Errorf("Se esperaban 2 actividades, se obtuvieron %d", response.TotalCount)
	}

	for _, doc := range response.Results {
		if doc.Type != "activity" {
			t.Errorf("Tipo esperado 'activity', obtenido '%s'", doc.Type)
		}
	}
}

func TestSearch_ByCategory(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga 1", Categoria: "yoga"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "activity", Titulo: "Yoga 2", Categoria: "yoga"})
	service.IndexDocument(dtos.SearchDocument{ID: "3", Type: "activity", Titulo: "Spinning", Categoria: "spinning"})

	// Buscar por categoría yoga
	response := service.Search(dtos.SearchRequest{
		Filters: map[string]string{"categoria": "yoga"},
	})

	if response.TotalCount != 2 {
		t.Errorf("Se esperaban 2 resultados de yoga, se obtuvieron %d", response.TotalCount)
	}

	for _, doc := range response.Results {
		if doc.Categoria != "yoga" {
			t.Errorf("Categoría esperada 'yoga', obtenida '%s'", doc.Categoria)
		}
	}
}

func TestSearch_ByDay(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga Lunes", Dia: "Lunes"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "activity", Titulo: "Yoga Martes", Dia: "Martes"})
	service.IndexDocument(dtos.SearchDocument{ID: "3", Type: "activity", Titulo: "Spinning Lunes", Dia: "Lunes"})

	// Buscar actividades del Lunes
	response := service.Search(dtos.SearchRequest{
		Filters: map[string]string{"dia": "Lunes"},
	})

	if response.TotalCount != 2 {
		t.Errorf("Se esperaban 2 actividades del Lunes, se obtuvieron %d", response.TotalCount)
	}

	for _, doc := range response.Results {
		if doc.Dia != "Lunes" {
			t.Errorf("Día esperado 'Lunes', obtenido '%s'", doc.Dia)
		}
	}
}

func TestSearch_ByInstructor(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga", Instructor: "María"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "activity", Titulo: "Spinning", Instructor: "Juan"})
	service.IndexDocument(dtos.SearchDocument{ID: "3", Type: "activity", Titulo: "Pilates", Instructor: "María"})

	// Buscar actividades de María
	response := service.Search(dtos.SearchRequest{
		Filters: map[string]string{"instructor": "María"},
	})

	if response.TotalCount != 2 {
		t.Errorf("Se esperaban 2 actividades de María, se obtuvieron %d", response.TotalCount)
	}

	for _, doc := range response.Results {
		if doc.Instructor != "María" {
			t.Errorf("Instructor esperado 'María', obtenido '%s'", doc.Instructor)
		}
	}
}

func TestSearch_CombinedFilters(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga Lunes", Categoria: "yoga", Dia: "Lunes"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "activity", Titulo: "Yoga Martes", Categoria: "yoga", Dia: "Martes"})
	service.IndexDocument(dtos.SearchDocument{ID: "3", Type: "activity", Titulo: "Spinning Lunes", Categoria: "spinning", Dia: "Lunes"})

	// Buscar yoga del Lunes
	response := service.Search(dtos.SearchRequest{
		Filters: map[string]string{
			"categoria": "yoga",
			"dia":       "Lunes",
		},
	})

	if response.TotalCount != 1 {
		t.Errorf("Se esperaba 1 resultado (yoga del Lunes), se obtuvieron %d", response.TotalCount)
	}

	if len(response.Results) > 0 {
		doc := response.Results[0]
		if doc.Categoria != "yoga" || doc.Dia != "Lunes" {
			t.Error("El resultado no coincide con los filtros combinados")
		}
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga", Categoria: "yoga"})

	// Buscar algo que no existe
	response := service.Search(dtos.SearchRequest{Query: "Natación"})

	if response.TotalCount != 0 {
		t.Errorf("Se esperaban 0 resultados, se obtuvieron %d", response.TotalCount)
	}

	if len(response.Results) != 0 {
		t.Error("La lista de resultados debería estar vacía")
	}
}

func TestSearch_Pagination(t *testing.T) {
	service := NewMockSearchService()

	// Indexar 25 documentos
	for i := 1; i <= 25; i++ {
		service.IndexDocument(dtos.SearchDocument{
			ID:        string(rune('0' + i)),
			Type:      "activity",
			Titulo:    "Actividad",
			Categoria: "yoga",
		})
	}

	// Página 1, 10 por página
	response := service.Search(dtos.SearchRequest{
		Type:     "activity",
		Page:     1,
		PageSize: 10,
	})

	if response.TotalCount != 25 {
		t.Errorf("Total esperado 25, obtenido %d", response.TotalCount)
	}

	if len(response.Results) != 10 {
		t.Errorf("Se esperaban 10 resultados en la página, se obtuvieron %d", len(response.Results))
	}

	if response.TotalPages != 3 {
		t.Errorf("Se esperaban 3 páginas totales, se obtuvieron %d", response.TotalPages)
	}

	// Página 3 (última)
	response = service.Search(dtos.SearchRequest{
		Type:     "activity",
		Page:     3,
		PageSize: 10,
	})

	if len(response.Results) != 5 {
		t.Errorf("Se esperaban 5 resultados en la última página, se obtuvieron %d", len(response.Results))
	}
}

func TestSearch_PaginationDefaults(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga"})

	// Sin especificar page y pageSize
	response := service.Search(dtos.SearchRequest{})

	if response.Page != 1 {
		t.Errorf("Página por defecto esperada 1, obtenida %d", response.Page)
	}

	if response.PageSize != 10 {
		t.Errorf("PageSize por defecto esperado 10, obtenido %d", response.PageSize)
	}
}

func TestSearch_PageOutOfRange(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{ID: "1", Type: "activity", Titulo: "Yoga"})
	service.IndexDocument(dtos.SearchDocument{ID: "2", Type: "activity", Titulo: "Spinning"})

	// Página que no existe
	response := service.Search(dtos.SearchRequest{
		Page:     100,
		PageSize: 10,
	})

	if len(response.Results) != 0 {
		t.Error("Página fuera de rango debería retornar lista vacía")
	}

	// TotalCount sigue siendo correcto
	if response.TotalCount != 2 {
		t.Errorf("TotalCount esperado 2, obtenido %d", response.TotalCount)
	}
}

func TestGetDocument_Exists(t *testing.T) {
	service := NewMockSearchService()

	doc := dtos.SearchDocument{
		ID:          "activity_1",
		Type:        "activity",
		Titulo:      "Yoga Matutino",
		Descripcion: "Clase de yoga",
	}

	service.IndexDocument(doc)

	retrieved, exists := service.GetDocument("activity_1")

	if !exists {
		t.Error("El documento debería existir")
	}

	if retrieved.Titulo != doc.Titulo {
		t.Errorf("Título esperado '%s', obtenido '%s'", doc.Titulo, retrieved.Titulo)
	}

	if retrieved.Descripcion != doc.Descripcion {
		t.Errorf("Descripción esperada '%s', obtenida '%s'", doc.Descripcion, retrieved.Descripcion)
	}
}

func TestGetDocument_NotExists(t *testing.T) {
	service := NewMockSearchService()

	_, exists := service.GetDocument("non_existent")

	if exists {
		t.Error("Documento inexistente no debería encontrarse")
	}
}

func TestUpdateDocument(t *testing.T) {
	service := NewMockSearchService()

	// Indexar documento inicial
	doc := dtos.SearchDocument{
		ID:     "1",
		Type:   "activity",
		Titulo: "Yoga Original",
	}
	service.IndexDocument(doc)

	// Actualizar documento
	updatedDoc := dtos.SearchDocument{
		ID:     "1",
		Type:   "activity",
		Titulo: "Yoga Actualizado",
	}
	service.IndexDocument(updatedDoc)

	// Verificar actualización
	retrieved, _ := service.GetDocument("1")
	if retrieved.Titulo != "Yoga Actualizado" {
		t.Errorf("Título esperado 'Yoga Actualizado', obtenido '%s'", retrieved.Titulo)
	}
}

func TestSearch_ByDescription(t *testing.T) {
	service := NewMockSearchService()

	service.IndexDocument(dtos.SearchDocument{
		ID:          "1",
		Type:        "activity",
		Titulo:      "Clase Matutina",
		Descripcion: "Yoga para principiantes",
	})
	service.IndexDocument(dtos.SearchDocument{
		ID:          "2",
		Type:        "activity",
		Titulo:      "Clase Vespertina",
		Descripcion: "Spinning avanzado",
	})

	// Buscar por descripción
	response := service.Search(dtos.SearchRequest{Query: "principiantes"})

	if response.TotalCount != 1 {
		t.Errorf("Se esperaba 1 resultado, se obtuvieron %d", response.TotalCount)
	}

	if len(response.Results) > 0 && response.Results[0].ID != "1" {
		t.Error("Resultado incorrecto")
	}
}
