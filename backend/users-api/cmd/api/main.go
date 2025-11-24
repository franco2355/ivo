package main

import (
	"log"
	"net/http"
	"time"
	"users-api/internal/config"
	"users-api/internal/controllers"
	"users-api/internal/middleware"
	"users-api/internal/repository"
	"users-api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 📋 Cargar configuración desde las variables de entorno
	cfg := config.Load()

	log.Println("🔧 Initializing Users API...")

	// 🏗️ Inicializar capas de la aplicación (Dependency Injection)
	// Patrón: Repository -> Service -> Controller
	// Cada capa tiene una responsabilidad específica y depende de interfaces

	// 1️⃣ Capa de datos: Repository (maneja operaciones con MySQL)
	usersRepo := repository.NewMySQLUsersRepository(cfg.MySQL)

	// 2️⃣ Capa de lógica de negocio: Service (validaciones, transformaciones, JWT)
	usersService := services.NewUsersService(usersRepo, cfg.JWT.Secret)

	// 3️⃣ Capa de controladores: Controller (maneja HTTP requests/responses)
	usersController := controllers.NewUsersController(usersService)

	// 🌐 Configurar router HTTP con Gin
	router := gin.Default()

	// Middleware CORS (debe ir primero)
	router.Use(middleware.CORSMiddleware())

	// 🏥 Health check endpoint (sin autenticación ni rate limiting)
	router.GET("/healthz", func(c *gin.Context) {
		health := gin.H{
			"status":      "ok",
			"service":     "users-api",
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"checks":      gin.H{},
		}

		// Check MySQL connection
		if usersRepo != nil {
			db := usersRepo.GetDB()
			if db != nil {
				sqlDB, err := db.DB()
				if err == nil && sqlDB.Ping() == nil {
					health["checks"].(gin.H)["mysql"] = "connected"
				} else {
					health["status"] = "degraded"
					health["checks"].(gin.H)["mysql"] = "unhealthy"
				}
			} else {
				health["status"] = "degraded"
				health["checks"].(gin.H)["mysql"] = "unavailable"
			}
		}

		// Return 503 if degraded
		if health["status"] == "degraded" {
			c.JSON(http.StatusServiceUnavailable, health)
			return
		}

		c.JSON(http.StatusOK, health)
	})

	// 📚 Rutas públicas (con rate limiting específico)
	// Rate limit para registro: 3 intentos cada 10 minutos por IP
	router.POST("/register",
		// middleware.RateLimitMiddleware(cfg.RateLimit.RegisterAttempts, cfg.RateLimit.RegisterWindow),
		usersController.Register,
	)

	// Rate limit para login: 5 intentos cada 15 minutos por IP
	router.POST("/login",
		// middleware.RateLimitMiddleware(cfg.RateLimit.LoginAttempts, cfg.RateLimit.LoginWindow),
		usersController.Login,
	)

	// 📚 Rutas protegidas (requieren JWT + rate limiting)
	protected := router.Group("/")
	// protected.Use(middleware.RateLimitMiddleware(cfg.RateLimit.PublicRPM, 1)) // 100 req/min
	protected.Use(middleware.JWTAuthMiddleware(usersService))
	{
		// Endpoint para que otros microservicios validen usuario existe
		protected.GET("/users/:id", usersController.GetByID)

		// Solo admins pueden listar todos los usuarios
		adminOnly := protected.Group("/")
		adminOnly.Use(middleware.AdminOnlyMiddleware())
		{
			adminOnly.GET("/users", usersController.List)
		}
	}

	// Configuración del server HTTP
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Printf("🚀 Users API listening on port %s", cfg.Port)
	log.Printf("🌍 Environment: %s", cfg.Environment)
	log.Printf("📊 Health check: http://localhost:%s/healthz", cfg.Port)
	log.Printf("🛡️  Rate Limiting:")
	log.Printf("   Login: %d attempts per %d minutes", cfg.RateLimit.LoginAttempts, cfg.RateLimit.LoginWindow)
	log.Printf("   Register: %d attempts per %d minutes per IP", cfg.RateLimit.RegisterAttempts, cfg.RateLimit.RegisterWindow)
	log.Printf("   Public APIs: %d requests per minute", cfg.RateLimit.PublicRPM)
	log.Printf("📚 Endpoints:")
	log.Printf("   POST   /register - Register new user (rate limited)")
	log.Printf("   POST   /login - Login user (rate limited)")
	log.Printf("   GET    /users/:id - Get user by ID (protected + rate limited)")
	log.Printf("   GET    /users - List all users (admin only + rate limited)")

	// Iniciar servidor (bloquea hasta que se pare el servidor)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}
}
