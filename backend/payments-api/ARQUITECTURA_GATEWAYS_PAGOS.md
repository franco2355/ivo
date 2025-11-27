# Arquitectura de Gateways de Pago

## Índice
- [Introducción](#introducción)
- [Patrones Utilizados](#patrones-utilizados)
- [Arquitectura Visual](#arquitectura-visual)
- [Componentes Principales](#componentes-principales)
- [Flujo de Datos](#flujo-de-datos)
- [Extensibilidad](#extensibilidad)
- [Referencias](#referencias)

---

## Introducción

Esta documentación describe la arquitectura genérica y extensible implementada para integrar múltiples pasarelas de pago (Mercado Pago, Stripe, PayPal, etc.) en el microservicio de Payments.

### Objetivos de la Arquitectura

- **Extensibilidad**: Agregar nuevas pasarelas sin modificar código existente
- **Desacoplamiento**: La lógica de negocio no conoce detalles de implementación de pasarelas
- **Testabilidad**: Facilitar testing unitario con implementaciones de prueba
- **Mantenibilidad**: Cada pasarela aislada en su propio módulo
- **Flexibilidad**: Cambiar de pasarela en runtime según necesidades del cliente

---

## Patrones Utilizados

### Patrones Arquitectónicos

#### 1. Gateway Pattern
**Nivel**: Arquitectónico (Integración)
**Propósito**: Encapsular el acceso a sistemas externos

```
Tu Aplicación  →  [Gateway]  →  Servicio Externo (Mercado Pago, Stripe, etc.)
```

- Abstrae la complejidad de comunicación HTTP/REST
- Maneja autenticación, timeouts, reintentos
- Traduce entre tu modelo de dominio y el de la API externa

#### 2. Repository Pattern
**Nivel**: Arquitectónico (Persistencia)
**Propósito**: Abstraer el acceso a datos

```
Service  →  [Repository Interface]  →  MongoDB/PostgreSQL/etc.
```

#### 3. DTO (Data Transfer Object)
**Nivel**: Arquitectónico (Transferencia)
**Propósito**: Objetos diseñados para transferir datos entre capas

```
API Request  →  [DTO]  →  Domain Entity  →  [DTO]  →  API Response
```

### Patrones de Diseño (GoF)

#### 4. Strategy Pattern
**Categoría**: Comportamiento
**Propósito**: Definir familia de algoritmos intercambiables

```go
// Una interfaz, múltiples implementaciones
type PaymentGateway interface {
    CreatePayment(...) (*PaymentResult, error)
}

// Estrategias concretas
type MercadoPagoGateway struct {}
type StripeGateway struct {}
type PayPalGateway struct {}
```

#### 5. Factory Pattern
**Categoría**: Creacional
**Propósito**: Encapsular la creación de objetos

```go
// En lugar de: gateway := NewMercadoPagoGateway(...)
// Usamos:
gateway := factory.CreateGateway("mercadopago")
```

#### 6. Dependency Injection
**Categoría**: Estructural
**Propósito**: Inyectar dependencias en lugar de crearlas internamente

```go
// Las dependencias se inyectan en el constructor
func NewPaymentService(repo Repository, factory *GatewayFactory) *PaymentService {
    return &PaymentService{
        repo:    repo,
        factory: factory,
    }
}
```

---

## Arquitectura Visual

```
┌─────────────────────────────────────────────────────────────────────┐
│                    CAPA DE PRESENTACIÓN (HTTP)                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────────┐              ┌──────────────────┐           │
│  │ Payment          │              │ Webhook          │           │
│  │ Controller       │              │ Controller       │           │
│  └────────┬─────────┘              └────────┬─────────┘           │
│           │                                 │                      │
│           │ (DTOs)                          │ (Raw Payload)        │
│           │                                 │                      │
├───────────┼─────────────────────────────────┼──────────────────────┤
│           │       CAPA DE APLICACIÓN        │                      │
│           ▼                                 ▼                      │
│  ┌─────────────────────────────────────────────────────┐          │
│  │         Payment Service (Business Logic)            │          │
│  │                                                      │          │
│  │  ┌────────────────────────────────────────────┐    │          │
│  │  │  Repository Interface                      │    │          │
│  │  │  (Abstracción de Persistencia)             │    │          │
│  │  └────────────────────────────────────────────┘    │          │
│  │                                                      │          │
│  │  ┌────────────────────────────────────────────┐    │          │
│  │  │  Gateway Factory                           │    │          │
│  │  │  (Creación de Gateways)                    │    │          │
│  │  └───────────────┬────────────────────────────┘    │          │
│  └──────────────────┼─────────────────────────────────┘          │
│                     │                                             │
│                     │ Strategy Pattern                            │
│                     │ (Selección en Runtime)                      │
│                     │                                             │
├─────────────────────┼─────────────────────────────────────────────┤
│  CAPA DE INFRAESTRUCTURA (Gateways)                               │
│                     │                                             │
│         ┌───────────┴───────────┬──────────────┬──────────┐      │
│         ▼                       ▼              ▼          ▼      │
│  ┌─────────────┐      ┌─────────────┐   ┌──────────┐ ┌───────┐ │
│  │ Mercado Pago│      │   Stripe    │   │  PayPal  │ │ Cash  │ │
│  │  Gateway    │      │  Gateway    │   │ Gateway  │ │Gateway│ │
│  └──────┬──────┘      └──────┬──────┘   └────┬─────┘ └───┬───┘ │
│         │                    │               │           │      │
│         │ (HTTP/REST)        │               │           │      │
│         ▼                    ▼               ▼           ▼      │
├─────────────────────────────────────────────────────────────────┤
│  SERVICIOS EXTERNOS                                              │
│                                                                   │
│  [Mercado Pago API]    [Stripe API]    [PayPal API]   [Manual] │
└───────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────┐
│  CAPA DE PERSISTENCIA                                             │
│                                                                    │
│  ┌────────────────────────────────────────────────┐              │
│  │  Payment Repository Mongo (Implementación)     │              │
│  └────────────────────┬───────────────────────────┘              │
│                       │                                           │
│                       ▼                                           │
│               [MongoDB Database]                                  │
└───────────────────────────────────────────────────────────────────┘
```

### Flujo de Llamadas

```
1. Cliente HTTP
   │
   ├─→ POST /payments
   │   │
   │   └─→ PaymentController.CreatePayment()
   │       │
   │       └─→ PaymentService.ProcessPaymentWithGateway(dto)
   │           │
   │           ├─→ Repository.Create(payment) // Guardar en DB
   │           │
   │           ├─→ GatewayFactory.CreateGateway("mercadopago")
   │           │   └─→ return MercadoPagoGateway instance
   │           │
   │           ├─→ gateway.CreatePayment(request)
   │           │   │
   │           │   └─→ HTTP POST https://api.mercadopago.com/v1/payments
   │           │       └─→ return PaymentResult
   │           │
   │           └─→ Repository.UpdateStatus(paymentID, result.Status)
   │
   └─→ Webhook Notification (Async)
       │
       └─→ WebhookController.HandleMercadoPagoWebhook()
           │
           └─→ gateway.ProcessWebhook(payload)
               │
               └─→ PaymentService.UpdatePaymentFromWebhook()
                   └─→ Repository.UpdateStatus(...)
```

---

## Componentes Principales

### 1. PaymentGateway Interface

**Ubicación**: `internal/gateways/payment_gateway.go`

Define el contrato que todas las pasarelas deben implementar:

```go
type PaymentGateway interface {
    GetName() string
    CreatePayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error)
    GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error)
    RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error)
    CancelPayment(ctx context.Context, transactionID string) error
    ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookEvent, error)
    ValidateCredentials(ctx context.Context) error
}
```

**Métodos Principales**:

| Método | Propósito | Retorna |
|--------|-----------|---------|
| `GetName()` | Identificador del gateway | `string` ("mercadopago", "stripe") |
| `CreatePayment()` | Crear pago en la pasarela | `PaymentResult` con transactionID, status, URL |
| `GetPaymentStatus()` | Consultar estado de un pago | `PaymentStatus` con estado actual |
| `RefundPayment()` | Procesar reembolso | `RefundResult` con detalles del reembolso |
| `CancelPayment()` | Cancelar pago pendiente | `error` si falla |
| `ProcessWebhook()` | Procesar notificaciones | `WebhookEvent` parseado |
| `ValidateCredentials()` | Verificar credenciales | `error` si son inválidas |

### 2. GatewayFactory

**Ubicación**: `internal/gateways/factory.go`

Responsable de crear instancias de gateways según el nombre solicitado.

```go
type GatewayFactory struct {
    config *config.Config
}

func (f *GatewayFactory) CreateGateway(gatewayName string) (PaymentGateway, error) {
    switch gatewayName {
    case "mercadopago":
        return NewMercadoPagoGateway(
            f.config.MercadoPago.AccessToken,
            f.config.MercadoPago.PublicKey,
            f.config.MercadoPago.WebhookSecret,
        ), nil
    case "stripe":
        return NewStripeGateway(...)
    // ...más gateways
    default:
        return nil, fmt.Errorf("gateway no soportado: %s", gatewayName)
    }
}
```

**Ventajas**:
- Centraliza la lógica de creación
- Maneja inyección de configuración automáticamente
- Facilita agregar nuevos gateways (solo un `case` nuevo)

### 3. Implementaciones Concretas

Cada pasarela tiene su propia implementación del contrato `PaymentGateway`.

**Estructura de carpetas**:
```
internal/gateways/
├── payment_gateway.go          # Interfaz + DTOs comunes
├── factory.go                  # Factory
├── mercadopago/
│   └── mercadopago_gateway.go  # Implementación Mercado Pago
├── stripe/
│   └── stripe_gateway.go       # Implementación Stripe
├── paypal/
│   └── paypal_gateway.go       # Implementación PayPal
└── cash/
    └── cash_gateway.go         # Gateway para pagos en efectivo (manual)
```

### 4. DTOs Genéricos

**Ubicación**: `internal/gateways/payment_gateway.go`

Estructuras de datos que abstraen las diferencias entre pasarelas:

```go
// Request genérico
type PaymentRequest struct {
    Amount          float64
    Currency        string
    Description     string
    CustomerEmail   string
    CustomerName    string
    PaymentMethod   string
    Metadata        map[string]interface{}
    CallbackURL     string
    WebhookURL      string
    ExternalID      string // Tu payment ID interno
}

// Response genérico
type PaymentResult struct {
    TransactionID   string    // ID del gateway
    Status          string    // pending, completed, failed
    PaymentURL      string    // URL para completar pago
    QRCode          string    // Código QR (si aplica)
    ExternalData    map[string]interface{} // Datos específicos
    CreatedAt       time.Time
}
```

### 5. Payment Service Enhanced

**Ubicación**: `internal/services/payment_service_enhanced.go`

Orquesta la lógica de negocio integrando repository y gateways.

```go
type PaymentServiceEnhanced struct {
    paymentRepo    repository.PaymentRepository
    gatewayFactory *gateways.GatewayFactory
}

func (s *PaymentServiceEnhanced) ProcessPaymentWithGateway(
    ctx context.Context,
    req dtos.CreatePaymentRequest,
) (dtos.PaymentResponse, error) {
    // 1. Crear registro en DB
    payment := entities.Payment{...}
    s.paymentRepo.Create(ctx, &payment)

    // 2. Obtener gateway apropiado
    gateway, _ := s.gatewayFactory.CreateGateway(req.PaymentGateway)

    // 3. Procesar en pasarela externa
    result, _ := gateway.CreatePayment(ctx, gatewayRequest)

    // 4. Actualizar con resultado
    s.paymentRepo.UpdateStatus(ctx, payment.ID, result.Status, result.TransactionID)

    return dtoResponse, nil
}
```

---

## Flujo de Datos

### Caso de Uso 1: Crear Pago

```
[Cliente] → POST /payments
            {
              "entity_type": "subscription",
              "entity_id": "sub_123",
              "user_id": "user_456",
              "amount": 99.99,
              "currency": "ARS",
              "payment_method": "credit_card",
              "payment_gateway": "mercadopago"
            }
            ↓
[Controller] → Validar request
            ↓
[Service] → 1. Crear Payment entity (status: pending)
            2. Repository.Create(payment) → MongoDB
            3. Factory.CreateGateway("mercadopago")
            4. MercadoPagoGateway.CreatePayment(request)
            ↓
[Gateway] → HTTP POST https://api.mercadopago.com/v1/payments
            {
              "transaction_amount": 99.99,
              "description": "Pago subscription #sub_123",
              ...
            }
            ↓
[Mercado Pago] → Response
                 {
                   "id": 123456789,
                   "status": "pending",
                   "init_point": "https://mercadopago.com/checkout/..."
                 }
            ↓
[Gateway] → Mapear respuesta a PaymentResult{
              TransactionID: "123456789",
              Status: "pending",
              PaymentURL: "https://mercadopago.com/checkout/..."
            }
            ↓
[Service] → Repository.UpdateStatus(payment.ID, "pending", "123456789")
            ↓
[Controller] → Response 201
               {
                 "id": "507f1f77bcf86cd799439011",
                 "status": "pending",
                 "transaction_id": "123456789",
                 "metadata": {
                   "payment_url": "https://mercadopago.com/checkout/..."
                 }
               }
```

### Caso de Uso 2: Webhook (Notificación de Pago Aprobado)

```
[Mercado Pago] → POST /webhooks/mercadopago
                 {
                   "action": "payment.updated",
                   "data": { "id": "123456789" },
                   "type": "payment"
                 }
                 ↓
[WebhookController] → 1. Leer payload
                      2. Factory.CreateGateway("mercadopago")
                      3. Gateway.ProcessWebhook(payload, headers)
                      ↓
[Gateway] → 1. Validar firma (si aplica)
            2. Parsear webhook
            3. Consultar estado: GetPaymentStatus("123456789")
            4. Retornar WebhookEvent{
                 EventType: "payment.updated",
                 TransactionID: "123456789",
                 Status: "completed"
               }
            ↓
[Service] → 1. Buscar payment por TransactionID
            2. Repository.UpdateStatus(paymentID, "completed", "123456789")
            3. Notificar a otros servicios (eventos, emails, etc.)
            ↓
[Controller] → Response 200 OK
```

---

## Extensibilidad

### Agregar una Nueva Pasarela (Ejemplo: Stripe)

**Paso 1**: Crear implementación

```go
// internal/gateways/stripe/stripe_gateway.go
package stripe

type StripeGateway struct {
    secretKey     string
    publicKey     string
    webhookSecret string
}

func NewStripeGateway(secretKey, publicKey, webhookSecret string) *StripeGateway {
    return &StripeGateway{...}
}

func (s *StripeGateway) GetName() string {
    return "stripe"
}

func (s *StripeGateway) CreatePayment(ctx context.Context, req gateways.PaymentRequest) (*gateways.PaymentResult, error) {
    // Implementar usando SDK de Stripe o HTTP directo
    // ...
}

// Implementar resto de métodos de la interfaz...
```

**Paso 2**: Registrar en Factory

```go
// internal/gateways/factory.go
func (f *GatewayFactory) CreateGateway(gatewayName string) (PaymentGateway, error) {
    switch gatewayName {
    case "mercadopago":
        return NewMercadoPagoGateway(...)
    case "stripe":  // ← AGREGAR AQUÍ
        return stripe.NewStripeGateway(
            f.config.Stripe.SecretKey,
            f.config.Stripe.PublicKey,
            f.config.Stripe.WebhookSecret,
        ), nil
    // ...
    }
}
```

**Paso 3**: Configurar credenciales

```bash
# .env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLIC_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

```go
// internal/config/config.go
type Config struct {
    // ...
    Stripe struct {
        SecretKey     string `env:"STRIPE_SECRET_KEY"`
        PublicKey     string `env:"STRIPE_PUBLIC_KEY"`
        WebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`
    }
}
```

**Paso 4**: Usar

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.00,
    "currency": "USD",
    "payment_gateway": "stripe"  # ← Cambio aquí
  }'
```

**¡Listo!** No modificaste ninguna lógica de negocio existente.

---

## Referencias

### Patrones de Diseño
- **Gang of Four (GoF)**: "Design Patterns: Elements of Reusable Object-Oriented Software" (1994)
  - Strategy Pattern
  - Factory Pattern

### Patrones Arquitectónicos
- **Martin Fowler**: "Patterns of Enterprise Application Architecture" (2002)
  - Gateway Pattern
  - Repository Pattern
  - Data Transfer Object (DTO)

### Principios SOLID
- **Open/Closed Principle**: Abierto para extensión, cerrado para modificación
  - Puedes agregar nuevos gateways sin modificar código existente
- **Dependency Inversion Principle**: Depender de abstracciones, no de implementaciones concretas
  - El servicio depende de la interfaz `PaymentGateway`, no de `MercadoPagoGateway`
- **Single Responsibility Principle**: Cada clase tiene una única razón para cambiar
  - Gateway: solo integración externa
  - Service: solo lógica de negocio
  - Repository: solo persistencia

### Recursos Adicionales
- [Martin Fowler - Gateway Pattern](https://martinfowler.com/eaaCatalog/gateway.html)
- [Refactoring Guru - Strategy Pattern](https://refactoring.guru/design-patterns/strategy)
- [Refactoring Guru - Factory Pattern](https://refactoring.guru/design-patterns/factory-method)

---

## Glosario

| Término | Definición |
|---------|------------|
| **Gateway** | Objeto que encapsula acceso a un sistema externo |
| **Strategy** | Patrón que permite intercambiar algoritmos en runtime |
| **Factory** | Patrón que encapsula creación de objetos |
| **DTO** | Objeto diseñado para transferir datos entre capas |
| **Repository** | Abstracción del acceso a datos |
| **Dependency Injection** | Inyectar dependencias en lugar de crearlas |
| **Pasarela de Pago** | Servicio externo que procesa pagos (Mercado Pago, Stripe, etc.) |
| **Webhook** | Notificación HTTP que envía un servicio externo cuando ocurre un evento |
| **Transaction ID** | Identificador único del pago en la pasarela externa |

---

**Última actualización**: 2025-10-21
**Autor**: Equipo de Arquitectura - Payments API
**Versión**: 1.0.0
