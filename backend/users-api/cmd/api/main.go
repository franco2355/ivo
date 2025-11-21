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

	// 🏥 Health check endpoint (sin autenticación)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "users-api",
			"version": "1.0.0",
		})
	})

	// 📚 Rutas públicas (sin autenticación)
	router.POST("/register", usersController.Register)
	router.POST("/login", usersController.Login)

	// 📚 Rutas protegidas (requieren JWT)
	protected := router.Group("/")
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
	log.Printf("📊 Health check: http://localhost:%s/healthz", cfg.Port)
	log.Printf("📚 Endpoints:")
	log.Printf("   POST   /register - Register new user")
	log.Printf("   POST   /login - Login user")
	log.Printf("   GET    /users/:id - Get user by ID (protected)")
	log.Printf("   GET    /users - List all users (admin only)")

	// Iniciar servidor (bloquea hasta que se pare el servidor)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}
}
