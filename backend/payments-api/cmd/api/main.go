package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/payments-api/internal/clients"
	"github.com/yourusername/payments-api/internal/config"
	"github.com/yourusername/payments-api/internal/controllers"
	"github.com/yourusername/payments-api/internal/dao"
	"github.com/yourusername/payments-api/internal/database"
	"github.com/yourusername/payments-api/internal/domain/dtos"
	"github.com/yourusername/payments-api/internal/gateways"
	"github.com/yourusername/payments-api/internal/middleware"
	"github.com/yourusername/payments-api/internal/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("🚀 Iniciando Payments API con arquitectura de gateways...")

	// ========== 1. CONFIGURACIÓN ==========
	cfg := config.LoadConfig()
	log.Printf("✅ Configuración cargada: Puerto=%s, MongoDB=%s", cfg.Port, cfg.MongoURI)

	// ========== 2. CONECTAR A MONGODB ==========
	mongoDB, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("❌ Error conectando a MongoDB: %v", err)
	}
	defer mongoDB.Close()
	log.Println("✅ Conectado a MongoDB exitosamente")

	// ========== 3. CAPA DE DATOS (REPOSITORY) ==========
	// Implementación del repository con MongoDB
	paymentRepo := dao.NewPaymentRepositoryMongo(mongoDB.Database)
	log.Println("✅ Repository inicializado (MongoDB)")

	// ========== 3.5. CREAR ÍNDICES DE MONGODB ==========
	if err := createPaymentIndexes(mongoDB); err != nil {
		log.Printf("⚠️  Error creando índices: %v", err)
	} else {
		log.Println("✅ Índices de MongoDB creados/verificados")
	}

	// ========== 4. CAPA DE GATEWAYS (FACTORY PATTERN) ==========
	// El factory permite crear instancias de gateways en runtime
	gatewayFactory := gateways.NewGatewayFactory(cfg)
	log.Println("✅ Gateway Factory inicializado")
	log.Printf("   Gateways soportados: %v", gatewayFactory.GetSupportedGateways())

	// ========== 4.5. RABBITMQ EVENT PUBLISHER (OPCIONAL) ==========
	// Si está configurado, publicará eventos cuando cambien estados de pagos
	var eventPublisher services.EventPublisher
	var rabbitMQClient *clients.RabbitMQPublisher
	if cfg.RabbitMQ.URL != "" {
		rabbitMQClient, err = clients.NewRabbitMQPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange)
		if err != nil {
			log.Printf("⚠️  RabbitMQ no disponible (continuando sin eventos): %v", err)
			eventPublisher = clients.NewNoOpEventPublisher()
			rabbitMQClient = nil
		} else {
			defer rabbitMQClient.Close()
			eventPublisher = clients.NewPaymentEventPublisher(rabbitMQClient)
			log.Println("✅ RabbitMQ Publisher inicializado")
		}
	} else {
		log.Println("ℹ️  RabbitMQ no configurado (eventos desactivados)")
		eventPublisher = clients.NewNoOpEventPublisher()
	}

	// ========== 5. CAPA DE NEGOCIO (SERVICES) ==========
	// Servicio completo con gateways integrados + eventos ⭐
	paymentService := services.NewPaymentService(paymentRepo, gatewayFactory, eventPublisher)
	log.Println("✅ Payment Service inicializado (con gateways + eventos)")

	// ========== 6. CAPA DE PRESENTACIÓN (CONTROLLERS) ==========
	// Payment controller (maneja todos los endpoints de pagos)
	paymentController := controllers.NewPaymentController(paymentService)

	// Webhook controller (procesa notificaciones de gateways)
	webhookController := controllers.NewWebhookController(gatewayFactory, paymentRepo, paymentService)
	log.Println("✅ Controllers inicializados (Payment + Webhook)")

	// ========== 7. CONFIGURAR GIN ROUTER ==========
	router := gin.Default()
	router.Use(middleware.CORS())

	// ========== 8. REGISTRAR RUTAS ==========
	registerRoutes(router, paymentController, webhookController, paymentService, mongoDB, rabbitMQClient, cfg)

	// ========== 9. INICIAR SERVIDOR ==========
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║         🎉 PAYMENTS API CON GATEWAYS INICIADA 🎉            ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Printf("🌐 Puerto: %s", cfg.Port)
	log.Println("📦 Arquitectura: Controllers → Services → Gateways → Repositories")
	log.Println("💉 Dependency Injection: Activada")
	log.Println("🏗️  Patrones: Strategy + Factory + Repository")
	log.Println("")
	log.Println("📋 Endpoints disponibles:")
	log.Println("   GET    /healthz                         - Health check")
	log.Println("   POST   /payments                        - Crear pago (básico)")
	log.Println("   GET    /payments                        - Listar todos los pagos")
	log.Println("   POST   /payments/process                - Pago único con Checkout Pro ⭐")
	log.Println("   POST   /payments/recurring              - Pago recurrente con Preapprovals ⭐")
	log.Println("   GET    /payments/:id                    - Obtener pago")
	log.Println("   GET    /payments/:id/sync               - Sincronizar estado con gateway ⭐")
	log.Println("   POST   /payments/:id/refund             - Procesar reembolso ⭐")
	log.Println("   POST   /webhooks/mercadopago            - Webhook de Mercado Pago")
	log.Println("   POST   /webhooks/:gateway               - Webhook genérico")
	log.Println("")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}

// registerRoutes - Registra todas las rutas HTTP
func registerRoutes(
	router *gin.Engine,
	paymentController *controllers.PaymentController,
	webhookController *controllers.WebhookController,
	paymentService *services.PaymentService,
	mongoDB *database.MongoDB,
	rabbitMQClient *clients.RabbitMQPublisher,
	cfg *config.Config,
) {
	// ========== HEALTH CHECK ==========
	router.GET("/healthz", func(ctx *gin.Context) {
		health := gin.H{
			"status":  "ok",
			"service": "payments-api",
			"checks":  gin.H{},
		}

		// Check MongoDB
		if mongoDB != nil && mongoDB.Client != nil {
			ctxTimeout, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
			defer cancel()
			if err := mongoDB.Client.Ping(ctxTimeout, nil); err == nil {
				health["checks"].(gin.H)["mongodb"] = "connected"
			} else {
				health["status"] = "degraded"
				health["checks"].(gin.H)["mongodb"] = "unhealthy"
			}
		} else {
			health["status"] = "degraded"
			health["checks"].(gin.H)["mongodb"] = "unavailable"
		}

		// Check RabbitMQ
		if rabbitMQClient != nil && rabbitMQClient.Conn != nil && !rabbitMQClient.Conn.IsClosed() {
			health["checks"].(gin.H)["rabbitmq"] = "connected"
		} else {
			health["checks"].(gin.H)["rabbitmq"] = "disconnected"
		}

		// Return 503 if degraded
		if health["status"] == "degraded" {
			ctx.JSON(http.StatusServiceUnavailable, health)
			return
		}

		ctx.JSON(http.StatusOK, health)
	})

	// ========== MIDDLEWARE DE AUTENTICACIÓN ==========
	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // Valor por defecto para desarrollo
	}
	authMiddleware := middleware.JWTAuthMiddleware(jwtSecret)
	adminMiddleware := middleware.AdminOnlyMiddleware()

	// ========== RUTAS BÁSICAS (compatibilidad) ==========
	paymentRoutes := router.Group("/payments")
	{
		// Usuario autenticado puede crear su propio pago (al suscribirse)
		// Admin también puede crear pagos para otros usuarios
		paymentRoutes.POST("", authMiddleware, paymentController.CreatePayment)

		// Admin puede ver todos los pagos
		paymentRoutes.GET("", authMiddleware, adminMiddleware, paymentController.GetAllPayments)

		// Usuario autenticado puede ver un pago específico (validar que es suyo o es admin)
		paymentRoutes.GET("/:id", authMiddleware, paymentController.GetPayment)

		// Usuario puede ver sus propios pagos
		paymentRoutes.GET("/user/:user_id", authMiddleware, paymentController.GetPaymentsByUser)

		// Admin puede filtrar por entidad y status
		paymentRoutes.GET("/entity", authMiddleware, adminMiddleware, paymentController.GetPaymentsByEntity)
		paymentRoutes.GET("/status/:status", authMiddleware, adminMiddleware, paymentController.GetPaymentsByStatus)

		// Solo admin puede actualizar status manualmente
		paymentRoutes.PATCH("/:id/status", authMiddleware, adminMiddleware, paymentController.UpdatePaymentStatus)
		paymentRoutes.POST("/:id/process", paymentController.ProcessPayment)

		// ========== RUTAS MEJORADAS CON GATEWAYS ⭐ ==========
		// Crear y procesar pago ÚNICO con Checkout Pro
		paymentRoutes.POST("/process", createPaymentWithGatewayHandler(paymentService))

		// Crear y procesar pago RECURRENTE con Preapprovals ⭐
		paymentRoutes.POST("/recurring", createRecurringPaymentHandler(paymentService))

		// Sincronizar estado con gateway
		paymentRoutes.GET("/:id/sync", syncPaymentStatusHandler(paymentService))

		// Procesar reembolso
		paymentRoutes.POST("/:id/refund", refundPaymentHandler(paymentService))

		// ========== RUTAS PARA PAGOS EN EFECTIVO (ADMIN) ⭐ ==========
		// Aprobar pago en efectivo (solo admin)
		paymentRoutes.POST("/:id/approve", paymentController.ApproveCashPayment)

		// Rechazar pago en efectivo (solo admin)
		paymentRoutes.POST("/:id/reject", paymentController.RejectCashPayment)
	}

	// ========== WEBHOOKS ==========
	webhookRoutes := router.Group("/webhooks")
	{
		// Webhook específico de Mercado Pago
		webhookRoutes.POST("/mercadopago", webhookController.HandleMercadoPagoWebhook)

		// Webhook genérico (funciona para cualquier gateway)
		webhookRoutes.POST("/:gateway", webhookController.HandleGenericWebhook)
	}
}

// ========== HANDLERS PARA RUTAS MEJORADAS ==========

// createPaymentWithGatewayHandler - Crea un pago y lo procesa en el gateway
func createPaymentWithGatewayHandler(service *services.PaymentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			EntityType     string                 `json:"entity_type" binding:"required"`
			EntityID       string                 `json:"entity_id" binding:"required"`
			UserID         string                 `json:"user_id" binding:"required"`
			Amount         float64                `json:"amount" binding:"required,gt=0"`
			Currency       string                 `json:"currency" binding:"required"`
			PaymentMethod  string                 `json:"payment_method" binding:"required"`
			PaymentGateway string                 `json:"payment_gateway" binding:"required"` // "mercadopago", "cash"
			IdempotencyKey string                 `json:"idempotency_key,omitempty"`         // UUID para prevenir duplicados
			CallbackURL    string                 `json:"callback_url,omitempty"`
			WebhookURL     string                 `json:"webhook_url,omitempty"`
			Metadata       map[string]interface{} `json:"metadata,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Convertir a DTO
		dto := dtos.CreatePaymentRequest{
			EntityType:     req.EntityType,
			EntityID:       req.EntityID,
			UserID:         req.UserID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			PaymentMethod:  req.PaymentMethod,
			PaymentGateway: req.PaymentGateway,
			IdempotencyKey: req.IdempotencyKey, // ⭐ Pasar idempotency key
			CallbackURL:    req.CallbackURL,
			WebhookURL:     req.WebhookURL,
			Metadata:       req.Metadata,
		}

		result, err := service.ProcessPaymentWithGateway(c.Request.Context(), dto)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, result)
	}
}

// syncPaymentStatusHandler - Sincroniza el estado de un pago con el gateway
func syncPaymentStatusHandler(service *services.PaymentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Param("id")

		result, err := service.SyncPaymentStatus(c.Request.Context(), paymentID)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, result)
	}
}

// createRecurringPaymentHandler - Crea un pago recurrente (débito automático)
func createRecurringPaymentHandler(service *services.PaymentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			EntityType     string                 `json:"entity_type" binding:"required"`
			EntityID       string                 `json:"entity_id" binding:"required"`
			UserID         string                 `json:"user_id" binding:"required"`
			Amount         float64                `json:"amount" binding:"required,gt=0"`
			Currency       string                 `json:"currency" binding:"required"`
			PaymentMethod  string                 `json:"payment_method" binding:"required"`
			PaymentGateway string                 `json:"payment_gateway" binding:"required"` // "mercadopago"
			IdempotencyKey string                 `json:"idempotency_key,omitempty"`         // UUID para prevenir duplicados
			Frequency      int                    `json:"frequency" binding:"required,gt=0"`  // 1, 2, 3...
			FrequencyType  string                 `json:"frequency_type" binding:"required"`  // "months", "days", "weeks"
			Metadata       map[string]interface{} `json:"metadata,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Convertir a DTO
		dto := dtos.CreatePaymentRequest{
			EntityType:     req.EntityType,
			EntityID:       req.EntityID,
			UserID:         req.UserID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			PaymentMethod:  req.PaymentMethod,
			PaymentGateway: req.PaymentGateway,
			IdempotencyKey: req.IdempotencyKey, // ⭐ Pasar idempotency key
			Metadata:       req.Metadata,
		}

		result, err := service.ProcessRecurringPayment(c.Request.Context(), dto, req.Frequency, req.FrequencyType)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, result)
	}
}

// refundPaymentHandler - Procesa un reembolso a través del gateway
func refundPaymentHandler(service *services.PaymentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Param("id")

		var req struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err := service.RefundPayment(c.Request.Context(), paymentID, req.Amount)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "Reembolso procesado exitosamente",
		})
	}
}

// createPaymentIndexes - Crea índices únicos y optimizaciones en la colección de payments
func createPaymentIndexes(mongoDB *database.MongoDB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mongoDB.Database.Collection("payments")

	// Índice único en idempotency_key para prevenir pagos duplicados
	// sparse: true -> solo indexa documentos que tienen el campo (permite null)
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "idempotency_key", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true).
			SetName("idx_idempotency_key_unique"),
	}

	indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("⚠️  Error creando índice idempotency_key: %v", err)
		return err
	}

	log.Printf("   Índice creado: %s", indexName)
	return nil
}
