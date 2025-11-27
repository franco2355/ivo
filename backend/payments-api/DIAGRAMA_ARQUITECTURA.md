# Diagrama de Arquitectura - Sistema de Gimnasio

## 1. Vista General del Sistema

```mermaid
graph TB
    subgraph "Cliente Web/Mobile"
        CLIENT[Usuario del Gimnasio]
    end

    subgraph "Microservicios"
        SUBS[Subscriptions API<br/>:8081]
        PAYMENTS[Payments API<br/>:8083]
        ACTIVITIES[Activities API<br/>:8082]
    end

    subgraph "Bases de Datos"
        MONGO_PAYMENTS[(MongoDB<br/>Payments<br/>:27017)]
        MONGO_SUBS[(MongoDB<br/>Subscriptions)]
        DB_ACTIVITIES[(DB Activities)]
    end

    subgraph "Servicios Externos"
        MERCADOPAGO[MercadoPago API<br/>Gateway de Pagos]
        RABBITMQ[RabbitMQ<br/>Exchange: gym_events]
    end

    CLIENT -->|HTTP REST| SUBS
    CLIENT -->|HTTP REST| ACTIVITIES
    SUBS -->|HTTP REST| PAYMENTS
    SUBS -->|Consume eventos| RABBITMQ
    ACTIVITIES -->|HTTP REST| SUBS

    PAYMENTS -->|Publica eventos| RABBITMQ
    PAYMENTS -->|HTTP REST| MERCADOPAGO
    PAYMENTS <-->|Read/Write| MONGO_PAYMENTS
    SUBS <-->|Read/Write| MONGO_SUBS
    ACTIVITIES <-->|Read/Write| DB_ACTIVITIES

    style PAYMENTS fill:#4CAF50,stroke:#2E7D32,stroke-width:3px,color:#fff
    style RABBITMQ fill:#FF6F00,stroke:#E65100,stroke-width:2px,color:#fff
    style MERCADOPAGO fill:#009EE3,stroke:#0277BD,stroke-width:2px,color:#fff
    style MONGO_PAYMENTS fill:#47A248,stroke:#2E7D32,stroke-width:2px,color:#fff
```

## 2. Flujo de Creación de Suscripción con Pago

```mermaid
sequenceDiagram
    participant C as Cliente
    participant S as Subscriptions API
    participant P as Payments API
    participant MP as MercadoPago
    participant MQ as RabbitMQ
    participant DB as MongoDB

    C->>S: POST /subscriptions/create<br/>{user_id, plan_type}

    S->>S: Crear suscripción<br/>(estado: pending_payment)
    S->>DB: Guardar suscripción

    S->>P: POST /payments/process<br/>{entity_type: subscription,<br/>entity_id, amount, etc.}

    P->>DB: Guardar pago (pending)
    P->>MQ: Publicar evento<br/>payment.created

    P->>MP: Crear Checkout Pro<br/>con back_urls y webhook
    MP-->>P: {payment_url, transaction_id}

    P->>DB: Actualizar pago con transaction_id
    P-->>S: {payment_id, payment_url,<br/>transaction_id, status}

    S->>DB: Actualizar subscription.payment_id
    S-->>C: {subscription_id, payment_url}

    C->>MP: Usuario paga en payment_url

    MP->>P: POST /webhooks/mercadopago<br/>{payment_id, status}

    P->>MP: GET /v1/payments/{id}<br/>(verificar estado real)
    MP-->>P: {status: approved}

    P->>DB: Actualizar pago (completed)
    P->>MQ: Publicar evento<br/>payment.completed

    MQ->>S: Consumir evento<br/>payment.completed
    S->>S: Activar suscripción<br/>(estado: active)
    S->>DB: Actualizar suscripción
```

## 3. Arquitectura Interna de Payments API

```mermaid
graph LR
    subgraph "Capa de Presentación"
        ROUTER[Gin Router]
        CTRL_PAY[Payment Controller]
        CTRL_WH[Webhook Controller]
        MIDDLEWARE[CORS Middleware]
    end

    subgraph "Capa de Negocio"
        SERVICE[Payment Service]
        EVENT_PUB[Event Publisher<br/>RabbitMQ]
    end

    subgraph "Capa de Gateways (Strategy)"
        FACTORY[Gateway Factory]
        GW_MP[MercadoPago Gateway]
        GW_CASH[Cash Gateway]
    end

    subgraph "Capa de Datos"
        REPO[Payment Repository]
        MONGO[(MongoDB)]
    end

    ROUTER --> MIDDLEWARE
    MIDDLEWARE --> CTRL_PAY
    MIDDLEWARE --> CTRL_WH

    CTRL_PAY --> SERVICE
    CTRL_WH --> SERVICE

    SERVICE --> EVENT_PUB
    SERVICE --> FACTORY
    SERVICE --> REPO

    FACTORY -.->|Crea| GW_MP
    FACTORY -.->|Crea| GW_CASH

    GW_MP -->|API REST| MERCADOPAGO[MercadoPago API]

    REPO --> MONGO
    EVENT_PUB --> RABBITMQ[RabbitMQ]

    style SERVICE fill:#4CAF50,stroke:#2E7D32,stroke-width:3px,color:#fff
    style FACTORY fill:#FF9800,stroke:#E65100,stroke-width:2px,color:#fff
    style EVENT_PUB fill:#FF6F00,stroke:#E65100,stroke-width:2px,color:#fff
```

## 4. Eventos de RabbitMQ (Payments API → RabbitMQ)

```mermaid
graph TD
    subgraph "Payments API - Event Publisher"
        SERVICE[Payment Service]
        PUB[RabbitMQ Publisher]
    end

    subgraph "RabbitMQ Exchange: gym_events"
        EXCHANGE[Topic Exchange<br/>gym_events]
    end

    subgraph "Routing Keys"
        RK1[payment.created]
        RK2[payment.completed]
        RK3[payment.failed]
        RK4[payment.refunded]
    end

    subgraph "Consumers (Otros Microservicios)"
        SUBS_CONSUMER[Subscriptions API<br/>Escucha: payment.*]
        OTHER[Otros consumidores<br/>futuros]
    end

    SERVICE -->|1. Crear pago| PUB
    SERVICE -->|2. Pago completado| PUB
    SERVICE -->|3. Pago fallido| PUB
    SERVICE -->|4. Reembolso| PUB

    PUB -->|Publica| EXCHANGE

    EXCHANGE -->|Routing| RK1
    EXCHANGE -->|Routing| RK2
    EXCHANGE -->|Routing| RK3
    EXCHANGE -->|Routing| RK4

    RK1 --> SUBS_CONSUMER
    RK2 --> SUBS_CONSUMER
    RK3 --> SUBS_CONSUMER
    RK4 --> SUBS_CONSUMER

    RK1 -.-> OTHER
    RK2 -.-> OTHER

    style PUB fill:#FF6F00,stroke:#E65100,stroke-width:3px,color:#fff
    style EXCHANGE fill:#FF9800,stroke:#F57C00,stroke-width:2px,color:#fff
    style SUBS_CONSUMER fill:#2196F3,stroke:#1565C0,stroke-width:2px,color:#fff
```

## 5. Patrones de Diseño Implementados

### Strategy Pattern (Gateways)
```
PaymentGateway (Interface)
    ├── MercadoPagoGateway
    └── CashGateway
```

### Factory Pattern
```
GatewayFactory
    ├── GetGateway(name string) → PaymentGateway
    └── GetSupportedGateways() → []string
```

### Repository Pattern
```
PaymentRepository (Interface)
    └── PaymentRepositoryMongo (Implementación)
```

### Dependency Injection
```
main.go
    ├── Repository → Service
    ├── GatewayFactory → Service
    └── EventPublisher → Service
```

## 6. Estructura de Eventos RabbitMQ

### Evento: payment.created
```json
{
  "event_type": "payment.created",
  "timestamp": "2025-01-30T15:30:00Z",
  "payment_id": "67890abc123def456",
  "entity_type": "subscription",
  "entity_id": "sub_gym_001",
  "user_id": "user_456",
  "amount": 2500.00,
  "currency": "ARS",
  "status": "pending",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "transaction_id": "MP-123456789",
  "metadata": {
    "plan_name": "Plan Premium Mensual"
  }
}
```

### Evento: payment.completed
```json
{
  "event_type": "payment.completed",
  "timestamp": "2025-01-30T15:35:00Z",
  "payment_id": "67890abc123def456",
  "entity_type": "subscription",
  "entity_id": "sub_gym_001",
  "user_id": "user_456",
  "amount": 2500.00,
  "currency": "ARS",
  "status": "completed",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "transaction_id": "MP-123456789",
  "processed_at": "2025-01-30T15:35:00Z",
  "metadata": {
    "plan_name": "Plan Premium Mensual"
  }
}
```

### Evento: payment.failed
```json
{
  "event_type": "payment.failed",
  "timestamp": "2025-01-30T15:35:00Z",
  "payment_id": "67890abc123def456",
  "entity_type": "subscription",
  "entity_id": "sub_gym_001",
  "user_id": "user_456",
  "amount": 2500.00,
  "currency": "ARS",
  "status": "failed",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "transaction_id": "MP-123456789",
  "error_message": "Insufficient funds",
  "metadata": {
    "plan_name": "Plan Premium Mensual"
  }
}
```

## 7. Endpoints de Payments API

### Procesamiento de Pagos
- `POST /payments/process` - Crear y procesar pago único (Checkout Pro)
- `POST /payments/recurring` - Crear pago recurrente (Preapprovals)
- `POST /payments/:id/refund` - Procesar reembolso
- `GET /payments/:id/sync` - Sincronizar estado con gateway

### Consultas
- `GET /payments/:id` - Obtener pago por ID
- `GET /payments/user/:user_id` - Pagos de un usuario
- `GET /payments/entity?entity_type=X&entity_id=Y` - Pagos de una entidad

### Webhooks
- `POST /webhooks/mercadopago` - Webhook de MercadoPago
- `POST /webhooks/:gateway` - Webhook genérico

### Health Check
- `GET /healthz` - Estado del servicio

## 8. Tecnologías Utilizadas

| Componente | Tecnología | Puerto |
|------------|------------|--------|
| Payments API | Go + Gin Framework | 8083 |
| Base de Datos | MongoDB 5.0 | 27017 |
| Message Broker | RabbitMQ | 5672 |
| Gateway de Pagos | MercadoPago Checkout Pro | - |
| Contenedores | Docker | - |

## 9. Variables de Entorno Necesarias

```bash
# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=payments_db

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=gym_events

# MercadoPago
MERCADOPAGO_ACCESS_TOKEN=APP_USR-xxxxxxxx
MERCADOPAGO_PUBLIC_KEY=APP_USR-xxxxxxxx

# API
PORT=8083
```

## 10. Flujo de Datos: Verificación de Acceso a Actividades

```mermaid
sequenceDiagram
    participant U as Usuario
    participant A as Activities API
    participant S as Subscriptions API

    U->>A: Intenta inscribirse a actividad<br/>(categoría: Yoga)

    A->>S: GET /subscriptions/user/{user_id}/active

    S->>S: Buscar suscripción activa

    alt Suscripción activa encontrada
        S-->>A: {plan_type: "basic", status: "active",<br/>payment_status: "completed"}

        A->>A: Verificar categoría permitida<br/>Basic: solo musculación<br/>Premium: todas

        alt Categoría permitida
            A-->>U: ✅ Inscripción exitosa
        else Categoría no permitida
            A-->>U: ❌ Error: Plan básico<br/>solo permite musculación
        end
    else Sin suscripción activa
        S-->>A: 404 Not Found
        A-->>U: ❌ Error: Necesita<br/>suscripción activa
    end
```

## 11. Casos de Uso Principales

### Caso 1: Usuario crea suscripción Premium
1. Usuario selecciona Plan Premium ($5000 ARS/mes)
2. Subscriptions API crea suscripción en estado `pending_payment`
3. Subscriptions API llama a Payments API para procesar pago
4. Payments API crea pago en MongoDB y publica evento `payment.created`
5. Payments API genera link de pago de MercadoPago
6. Usuario completa el pago en MercadoPago
7. MercadoPago notifica a Payments API via webhook
8. Payments API actualiza pago a `completed` y publica evento `payment.completed`
9. Subscriptions API consume evento y activa la suscripción
10. Usuario ahora puede inscribirse a todas las actividades

### Caso 2: Usuario con Plan Básico intenta yoga
1. Usuario con Plan Básico intenta inscribirse a clase de Yoga
2. Activities API consulta a Subscriptions API el plan del usuario
3. Subscriptions API retorna plan_type: "basic"
4. Activities API valida que Yoga requiere Premium
5. Rechaza la inscripción con mensaje de error
6. Usuario debe upgradear a Premium para acceder

### Caso 3: Suscripción vencida (30 días)
1. Pasan 30 días desde la activación
2. Subscriptions API marca suscripción como `expired`
3. Usuario intenta inscribirse a actividad
4. Activities API detecta suscripción vencida
5. Bloquea acceso y sugiere renovar suscripción
6. Usuario debe crear nueva suscripción y pagar nuevamente
