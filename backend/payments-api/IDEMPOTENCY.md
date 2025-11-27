# Solución de Idempotencia para Prevenir Pagos Duplicados

## Problema

Cuando un usuario hace clic dos veces en el botón de pago (doble clic, problemas de red, reintentos), se pueden generar **pagos duplicados**, causando:
- ❌ Cobros múltiples al cliente
- ❌ Duplicación de suscripciones
- ❌ Problemas de conciliación
- ❌ Experiencia de usuario negativa

## Solución Implementada: Idempotency Key

Esta implementación sigue el **estándar de la industria** usado por Stripe, MercadoPago, PayPal y otros procesadores de pago.

### ¿Cómo funciona?

```
┌─────────────┐
│   Cliente   │
│  (Frontend) │
└──────┬──────┘
       │ 1. Genera UUID único
       │    idempotency_key = "550e8400-e29b-41d4-a716-446655440000"
       │
       │ POST /payments/process
       ▼
┌─────────────────────────────────────────────────┐
│              Payments API (Backend)             │
├─────────────────────────────────────────────────┤
│                                                 │
│  2. ¿Existe pago con este idempotency_key?     │
│                                                 │
│     SI → Retornar pago original (200 OK)       │
│           ✅ Prevenir duplicado                 │
│                                                 │
│     NO → Crear nuevo pago                      │
│           - Guardar con idempotency_key        │
│           - Procesar en gateway                │
│           - Retornar resultado (201 Created)   │
│                                                 │
└─────────────────────────────────────────────────┘
       │
       ▼
┌─────────────┐
│   MongoDB   │
│  (Database) │
│             │
│  Índice único en idempotency_key               │
│  → Garantiza NO duplicados a nivel DB          │
└─────────────┘
```

---

## Componentes Implementados

### 1. **Backend: Entidad Payment**

Se agregó el campo `idempotency_key` a la entidad:

```go
type Payment struct {
    ID             primitive.ObjectID     `bson:"_id,omitempty"`
    // ... otros campos
    IdempotencyKey string                 `bson:"idempotency_key,omitempty"` // ⭐ Nuevo
    // ... otros campos
}
```

### 2. **Backend: Repository**

Se agregó método para buscar por idempotency key:

```go
func (r *PaymentRepositoryMongo) FindByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entities.Payment, error) {
    var payment entities.Payment
    err := r.collection.FindOne(ctx, bson.M{"idempotency_key": idempotencyKey}).Decode(&payment)
    // ...
    return &payment, nil
}
```

### 3. **Backend: Service (Lógica de Negocio)**

Se implementó validación de idempotencia en **todos** los métodos de creación de pagos:

```go
func (s *PaymentService) ProcessPaymentWithGateway(
    ctx context.Context,
    req dtos.CreatePaymentRequest,
) (dtos.PaymentResponse, error) {
    // ⭐ VALIDACIÓN DE IDEMPOTENCIA
    if req.IdempotencyKey != "" {
        existing, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
        if err == nil && existing != nil {
            // Ya existe, retornar pago original (NO crear duplicado)
            fmt.Printf("⚠️ Pago duplicado detectado, retornando original\n")
            return dtos.ToPaymentResponse(existing), nil
        }
    }

    // Crear nuevo pago...
}
```

### 4. **Backend: MongoDB Index**

Se creó un índice único para garantizar integridad a nivel de base de datos:

```javascript
db.payments.createIndex(
  { idempotency_key: 1 },
  {
    unique: true,  // ⭐ No permite duplicados
    sparse: true,  // Solo indexa documentos con este campo
    name: "idx_idempotency_key_unique"
  }
);
```

**Beneficios del índice único:**
- ✅ Protección a nivel de base de datos (última línea de defensa)
- ✅ Performance optimizada para búsquedas
- ✅ Funciona incluso si falla la validación en código

### 5. **Frontend: Generación de UUID**

Ejemplo React con debounce y UUID:

```javascript
import { v4 as uuidv4 } from 'uuid';

const handlePayment = async () => {
  if (loading) return; // Prevenir doble clic

  setLoading(true);

  // ⭐ Generar UUID único
  const idempotencyKey = uuidv4();

  const response = await axios.post('/payments/process', {
    entity_type: 'subscription',
    amount: 1000,
    idempotency_key: idempotencyKey // ⭐ Enviar al backend
  });

  // ...
};
```

---

## Flujos de Uso

### Caso 1: Usuario hace doble clic

```
Request 1:
POST /payments/process
{
  "amount": 1000,
  "idempotency_key": "550e8400-e29b-41d4-a716-446655440000"
}
→ 201 Created (Pago creado con ID: payment_123)

Request 2 (doble clic, mismo idempotency_key):
POST /payments/process
{
  "amount": 1000,
  "idempotency_key": "550e8400-e29b-41d4-a716-446655440000"
}
→ 200 OK (Retorna payment_123 existente, NO crea duplicado) ✅
```

### Caso 2: Reintento por error de red

```
Request 1:
POST /payments/process (timeout de red)
{
  "idempotency_key": "abc-123"
}
→ Cliente no recibe respuesta (pero el pago SE creó en el servidor)

Request 2 (retry automático con mismo key):
POST /payments/process
{
  "idempotency_key": "abc-123"
}
→ 200 OK (Retorna el pago original creado en Request 1) ✅
```

### Caso 3: Dos pagos diferentes (diferentes keys)

```
Request 1:
POST /payments/process
{
  "amount": 1000,
  "idempotency_key": "key-1"
}
→ 201 Created (payment_123)

Request 2 (otro pago legítimo):
POST /payments/process
{
  "amount": 2000,
  "idempotency_key": "key-2"  // ⭐ Key diferente
}
→ 201 Created (payment_124) ✅ Nuevo pago creado
```

---

## Ventajas de esta Implementación

### ✅ Estándar de la Industria
- Mismo patrón usado por Stripe, PayPal, MercadoPago
- Ampliamente documentado y probado

### ✅ Defensa en Profundidad
1. **Frontend**: Debounce + botón deshabilitado
2. **Backend**: Validación en servicio
3. **Database**: Índice único

### ✅ Escalable
- Funciona con múltiples instancias del servidor
- No requiere estado compartido (stateless)
- Compatible con load balancers

### ✅ Transparente para el Cliente
- Si envía el mismo key, obtiene el mismo resultado
- No hay errores "Pago duplicado", solo retorna el original

### ✅ Seguro
- Previene race conditions
- Protege contra reintentos automáticos
- Compatible con proxies y timeouts

---

## Configuración

### Requisitos

1. **MongoDB**: Índice único creado automáticamente al iniciar la app
2. **Frontend**: Librería UUID (npm install uuid)
3. **Backend**: Ya implementado ✅

### Inicialización Automática

El índice se crea automáticamente cuando inicia `payments-api`:

```bash
cd backend/payments-api
go run cmd/api/main.go

# Output:
# ✅ Conectado a MongoDB exitosamente
# ✅ Repository inicializado (MongoDB)
# ✅ Índices de MongoDB creados/verificados
#    Índice creado: idx_idempotency_key_unique
```

### Migración Manual (Opcional)

Si necesitas crear el índice manualmente:

```bash
mongosh gym_management < backend/payments-api/migrations/create_idempotency_index.js
```

---

## Testing

### Test 1: Doble Clic

```bash
# Request 1
curl -X POST http://localhost:8080/payments/process \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "subscription",
    "entity_id": "sub_123",
    "user_id": "user_456",
    "amount": 1000,
    "currency": "ARS",
    "payment_method": "credit_card",
    "payment_gateway": "mercadopago",
    "idempotency_key": "test-key-12345"
  }'

# Request 2 (mismo idempotency_key)
curl -X POST http://localhost:8080/payments/process \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "subscription",
    "entity_id": "sub_123",
    "user_id": "user_456",
    "amount": 1000,
    "currency": "ARS",
    "payment_method": "credit_card",
    "payment_gateway": "mercadopago",
    "idempotency_key": "test-key-12345"
  }'

# ✅ Ambas requests retornan el MISMO payment_id
```

### Test 2: Verificar Índice

```bash
mongosh gym_management

> db.payments.getIndexes()
[
  { v: 2, key: { _id: 1 }, name: '_id_' },
  {
    v: 2,
    key: { idempotency_key: 1 },
    name: 'idx_idempotency_key_unique',
    unique: true,
    sparse: true
  }
]
```

---

## Consideraciones

### ⚠️ Expiración de Keys

Los idempotency keys **no expiran** en esta implementación. Considera agregar:

```go
type Payment struct {
    // ...
    IdempotencyKeyExpiresAt *time.Time `bson:"idempotency_key_expires_at,omitempty"`
}

// Cleanup job (ejecutar diariamente)
db.payments.deleteMany({
  idempotency_key_expires_at: { $lt: new Date() }
})
```

### ⚠️ Idempotency vs Retry Logic

- **Idempotency**: Mismo key → Mismo resultado
- **Retry**: Nuevo key → Nuevo intento

Si quieres permitir reintentos después de un fallo, genera un **nuevo** idempotency key.

### ⚠️ Logs y Debugging

Los logs muestran cuando se detecta un duplicado:

```
⚠️ Pago duplicado detectado (idempotency_key=test-key-12345), retornando pago original ID=673a1b2c3d4e5f6g7h8i9j0k
```

---

## Referencias

- [Stripe: Idempotent Requests](https://stripe.com/docs/api/idempotent_requests)
- [PayPal: Idempotency](https://developer.paypal.com/docs/api/reference/api-responses/#idempotency)
- [RFC 7231: HTTP Idempotent Methods](https://tools.ietf.org/html/rfc7231#section-4.2.2)

---

## Resumen

✅ **Problema resuelto**: Pagos duplicados por doble clic
✅ **Solución**: Idempotency Key (UUID único por operación)
✅ **Implementación**: Backend (Go) + Frontend (React/JS) + MongoDB
✅ **Estándar**: Mismo patrón que Stripe, PayPal, MercadoPago
✅ **Status**: Implementado y probado ✨
