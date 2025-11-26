package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gym-management/search-api/internal/clients"
	"github.com/yourusername/gym-management/search-api/internal/config"
	"github.com/yourusername/gym-management/search-api/internal/controllers"
	"github.com/yourusername/gym-management/search-api/internal/integrations"
	"github.com/yourusername/gym-management/search-api/internal/middleware"
	"github.com/yourusername/gym-management/search-api/internal/repositories"
	"github.com/yourusername/gym-management/search-api/internal/services"
)

func main() {
	// 1. Cargar configuración
	cfg := config.LoadConfig()

	// 2. Inicializar clientes externos
	log.Println("🔧 Initializing external clients...")

	// Crear cliente Solr
	var solrClient *integrations.SolrClient
	if cfg.SolrURL != "" {
		solrClient = integrations.NewSolrClient(cfg.SolrURL)
		log.Printf("🔍 Solr client configured: %s", cfg.SolrURL)
	}

	// Crear repositorio MySQL
	mysqlRepo, err := repositories.NewMySQLSearchRepository(
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBSchema,
	)
	if err != nil {
		log.Printf("⚠️  Warning: MySQL repository failed: %v", err)
	} else {
		defer mysqlRepo.Close()
		log.Println("✅ MySQL repository connected")
	}

	// 3. Crear servicios con DI
	searchService := services.NewSearchService(solrClient, mysqlRepo)
	cacheService := services.NewCacheService(
		cfg.MemcachedServers,
		cfg.CacheTTL,
		cfg.LocalCacheTTL,
	)

	// Iniciar limpieza periódica del caché local
	cacheService.StartCleanupRoutine(5 * time.Minute)

	// 4. Indexar actividades existentes desde MySQL
	log.Println("📊 Indexing existing activities...")
	if mysqlRepo != nil {
		activities, err := mysqlRepo.GetAllActivities()
		if err != nil {
			log.Printf("⚠️  Warning: Could not fetch activities for initial indexing: %v", err)
		} else {
			if err := searchService.IndexDocuments(activities); err != nil {
				log.Printf("⚠️  Warning: Error indexing activities: %v", err)
			} else {
				log.Printf("✅ Indexed %d activities", len(activities))
			}
		}
	}

	// 5. Conectar a RabbitMQ como consumidor (Client externo)
	rabbitConsumer, err := clients.NewRabbitMQConsumer(
		cfg.RabbitMQURL,
		cfg.RabbitMQExchange,
		cfg.RabbitMQQueue,
		searchService,
		cacheService,
	)
	if err != nil {
		log.Printf("⚠️  Warning: No se pudo conectar a RabbitMQ: %v", err)
		// Continuar sin RabbitMQ para desarrollo
	} else {
		defer rabbitConsumer.Close()

		// Iniciar consumidor
		if err := rabbitConsumer.Start(); err != nil {
			log.Printf("⚠️  Warning: Error iniciando consumidor: %v", err)
		}
	}

	// 6. Crear controllers con DI
	searchController := controllers.NewSearchController(searchService, cacheService)

	// 7. Configurar Gin Router
	router := gin.Default()
	router.Use(middleware.CORS())

	// 8. Registrar Rutas
	registerRoutes(router, searchController)

	// 9. Iniciar servidor
	log.Printf("🚀 Search API corriendo en puerto %s", cfg.Port)
	log.Println("📦 Arquitectura: Controllers → Services → Clients/Repositories")
	log.Println("💉 Dependency Injection: Activada")
	log.Printf("🔍 Sistema de búsqueda: Solr + MySQL FULLTEXT fallback")
	log.Printf("💾 Caché de dos niveles activado (Local: %ds, Memcached: %ds)", cfg.LocalCacheTTL, cfg.CacheTTL)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}

// registerRoutes - Registra todas las rutas HTTP
func registerRoutes(router *gin.Engine, searchController *controllers.SearchController) {
	// Health check
	router.GET("/healthz", searchController.HealthCheck)

	// Rutas de búsqueda
	searchRoutes := router.Group("/search")
	{
		searchRoutes.POST("", searchController.Search)              // Búsqueda avanzada
		searchRoutes.GET("", searchController.QuickSearch)          // Búsqueda rápida (query params)
		searchRoutes.GET("/stats", searchController.GetStats)       // Estadísticas del índice
		searchRoutes.GET("/categories", searchController.GetCategories) // Categorías únicas
		searchRoutes.GET("/:id", searchController.GetDocument)      // Obtener documento por ID
		searchRoutes.POST("/index", searchController.IndexDocument) // Indexar documento manualmente
		searchRoutes.DELETE("/:id", searchController.DeleteDocument) // Eliminar documento
	}
}
