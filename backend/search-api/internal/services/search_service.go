package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/yourusername/gym-management/search-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/search-api/internal/integrations"
	"github.com/yourusername/gym-management/search-api/internal/repositories"
)

// SearchService - Servicio de búsqueda con Solr + MySQL fallback
type SearchService struct {
	solrClient   *integrations.SolrClient
	mysqlRepo    *repositories.MySQLSearchRepository
	useSolr      bool
	documents    map[string]dtos.SearchDocument // Cache en memoria
	mu           sync.RWMutex
}

// NewSearchService - Constructor con Solr + MySQL fallback
func NewSearchService(solrClient *integrations.SolrClient, mysqlRepo *repositories.MySQLSearchRepository) *SearchService {
	service := &SearchService{
		solrClient: solrClient,
		mysqlRepo:  mysqlRepo,
		documents:  make(map[string]dtos.SearchDocument),
		useSolr:    true,
	}

	// Verificar disponibilidad de Solr
	if solrClient != nil {
		if err := solrClient.Ping(); err != nil {
			log.Printf("⚠️  Solr not available, using MySQL fallback: %v", err)
			service.useSolr = false
		} else {
			log.Println("✅ Solr connected successfully")
		}
	} else {
		log.Println("⚠️  Solr client not configured, using MySQL fallback")
		service.useSolr = false
	}

	return service
}

// IndexDocument indexa un documento
func (s *SearchService) IndexDocument(doc dtos.SearchDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Indexar en cache local
	s.documents[doc.ID] = doc

	// Intentar indexar en Solr si está disponible
	if s.useSolr && s.solrClient != nil {
		if err := s.solrClient.IndexDocument(doc); err != nil {
			log.Printf("⚠️  Error indexing in Solr: %v", err)
			// No retornamos error, seguimos con cache local
		}
	}

	return nil
}

// IndexDocuments indexa múltiples documentos en batch
func (s *SearchService) IndexDocuments(docs []dtos.SearchDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Indexar en cache local
	for _, doc := range docs {
		s.documents[doc.ID] = doc
	}

	// Intentar indexar en Solr si está disponible
	if s.useSolr && s.solrClient != nil {
		if err := s.solrClient.IndexDocuments(docs); err != nil {
			log.Printf("⚠️  Error batch indexing in Solr: %v", err)
			// No retornamos error, seguimos con cache local
		} else {
			log.Printf("✅ Indexed %d documents in Solr", len(docs))
		}
	}

	return nil
}

// DeleteDocument elimina un documento del índice
func (s *SearchService) DeleteDocument(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Eliminar de cache local
	delete(s.documents, id)

	// Intentar eliminar de Solr si está disponible
	if s.useSolr && s.solrClient != nil {
		if err := s.solrClient.DeleteDocument(id); err != nil {
			log.Printf("⚠️  Error deleting from Solr: %v", err)
		}
	}

	return nil
}

// Search realiza una búsqueda en los documentos
func (s *SearchService) Search(req dtos.SearchRequest) (*dtos.SearchResponse, error) {
	// Valores por defecto
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	var results []dtos.SearchDocument
	var totalCount int
	var err error

	// Intentar búsqueda en Solr primero
	if s.useSolr && s.solrClient != nil {
		results, totalCount, err = s.solrClient.Search(req)
		if err != nil {
			log.Printf("⚠️  Solr search failed, falling back to MySQL: %v", err)
			s.useSolr = false // Desactivar Solr temporalmente
		} else {
			log.Printf("✅ Solr search successful: %d results", totalCount)
		}
	}

	// Fallback a MySQL si Solr falla o no está disponible
	if !s.useSolr && s.mysqlRepo != nil {
		log.Println("🔍 Using MySQL fulltext search")
		results, totalCount, err = s.mysqlRepo.SearchActivities(req)
		if err != nil {
			return nil, fmt.Errorf("MySQL search failed: %w", err)
		}
	}

	// Si ambos fallan, usar cache en memoria (último recurso)
	if results == nil {
		log.Println("⚠️  Using in-memory cache as last resort")
		results, totalCount = s.searchInMemory(req)
	}

	// Calcular paginación
	totalPages := (totalCount + req.PageSize - 1) / req.PageSize

	return &dtos.SearchResponse{
		Results:    results,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// searchInMemory - Búsqueda en memoria (último recurso)
func (s *SearchService) searchInMemory(req dtos.SearchRequest) ([]dtos.SearchDocument, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []dtos.SearchDocument

	// Filtrar documentos
	for _, doc := range s.documents {
		if s.matchesSearch(doc, req) {
			results = append(results, doc)
		}
	}

	// Calcular paginación
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

	return results, totalCount
}

// matchesSearch verifica si un documento coincide con los criterios de búsqueda
func (s *SearchService) matchesSearch(doc dtos.SearchDocument, req dtos.SearchRequest) bool {
	// Filtrar por tipo si se especifica
	if req.Type != "" && doc.Type != req.Type {
		return false
	}

	// Búsqueda por query (búsqueda de texto)
	if req.Query != "" {
		query := strings.ToLower(req.Query)
		text := strings.ToLower(fmt.Sprintf("%s %s %s %s %s",
			doc.Titulo, doc.Descripcion, doc.Categoria,
			doc.Instructor, doc.PlanNombre))

		if !strings.Contains(text, query) {
			return false
		}
	}

	// Aplicar filtros adicionales
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
		case "sucursal_id":
			if doc.SucursalID != value {
				return false
			}
		case "requiere_premium":
			reqPremium := value == "true"
			if doc.RequierePremium != reqPremium {
				return false
			}
		case "estado":
			if doc.Estado != value {
				return false
			}
		}
	}

	return true
}

// GetDocumentByID obtiene un documento por ID
func (s *SearchService) GetDocumentByID(id string) (*dtos.SearchDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.documents[id]
	if !exists {
		return nil, fmt.Errorf("documento no encontrado")
	}

	return &doc, nil
}

// IndexFromEvent indexa un documento desde un evento de RabbitMQ
func (s *SearchService) IndexFromEvent(event dtos.RabbitMQEvent) error {
	// Convertir data a SearchDocument
	docBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	var doc dtos.SearchDocument
	if err := json.Unmarshal(docBytes, &doc); err != nil {
		return err
	}

	doc.ID = event.ID  // Usar solo el ID numérico (sin prefijo) para consistencia con MySQL
	doc.Type = event.Type

	return s.IndexDocument(doc)
}

// ReindexActivityByID obtiene una actividad desde MySQL y la reindexa en Solr
func (s *SearchService) ReindexActivityByID(activityID string) error {
	if s.mysqlRepo == nil {
		return fmt.Errorf("mysql repository not available")
	}

	// Obtener actividad actualizada desde MySQL
	activity, err := s.mysqlRepo.GetActivityByID(activityID)
	if err != nil {
		return fmt.Errorf("error obteniendo actividad desde MySQL: %w", err)
	}

	// Indexar en Solr
	return s.IndexDocument(activity)
}

// UpdateActivityCupo actualiza solo el cupo_disponible de una actividad (partial update optimizado)
func (s *SearchService) UpdateActivityCupo(activityID string) error {
	if s.mysqlRepo == nil {
		return fmt.Errorf("mysql repository not available")
	}

	// Solo obtener cupo_disponible desde MySQL (query optimizada)
	cupoDisponible, err := s.mysqlRepo.GetCupoDisponible(activityID)
	if err != nil {
		return fmt.Errorf("error obteniendo cupo_disponible: %w", err)
	}

	// Actualizar cache local
	s.mu.Lock()
	if doc, exists := s.documents[activityID]; exists {
		doc.CupoDisponible = cupoDisponible
		s.documents[activityID] = doc
	}
	s.mu.Unlock()

	// Partial update en Solr (solo actualiza cupo_disponible)
	if s.useSolr && s.solrClient != nil {
		if err := s.solrClient.PartialUpdateCupo(activityID, cupoDisponible); err != nil {
			log.Printf("⚠️  Error en partial update de Solr: %v", err)
			return err
		}
	}

	return nil
}

// GetStats retorna estadísticas del índice
func (s *SearchService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_documents": len(s.documents),
		"types":           make(map[string]int),
	}

	typeCounts := make(map[string]int)
	for _, doc := range s.documents {
		typeCounts[doc.Type]++
	}
	stats["types"] = typeCounts

	return stats
}

// GetUniqueCategories retorna las categorías únicas de las actividades
func (s *SearchService) GetUniqueCategories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categorySet := make(map[string]bool)
	for _, doc := range s.documents {
		if doc.Categoria != "" {
			categorySet[doc.Categoria] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	// Ordenar alfabéticamente
	for i := 0; i < len(categories)-1; i++ {
		for j := i + 1; j < len(categories); j++ {
			if categories[i] > categories[j] {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	return categories
}
