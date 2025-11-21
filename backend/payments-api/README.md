# Payments API

Microservicio **genérico** de gestión de pagos. Totalmente desacoplado del dominio, reutilizable en cualquier proyecto.

## Características

- **100% Agnóstico del Dominio**: Solo necesita saber quién pagó, cuánto y por qué
- **Flexible**: Campo `metadata` para información específica del contexto
- **Reutilizable**: Sirve para gimnasios, e-commerce, SaaS, etc.
- **Fácil Integración**: API REST simple y clara

## Tecnologías

- **Go 1.23**
- **MongoDB** - Base de datos NoSQL
- **Gin** - Framework web

## Estructura

```
payments-api/
├── cmd/api/                 # Punto de entrada
├── internal/
│   ├── config/             # Configuración
│   ├── database/           # MongoDB
│   ├── models/             # Modelo Payment genérico
│   ├── services/           # Lógica de negocio
│   ├── handlers/           # API REST
│   └── middleware/         # CORS, etc.
├── .env.example
├── Dockerfile
└── go.mod
```

## Modelo de Datos

```json
{
  "_id": "ObjectId",
  "entity_type": "subscription",     // Tipo de entidad (cualquiera)
  "entity_id": "507f...",           // ID de la entidad
  "user_id": "123",                 // ID del usuario que paga
  "amount": 100.00,                 // Monto
  "currency": "USD",                // Moneda
  "status": "completed",            // pending, completed, failed, refunded
  "payment_method": "credit_card",  // Método de pago
  "payment_gateway": "stripe",      // Gateway (opcional)
  "transaction_id": "TXN_123",      // ID de transacción
  "metadata": {                     // Información adicional flexible
    "plan_nombre": "Premium",
    "duracion_dias": 30
  },
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:05:00Z",
  "processed_at": "2025-01-15T10:05:00Z"
}
```

## Endpoints

- `POST /payments` - Crear pago
- `GET /payments/:id` - Obtener pago
- `GET /payments/user/:user_id` - Pagos de un usuario
- `GET /payments/entity?entity_type=X&entity_id=Y` - Pagos de una entidad
- `GET /payments/status?status=pending` - Pagos por estado
- `PATCH /payments/:id/status` - Actualizar estado
- `POST /payments/:id/process` - Procesar pago (simulado)
- `GET /healthz` - Health check

## Uso en Gimnasio

```bash
# Crear pago por suscripción
curl -X POST http://localhost:8083/payments \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "subscription",
    "entity_id": "507f1f77bcf86cd799439011",
    "user_id": "5",
    "amount": 100.00,
    "currency": "USD",
    "payment_method": "credit_card",
    "metadata": {
      "plan_nombre": "Plan Premium",
      "duracion_dias": 30
    }
  }'
```

## Uso en E-Commerce

```bash
# Crear pago por orden
curl -X POST http://localhost:8083/payments \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "order",
    "entity_id": "order_12345",
    "user_id": "456",
    "amount": 250.00,
    "currency": "USD",
    "payment_method": "paypal",
    "metadata": {
      "order_items": ["producto1", "producto2"],
      "shipping_address": "Calle 123"
    }
  }'
```

## Ejecución

```bash
# Local
go mod download
go run cmd/api/main.go

# Docker
docker build -t payments-api .
docker run -p 8083:8083 --env-file .env payments-api
```

## Integración con Payment Gateways

Este microservicio está preparado para integrarse con:
- **Stripe**
- **MercadoPago**
- **PayPal**
- O cualquier otro gateway

Actualmente incluye un procesamiento simulado. Para producción, implementar la integración en `internal/services/payment_service.go:ProcessPayment()`.
