# 🤝 Contrato de Integración: Payments API ↔ Subscriptions API

## 📋 Índice
1. [Contexto y Reglas de Negocio](#contexto)
2. [Arquitectura de Integración](#arquitectura)
3. [Endpoints que PAYMENTS-API expone](#endpoints-payments-api)
4. [Endpoints que SUBSCRIPTIONS-API debe exponer](#endpoints-subscriptions-api)
5. [Eventos RabbitMQ](#eventos-rabbitmq)
6. [Flujos Completos](#flujos-completos)
7. [DTOs y Contratos](#dtos)

---

## 🎯 Contexto y Reglas de Negocio {#contexto}

### Planes de Suscripción
- **Plan Básico**: Solo acceso a categoría "musculación" - $2,500 ARS/mes
- **Plan Premium**: Acceso a todas las categorías - $5,000 ARS/mes

### Reglas
1. Las suscripciones duran **30 días** desde la fecha de activación
2. Al vencerse, el usuario **NO puede acceder** a actividades hasta que pague
3. El pago se procesa mediante **MercadoPago** o **Mock** (testing)
4. Cuando un pago se completa, la suscripción se **activa/renueva automáticamente**

---

## 🏗️ Arquitectura de Integración {#arquitectura}

```
┌─────────────────────┐         ┌─────────────────────┐
│  SUBSCRIPTIONS-API  │◄───────►│    PAYMENTS-API     │
│   (Puerto 8082)     │  REST   │   (Puerto 8083)     │
└──────────┬──────────┘         └──────────┬──────────┘
           │                               │
           │         ┌──────────────┐      │
           └────────►│  RabbitMQ    │◄─────┘
                     │ (gym_events) │
                     └──────────────┘

Flujo:
1. Usuario solicita suscripción → SUBSCRIPTIONS-API
2. SUBSCRIPTIONS-API crea suscripción con estado "pending_payment"
3. SUBSCRIPTIONS-API llama a PAYMENTS-API para crear preferencia de pago
4. Usuario paga en MercadoPago
5. PAYMENTS-API recibe webhook y publica evento "payment.completed"
6. SUBSCRIPTIONS-API escucha el evento y activa la suscripción
```

---

## 📤 Endpoints que PAYMENTS-API expone (YA IMPLEMENTADOS) {#endpoints-payments-api}

### 1. Crear Pago para Suscripción

```http
POST http://localhost:8083/payments/process
Content-Type: application/json

{
  "entity_type": "subscription",
  "entity_id": "sub_12345",          // ← ID de la suscripción
  "user_id": "user_789",
  "amount": 5000.00,                  // ← Monto del plan
  "currency": "ARS",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "callback_url": "https://tugimnasio.com/subscription/success",
  "webhook_url": "http://localhost:8083/webhooks/mercadopago",
  "metadata": {
    "customer_email": "juan@email.com",
    "customer_name": "Juan Pérez",
    "plan_name": "Plan Premium",
    "plan_duration_days": 30
  }
}
```

**Respuesta (200 OK):**
```json
{
  "id": "payment_abc123",
  "entity_type": "subscription",
  "entity_id": "sub_12345",
  "user_id": "user_789",
  "amount": 5000,
  "currency": "ARS",
  "status": "pending",               // ← Estado inicial
  "payment_gateway": "mercadopago",
  "transaction_id": "MP-123456",
  "metadata": {
    "payment_url": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=...",
    "gateway_message": "Preferencia creada. Redirigir al usuario a payment_url"
  },
  "created_at": "2025-11-01T10:00:00Z"
}
```

### 2. Consultar Estado de Pago

```http
GET http://localhost:8083/payments/entity?entity_type=subscription&entity_id=sub_12345
```

**Respuesta (200 OK):**
```json
[
  {
    "id": "payment_abc123",
    "entity_type": "subscription",
    "entity_id": "sub_12345",
    "user_id": "user_789",
    "amount": 5000,
    "status": "completed",           // ← ESTADO ACTUALIZADO
    "payment_gateway": "mercadopago",
    "transaction_id": "MP-123456",
    "processed_at": "2025-11-01T10:05:00Z",
    "created_at": "2025-11-01T10:00:00Z"
  }
]
```

### 3. Obtener Pagos de un Usuario

```http
GET http://localhost:8083/payments/user/user_789
```

**Útil para:** Historial de pagos del usuario en su perfil.

---

## 📥 Endpoints que SUBSCRIPTIONS-API debe exponer {#endpoints-subscriptions-api}

### 1. Obtener Suscripción Activa de un Usuario

**Para que PAYMENTS-API pueda validar antes de crear un pago**

```http
GET http://localhost:8082/subscriptions/user/{user_id}/active
```

**Respuesta (200 OK):**
```json
{
  "id": "sub_12345",
  "user_id": "user_789",
  "plan_type": "premium",             // "basic" o "premium"
  "status": "active",                 // "pending_payment", "active", "expired"
  "amount": 5000.00,
  "currency": "ARS",
  "start_date": "2025-11-01T00:00:00Z",
  "end_date": "2025-12-01T00:00:00Z",
  "auto_renew": true
}
```

**Respuesta (404 Not Found):** Si no tiene suscripción activa
```json
{
  "error": "No active subscription found for user"
}
```

### 2. Validar Acceso a Categoría

**Para que ACTIVITIES-API valide si el usuario puede inscribirse**

```http
GET http://localhost:8082/subscriptions/user/{user_id}/can-access-category?category={category_name}
```

**Ejemplos:**
```http
GET http://localhost:8082/subscriptions/user/user_789/can-access-category?category=musculacion
GET http://localhost:8082/subscriptions/user/user_789/can-access-category?category=yoga
```

**Respuesta (200 OK):**
```json
{
  "user_id": "user_789",
  "category": "yoga",
  "has_access": true,                 // ← TRUE si puede acceder
  "subscription": {
    "plan_type": "premium",
    "status": "active",
    "end_date": "2025-12-01T00:00:00Z"
  }
}
```

**Respuesta (200 OK) - Sin acceso:**
```json
{
  "user_id": "user_789",
  "category": "yoga",
  "has_access": false,                // ← FALSE - Plan básico no permite
  "reason": "Plan básico solo permite categoría musculación",
  "subscription": {
    "plan_type": "basic",
    "status": "active"
  }
}
```

### 3. Crear/Renovar Suscripción

**Llamado desde el frontend cuando el usuario elige un plan**

```http
POST http://localhost:8082/subscriptions
Content-Type: application/json

{
  "user_id": "user_789",
  "plan_type": "premium",             // "basic" o "premium"
  "auto_renew": true
}
```

**Respuesta (201 Created):**
```json
{
  "id": "sub_12345",
  "user_id": "user_789",
  "plan_type": "premium",
  "status": "pending_payment",        // ← Estado inicial
  "amount": 5000.00,
  "currency": "ARS",
  "payment": {
    "payment_id": "payment_abc123",
    "payment_url": "https://www.mercadopago.com.ar/checkout/...",
    "status": "pending"
  },
  "created_at": "2025-11-01T10:00:00Z"
}
```

---

## 🔔 Eventos RabbitMQ {#eventos-rabbitmq}

### Exchange: `gym_events` (ya configurado en tu código)

### Eventos que PAYMENTS-API publica:

#### 1. `payment.created`
```json
{
  "event_type": "payment.created",
  "timestamp": "2025-11-01T10:00:00Z",
  "data": {
    "payment_id": "payment_abc123",
    "entity_type": "subscription",
    "entity_id": "sub_12345",
    "user_id": "user_789",
    "amount": 5000.00,
    "currency": "ARS",
    "status": "pending",
    "payment_gateway": "mercadopago"
  }
}
```

#### 2. `payment.completed` ⭐ **IMPORTANTE**
```json
{
  "event_type": "payment.completed",
  "timestamp": "2025-11-01T10:05:00Z",
  "data": {
    "payment_id": "payment_abc123",
    "entity_type": "subscription",
    "entity_id": "sub_12345",        // ← SUBSCRIPTIONS-API usa este ID
    "user_id": "user_789",
    "amount": 5000.00,
    "currency": "ARS",
    "status": "completed",
    "transaction_id": "MP-123456",
    "processed_at": "2025-11-01T10:05:00Z"
  }
}
```

**Acción de SUBSCRIPTIONS-API:**
1. Escuchar evento `payment.completed`
2. Si `entity_type == "subscription"` → activar suscripción
3. Actualizar `status` de `pending_payment` → `active`
4. Setear `start_date` = ahora, `end_date` = ahora + 30 días

#### 3. `payment.failed`
```json
{
  "event_type": "payment.failed",
  "timestamp": "2025-11-01T10:05:00Z",
  "data": {
    "payment_id": "payment_abc123",
    "entity_type": "subscription",
    "entity_id": "sub_12345",
    "user_id": "user_789",
    "status": "failed",
    "reason": "Tarjeta rechazada"
  }
}
```

---

## 📊 Flujos Completos {#flujos-completos}

### Flujo 1: Usuario se suscribe por primera vez

```
┌─────────┐      ┌──────────────┐      ┌─────────────┐      ┌──────────┐
│ Usuario │      │ Subscriptions│      │  Payments   │      │MercadoPago│
└────┬────┘      └──────┬───────┘      └──────┬──────┘      └─────┬────┘
     │                  │                      │                   │
     │ 1. POST /subscriptions (plan: premium) │                   │
     ├─────────────────►│                      │                   │
     │                  │                      │                   │
     │                  │ 2. POST /payments/process                │
     │                  ├─────────────────────►│                   │
     │                  │                      │                   │
     │                  │                      │ 3. Create preference
     │                  │                      ├──────────────────►│
     │                  │                      │                   │
     │                  │                      │ 4. Preference ID  │
     │                  │                      │◄──────────────────┤
     │                  │                      │                   │
     │                  │ 5. {payment_url, status: pending}        │
     │                  │◄─────────────────────┤                   │
     │                  │                      │                   │
     │ 6. {subscription, payment_url}          │                   │
     │◄─────────────────┤                      │                   │
     │                  │                      │                   │
     │ 7. Usuario redirigido a MercadoPago     │                   │
     ├─────────────────────────────────────────────────────────────►│
     │                  │                      │                   │
     │ 8. Usuario paga con tarjeta             │                   │
     │◄────────────────────────────────────────────────────────────┤
     │                  │                      │                   │
     │                  │                      │ 9. Webhook: payment completed
     │                  │                      │◄──────────────────┤
     │                  │                      │                   │
     │                  │                      │ 10. Publica evento:
     │                  │                      │     payment.completed
     │                  │                      ├────┐              │
     │                  │                      │    │              │
     │                  │                      │◄───┘              │
     │                  │                      │                   │
     │                  │ 11. Escucha evento   │                   │
     │                  │◄─────────────────────┤                   │
     │                  │                      │                   │
     │                  │ 12. Activa suscripción                   │
     │                  │      status: active  │                   │
     │                  │      end_date: +30d  │                   │
     │                  ├────┐                 │                   │
     │                  │    │                 │                   │
     │                  │◄───┘                 │                   │
     │                  │                      │                   │
     │ 13. Email: "Suscripción activada"       │                   │
     │◄─────────────────┤                      │                   │
     │                  │                      │                   │
```

### Flujo 2: Usuario intenta acceder a una actividad

```
┌─────────┐      ┌──────────────┐      ┌─────────────┐
│ Usuario │      │ Activities   │      │Subscriptions│
└────┬────┘      └──────┬───────┘      └──────┬──────┘
     │                  │                      │
     │ POST /activities/123/register           │
     ├─────────────────►│                      │
     │                  │                      │
     │                  │ GET /subscriptions/user/789/can-access-category?category=yoga
     │                  ├─────────────────────►│
     │                  │                      │
     │                  │ {has_access: true}   │
     │                  │◄─────────────────────┤
     │                  │                      │
     │ {success: true}  │                      │
     │◄─────────────────┤                      │
     │                  │                      │
```

**Si NO tiene acceso:**
```json
{
  "error": "No tienes acceso a esta categoría",
  "details": {
    "required_plan": "premium",
    "current_plan": "basic",
    "upgrade_url": "/subscriptions/upgrade"
  }
}
```

### Flujo 3: Suscripción vence y se renueva

```
1. Cron job en SUBSCRIPTIONS-API corre diariamente
2. Detecta suscripciones con end_date < hoy
3. Si auto_renew == true:
   → Llama a PAYMENTS-API para crear nuevo pago
   → Usuario recibe email con link de pago
4. Si auto_renew == false:
   → Marca suscripción como "expired"
   → Usuario pierde acceso
```

---

## 📦 DTOs y Contratos {#dtos}

### DTO para crear suscripción (SUBSCRIPTIONS-API)

```go
type CreateSubscriptionRequest struct {
    UserID    string `json:"user_id" binding:"required"`
    PlanType  string `json:"plan_type" binding:"required,oneof=basic premium"`
    AutoRenew bool   `json:"auto_renew"`
}
```

### DTO de respuesta de suscripción

```go
type SubscriptionResponse struct {
    ID        string                 `json:"id"`
    UserID    string                 `json:"user_id"`
    PlanType  string                 `json:"plan_type"`
    Status    string                 `json:"status"` // pending_payment, active, expired
    Amount    float64                `json:"amount"`
    Currency  string                 `json:"currency"`
    StartDate *time.Time             `json:"start_date,omitempty"`
    EndDate   *time.Time             `json:"end_date,omitempty"`
    AutoRenew bool                   `json:"auto_renew"`
    Payment   *PaymentInfo           `json:"payment,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
}

type PaymentInfo struct {
    PaymentID  string `json:"payment_id"`
    PaymentURL string `json:"payment_url"`
    Status     string `json:"status"`
}
```

### Evento de RabbitMQ (ya implementado en PAYMENTS-API)

```go
type PaymentEvent struct {
    EventType string                 `json:"event_type"` // "payment.completed", etc.
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}
```

---

## 🔐 Validaciones Importantes

### En SUBSCRIPTIONS-API:
1. **No permitir crear suscripción** si ya tiene una activa
2. **Validar plan_type** sea "basic" o "premium"
3. **Al activar suscripción**, verificar que el pago esté realmente completed
4. **Validar acceso**:
   - Basic → solo "musculacion"
   - Premium → todas las categorías
   - Expired → ninguna categoría

### En PAYMENTS-API (ya implementado):
1. Validar que `entity_type == "subscription"`
2. Guardar `entity_id` correctamente para linking
3. Publicar eventos solo cuando cambia el estado del pago

---

## 🎯 Checklist para el desarrollador de SUBSCRIPTIONS-API

- [ ] Crear endpoint `GET /subscriptions/user/{user_id}/active`
- [ ] Crear endpoint `GET /subscriptions/user/{user_id}/can-access-category`
- [ ] Crear endpoint `POST /subscriptions`
- [ ] Implementar consumer de RabbitMQ para `payment.completed`
- [ ] Al recibir `payment.completed`, activar suscripción (status: active)
- [ ] Setear `start_date` y `end_date` (end_date = start_date + 30 días)
- [ ] Implementar lógica de validación de acceso por plan
- [ ] (Opcional) Cron job para detectar suscripciones vencidas
- [ ] (Opcional) Endpoint para renovar suscripción manualmente

---

## 📞 Contacto y Coordinación

**PAYMENTS-API está listo y esperando:**
- Puerto: `8083`
- RabbitMQ: Publicando eventos en exchange `gym_events`
- Endpoints documentados arriba funcionando

**Siguiente paso:** El equipo de SUBSCRIPTIONS-API debe implementar los endpoints y consumidores descritos.

---

## 📚 Documentación Adicional

- Ver archivo: `GUIA_RABBITMQ_EVENTOS.md` para detalles de eventos
- Ver archivo: `Postman_Collection_Payments_API.json` para probar endpoints
- MercadoPago Docs: https://www.mercadopago.com.ar/developers/es/docs
