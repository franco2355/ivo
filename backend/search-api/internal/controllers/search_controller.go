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

	req := dtos.SearchRequest{
		Query:    query,
		Type:     typeFilter,
		Page:     1,
		PageSize: 20,
	}

	response, err := c.searchService.Search(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
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

// HealthCheck verifica el estado del servicio
func (c *SearchController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "search-api",
	})
}
