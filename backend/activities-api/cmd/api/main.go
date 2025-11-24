package main

import (
	"activities-api/internal/clients"
	"activities-api/internal/config"
	"activities-api/internal/controllers"
	"activities-api/internal/middleware"
	"activities-api/internal/repository"
	"activities-api/internal/services"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Variables globales para health check
var (
	rabbitPublisher  *clients.RabbitMQEventPublisher
	actividadesRepo  *repository.MySQLActividadesRepository
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// ========== CAPA DE DATOS (REPOSITORY) ==========
	// Crear repositorio de actividades (comparte conexión DB)
	actividadesRepo = repository.NewMySQLActividadesRepository(cfg.MySQL)
	if actividadesRepo == nil {
		log.Fatal("Failed to initialize actividades repository")
	}

	// Crear repositorio de inscripciones (comparte la misma DB)
	inscripcionesRepo := repository.NewMySQLInscripcionesRepository(actividadesRepo.GetDB())

	// TODO: Cuando el equipo implemente Sucursales:
	// sucursalesRepo := repository.NewMySQLSucursalesRepository(actividadesRepo.GetDB())

	// ========== RABBITMQ EVENT PUBLISHER ==========
	// Inicializar RabbitMQ con fallback a NullEventPublisher
	var eventPublisher services.EventPublisher
	var err error
	rabbitPublisher, err = clients.NewRabbitMQEventPublisher(cfg.RabbitMQURL, cfg.RabbitMQExchange)
	if err != nil {
		log.Printf("⚠️ Warning: No se pudo conectar a RabbitMQ: %v", err)
		log.Println("⚠️ Usando NullEventPublisher como fallback")
		eventPublisher = clients.NewNullEventPublisher()
	} else {
		eventPublisher = rabbitPublisher
		defer rabbitPublisher.Close()
		log.Println("✅ RabbitMQ Publisher inicializado correctamente")
	}

	// ========== CAPA DE NEGOCIO (SERVICES) ==========
	// Crear servicios con dependency injection (incluyendo eventPublisher)
	actividadesService := services.NewActividadesService(actividadesRepo, eventPublisher)
	inscripcionesService := services.NewInscripcionesService(inscripcionesRepo, actividadesRepo, eventPublisher)
	// TODO: sucursalesService := services.NewSucursalesService(sucursalesRepo)

	// ========== CAPA DE PRESENTACIÓN (CONTROLLERS) ==========
	// Crear controllers con dependency injection
	actividadesController := controllers.NewActividadesController(actividadesService)
	inscripcionesController := controllers.NewInscripcionesController(inscripcionesService)
	// TODO: sucursalesController := controllers.NewSucursalesController(sucursalesService)

	// ========== CONFIGURACIÓN DE GIN ==========
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/healthz", healthCheckHandler)

	// ========== RUTAS PÚBLICAS ==========
	// Actividades (solo lectura sin auth)
	router.GET("/actividades", actividadesController.List)
	router.GET("/actividades/buscar", actividadesController.Search)
	router.GET("/actividades/:id", actividadesController.GetByID)

	// TODO: Sucursales (solo lectura sin auth)
	// router.GET("/sucursales", sucursalesController.List)
	// router.GET("/sucursales/:id", sucursalesController.GetByID)

	// ========== RUTAS PROTEGIDAS (REQUIEREN JWT) ==========
	protected := router.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(cfg.JWT.Secret))
	{
		// Inscripciones (requieren autenticación)
		protected.GET("/inscripciones", inscripcionesController.List)
		protected.POST("/inscripciones", inscripcionesController.Create)
		protected.DELETE("/inscripciones", inscripcionesController.Deactivate)
	}

	// ========== RUTAS DE ADMIN (REQUIEREN JWT + ADMIN) ==========
	adminOnly := protected.Group("/")
	adminOnly.Use(middleware.AdminOnlyMiddleware())
	{
		// Actividades (CRUD completo solo admin)
		adminOnly.POST("/actividades", actividadesController.Create)
		adminOnly.PUT("/actividades/:id", actividadesController.Update)
		adminOnly.DELETE("/actividades/:id", actividadesController.Delete)

		// TODO: Sucursales (CRUD completo solo admin)
		// adminOnly.POST("/sucursales", sucursalesController.Create)
		// adminOnly.PUT("/sucursales/:id", sucursalesController.Update)
		// adminOnly.DELETE("/sucursales/:id", sucursalesController.Delete)
	}

	// ========== INICIAR SERVIDOR ==========
	port := cfg.Port
	log.Printf("🚀 Activities API running on port %s", port)
	log.Printf("📋 Endpoints disponibles:")
	log.Printf("   GET    /healthz")
	log.Printf("   GET    /actividades")
	log.Printf("   GET    /actividades/buscar?id=&titulo=&horario=&categoria=")
	log.Printf("   GET    /actividades/:id")
	log.Printf("   POST   /actividades (admin)")
	log.Printf("   PUT    /actividades/:id (admin)")
	log.Printf("   DELETE /actividades/:id (admin)")
	log.Printf("   GET    /inscripciones (auth)")
	log.Printf("   POST   /inscripciones (auth)")
	log.Printf("   DELETE /inscripciones (auth)")

	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheckHandler(ctx *gin.Context) {
	health := gin.H{
		"status":  "ok",
		"service": "activities-api",
		"version": "1.0.0",
		"checks":  gin.H{},
	}

	// Check RabbitMQ connectivity
	if rabbitPublisher != nil && rabbitPublisher.Conn != nil && !rabbitPublisher.Conn.IsClosed() {
		health["checks"].(gin.H)["rabbitmq"] = "connected"
	} else {
		health["checks"].(gin.H)["rabbitmq"] = "disconnected"
		health["status"] = "degraded"
	}

	// Check MySQL connectivity
	if actividadesRepo != nil && actividadesRepo.GetDB() != nil {
		sqlDB, err := actividadesRepo.GetDB().DB()
		if err == nil && sqlDB.Ping() == nil {
			health["checks"].(gin.H)["mysql"] = "connected"
		} else {
			health["checks"].(gin.H)["mysql"] = "disconnected"
			health["status"] = "degraded"
		}
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	ctx.JSON(statusCode, health)
}
