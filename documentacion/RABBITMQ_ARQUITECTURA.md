# Arquitectura RabbitMQ - Gym Management System

## Exchange

| Nombre | Tipo | Descripción |
|--------|------|-------------|
| `gym.events` | topic | Exchange principal para todos los eventos del sistema |

---

## Publishers (Productores de eventos)

### activities-api

| Routing Key | Cuándo se dispara | Datos principales |
|-------------|-------------------|-------------------|
| `activity.create` | Se crea una actividad | id, titulo, categoria, cupo, instructor |
| `activity.update` | Se edita una actividad | id, campos actualizados |
| `activity.delete` | Se elimina una actividad | id |
| `inscription.create` | Usuario se inscribe a actividad | inscripcion_id, usuario_id, actividad_id |
| `inscription.delete` | Usuario cancela inscripción | inscripcion_id, usuario_id, actividad_id |

### subscriptions-api

| Routing Key | Cuándo se dispara | Datos principales |
|-------------|-------------------|-------------------|
| `subscription.create` | Nueva suscripción creada | subscription_id, usuario_id, plan_id |
| `subscription.activated` | Suscripción activada (pago confirmado) | subscription_id, usuario_id |
| `subscription.cancelled` | Suscripción cancelada | subscription_id, usuario_id |
| `subscription.expired` | Suscripción expirada | subscription_id, usuario_id |
| `subscription.payment_failed` | Pago de suscripción falló | subscription_id, usuario_id |
| `subscription.cancelled_by_refund` | Cancelada por reembolso | subscription_id, usuario_id |

### payments-api

| Routing Key | Cuándo se dispara | Datos principales |
|-------------|-------------------|-------------------|
| `payment.created` | Pago creado (pendiente) | payment_id, subscription_id, amount |
| `payment.completed` | Pago completado exitosamente | payment_id, subscription_id |
| `payment.failed` | Pago falló | payment_id, subscription_id, error |
| `payment.refunded` | Pago reembolsado | payment_id, subscription_id, refund_amount |

---

## Queues y Consumers (Consumidores de eventos)

### search_indexer_queue

| Propiedad | Valor |
|-----------|-------|
| **Microservicio** | search-api |
| **Bindings** | `activity.*`, `inscription.*` |
| **Propósito** | Indexar actividades en Solr y actualizar cupo disponible |

**Acciones:**
- `activity.create` → Indexa nueva actividad en Solr
- `activity.update` → Actualiza documento en Solr
- `activity.delete` → Elimina documento de Solr
- `inscription.create` → Reindexa actividad (actualiza cupo disponible)
- `inscription.delete` → Reindexa actividad (actualiza cupo disponible)

### activities_subscription_events

| Propiedad | Valor |
|-----------|-------|
| **Microservicio** | activities-api |
| **Bindings** | `subscription.cancelled` |
| **Propósito** | Desinscribir usuarios de actividades cuando se cancela su suscripción |

**Acciones:**
- `subscription.cancelled` → Elimina todas las inscripciones del usuario

### subscriptions_payment_queue

| Propiedad | Valor |
|-----------|-------|
| **Microservicio** | subscriptions-api |
| **Bindings** | `payment.*` |
| **Propósito** | Actualizar estado de suscripción según resultado del pago |

**Acciones:**
- `payment.completed` → Activa la suscripción
- `payment.failed` → Marca suscripción como pago fallido
- `payment.refunded` → Cancela la suscripción

---

## Diagrama de Flujo

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                  PUBLISHERS                                      │
├─────────────────────┬─────────────────────────┬─────────────────────────────────┤
│   activities-api    │   subscriptions-api     │         payments-api            │
│                     │                         │                                 │
│ • activity.create   │ • subscription.create   │ • payment.created               │
│ • activity.update   │ • subscription.activated│ • payment.completed             │
│ • activity.delete   │ • subscription.cancelled│ • payment.failed                │
│ • inscription.create│ • subscription.expired  │ • payment.refunded              │
│ • inscription.delete│ • subscription.*_failed │                                 │
└──────────┬──────────┴───────────┬─────────────┴────────────────┬────────────────┘
           │                      │                              │
           ▼                      ▼                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        EXCHANGE: gym.events (topic)                             │
└─────────────────────────────────────────────────────────────────────────────────┘
           │                      │                              │
           │ activity.*           │ subscription.cancelled       │ payment.*
           │ inscription.*        │                              │
           ▼                      ▼                              ▼
┌─────────────────────┬─────────────────────────┬─────────────────────────────────┐
│ search_indexer_     │ activities_subscription_│ subscriptions_payment_          │
│ queue               │ events                  │ queue                           │
├─────────────────────┼─────────────────────────┼─────────────────────────────────┤
│   search-api        │   activities-api        │   subscriptions-api             │
│                     │                         │                                 │
│ Indexa en Solr      │ Desinscribe usuarios    │ Activa/cancela suscripción      │
│ Actualiza cupo      │ sin suscripción activa  │ según resultado del pago        │
└─────────────────────┴─────────────────────────┴─────────────────────────────────┘
                                  CONSUMERS
```

---

## Manejo de Errores

### Acknowledgment de mensajes

| Método | Uso | Efecto |
|--------|-----|--------|
| `msg.Ack(false)` | Mensaje procesado OK | Elimina mensaje de la cola |
| `msg.Nack(false, true)` | Error temporal (Solr caído, etc.) | Re-encola para reintentar |
| `msg.Nack(false, false)` | Error permanente (JSON inválido) | Descarta mensaje |

### Fallbacks

| Microservicio | Si RabbitMQ no está disponible |
|---------------|--------------------------------|
| activities-api | Usa `NullEventPublisher` (no publica) |
| subscriptions-api | Usa `NullEventPublisher` (no publica) |
| payments-api | Usa `NoOpEventPublisher` (no publica) |
| search-api | Usa fallback a MySQL para búsquedas |

---

## Configuración

### Variables de entorno

```env
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
RABBITMQ_EXCHANGE=gym.events
```

### Docker Compose

```yaml
rabbitmq:
  image: rabbitmq:3-management
  ports:
    - "5672:5672"    # AMQP
    - "15672:15672"  # Management UI
  environment:
    RABBITMQ_DEFAULT_USER: guest
    RABBITMQ_DEFAULT_PASS: guest
```

---

## Comandos útiles

### Ver colas y bindings

```bash
# Listar colas
curl -s -u guest:guest "http://localhost:15672/api/queues" | python -m json.tool

# Ver bindings de una cola específica
curl -s -u guest:guest "http://localhost:15672/api/queues/%2F/search_indexer_queue/bindings"

# Ver mensajes en cola (sin consumir)
curl -s -u guest:guest "http://localhost:15672/api/queues/%2F/search_indexer_queue"
```

### Eliminar binding manualmente

```bash
curl -X DELETE -u guest:guest \
  "http://localhost:15672/api/bindings/%2F/e/gym.events/q/QUEUE_NAME/ROUTING_KEY"
```

---

## Flujos de negocio completos

### Usuario se inscribe a actividad

```
1. Frontend → POST /inscripciones
2. activities-api guarda en MySQL
3. activities-api publica inscription.create
4. search-api consume → reindexa actividad en Solr (cupo -1)
```

### Usuario paga suscripción

```
1. Frontend → POST /payments
2. payments-api crea pago (pending)
3. payments-api publica payment.created
4. Admin confirma pago → payment.completed
5. subscriptions-api consume → activa suscripción
```

### Usuario cancela suscripción

```
1. Frontend → DELETE /subscriptions/:id
2. subscriptions-api cancela en MongoDB
3. subscriptions-api publica subscription.cancelled
4. activities-api consume → elimina inscripciones del usuario
```
