package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gym-management/search-api/internal/domain/dtos"
	"github.com/yourusername/gym-management/search-api/internal/services"
)

// SearchController - Controlador HTTP para búsquedas con DI
type SearchController struct {
	searchService *services.SearchService
	cacheService  *services.CacheService
}

// NewSearchController - Constructor con Dependency Injection
func NewSearchController(searchService *services.SearchService, cacheService *services.CacheService) *SearchController {
	return &SearchController{
		searchService: searchService,
		cacheService:  cacheService,
	}
}

// Search realiza una búsqueda avanzada con caché
func (c *SearchController) Search(ctx *gin.Context) {
	var req dtos.SearchRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generar clave de caché
	cacheKey := services.GenerateCacheKey("search", req)

	// Intentar obtener del caché
	if cachedData, found := c.cacheService.Get(cacheKey); found {
		var response dtos.SearchResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			ctx.Header("X-Cache", "HIT")
			ctx.JSON(http.StatusOK, response)
			return
		}
	}

	// Si no está en caché, realizar búsqueda
	response, err := c.searchService.Search(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Guardar en caché
	if responseData, err := json.Marshal(response); err == nil {
		c.cacheService.Set(cacheKey, responseData)
	}

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, response)
}

// QuickSearch realiza una búsqueda rápida desde query params
func (c *SearchController) QuickSearch(ctx *gin.Context) {
	query := ctx.Query("q")
	typeFilter := ctx.Query("type")
	categoria := ctx.Query("categoria")
	dia := ctx.Query("dia")
	instructor := ctx.Query("instructor")

	// Parsear page y page_size
	page := 1
	pageSize := 20
	if p := ctx.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	// Construir filtros
	filters := make(map[string]string)
	if categoria != "" {
		filters["categoria"] = categoria
	}
	if dia != "" {
		filters["dia"] = dia
	}
	if instructor != "" {
		filters["instructor"] = instructor
	}

	req := dtos.SearchRequest{
		Query:    query,
		Type:     typeFilter,
		Filters:  filters,
		Page:     page,
		PageSize: pageSize,
	}

	// Generar clave de caché
	cacheKey := services.GenerateCacheKey("quicksearch", req)

	// Intentar obtener del caché
	if cachedData, found := c.cacheService.Get(cacheKey); found {
		var response dtos.SearchResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			ctx.Header("X-Cache", "HIT")
			ctx.JSON(http.StatusOK, response)
			return
		}
	}

	response, err := c.searchService.Search(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Guardar en caché
	if responseData, err := json.Marshal(response); err == nil {
		c.cacheService.Set(cacheKey, responseData)
	}

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, response)
}

// parseInt helper function
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

// GetDocument obtiene un documento por ID
func (c *SearchController) GetDocument(ctx *gin.Context) {
	docID := ctx.Param("id")

	doc, err := c.searchService.GetDocumentByID(docID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

// IndexDocument indexa un documento manualmente
func (c *SearchController) IndexDocument(ctx *gin.Context) {
	var doc dtos.SearchDocument

	if err := ctx.ShouldBindJSON(&doc); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.searchService.IndexDocument(doc)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidar caché
	c.cacheService.InvalidatePattern(doc.Type)

	ctx.JSON(http.StatusCreated, gin.H{"message": "Documento indexado correctamente"})
}

// DeleteDocument elimina un documento del índice
func (c *SearchController) DeleteDocument(ctx *gin.Context) {
	docID := ctx.Param("id")

	err := c.searchService.DeleteDocument(docID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Documento eliminado correctamente"})
}

// GetStats obtiene estadísticas del índice
func (c *SearchController) GetStats(ctx *gin.Context) {
	stats := c.searchService.GetStats()
	ctx.JSON(http.StatusOK, stats)
}

// GetCategories devuelve las categorías únicas de las actividades
func (c *SearchController) GetCategories(ctx *gin.Context) {
	categories := c.searchService.GetUniqueCategories()
	ctx.JSON(http.StatusOK, gin.H{"categories": categories})
}

// HealthCheck verifica el estado del servicio
func (c *SearchController) HealthCheck(ctx *gin.Context) {
	health := gin.H{
		"status":  "ok",
		"service": "search-api",
		"checks":  gin.H{},
	}

	// Check Solr via SearchService
	if c.searchService != nil {
		stats := c.searchService.GetStats()
		if stats != nil {
			health["checks"].(gin.H)["solr"] = "connected"
		} else {
			health["checks"].(gin.H)["solr"] = "unavailable"
		}
	}

	// Check Memcached via CacheService
	if c.cacheService != nil {
		// Try to set and get a test value
		testKey := "health_check_test"
		testValue := []byte("ok")
		c.cacheService.Set(testKey, testValue)
		if _, found := c.cacheService.Get(testKey); found {
			health["checks"].(gin.H)["cache"] = "connected"
		} else {
			health["checks"].(gin.H)["cache"] = "degraded"
		}
	}

	ctx.JSON(http.StatusOK, health)
}
