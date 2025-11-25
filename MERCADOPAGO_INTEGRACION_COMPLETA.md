# Integración Completa de MercadoPago

## Estado: ✅ COMPLETADO E INTEGRADO

La integración de MercadoPago está **100% funcional** tanto en backend como en frontend.

---

## 🎯 Flujo Completo de Pago

```
Usuario → Selecciona Plan → Checkout → Crea Suscripción → Crea Pago →
→ MercadoPago Checkout Pro → Usuario Paga → Callback → PagoResultado →
→ Sincroniza con Backend → Muestra Resultado → Navega a Mi Suscripción
```

---

## 📁 Archivos Involucrados

### Backend

1. **`backend/payments-api/internal/gateways/mercadopago_gateway.go`**
   - Gateway completo de MercadoPago
   - Crea preferencias de pago
   - Procesa webhooks
   - Sincroniza estados

2. **`.env`** (líneas 91-97)
   ```bash
   MERCADOPAGO_ACCESS_TOKEN=APP_USR-988447898124356-103117-055b2effa25fcae685ce1e460f2d1d1f-2959351968
   MERCADOPAGO_PUBLIC_KEY=APP_USR-ece25914-5b84-4029-9146-5ace6356a8cf
   MERCADOPAGO_WEBHOOK_SECRET=
   ```

3. **`docker-compose.yml`** (líneas 191-194)
   - Pasa credenciales al contenedor payments-api

### Frontend

1. **`frontend/src/pages/Checkout.jsx`**
   - Crea suscripción
   - Crea pago con MercadoPago
   - Redirige al checkout de MP
   - Callback URL: `http://localhost:5173/pago/resultado`

2. **`frontend/src/pages/PagoResultado.jsx`** ✨ NUEVO
   - Recibe parámetros de MercadoPago (status, collection_id, payment_id)
   - Sincroniza estado con backend vía `/payments/{id}/sync`
   - Muestra resultado según estado: success, pending, rejected
   - Navega a próximos pasos

3. **`frontend/src/styles/PagoResultado.css`** ✨ NUEVO
   - Estilos completos con animaciones
   - Diseño responsive
   - Colores según estado del pago

4. **`frontend/src/pages/main.jsx`** (líneas 19, 44)
   - Route agregada: `/pago/resultado`
   - Lazy loading del componente

---

## 🔄 Flujo Técnico Detallado

### 1️⃣ Usuario Selecciona Plan
- Navega a `/planes`
- Selecciona un plan (ej: Premium $3000/mes)
- Click en "Suscribirse"

### 2️⃣ Página de Checkout (`/checkout/:planId`)
```javascript
// frontend/src/pages/Checkout.jsx

// Usuario selecciona método de pago: "Mercado Pago" o "Efectivo"
formData.payment_method = 'mercadopago'

// Al submit:
// 1. Crear suscripción
POST http://localhost:8081/subscriptions
Body: {
  usuario_id: "5",
  plan_id: "premium_plan_123",
  metodo_pago: "mercadopago"
}
Response: { id: "sub_premium_123", estado: "pendiente" }

// 2. Crear pago con MercadoPago
POST http://localhost:8083/payments/process
Body: {
  entity_type: "subscription",
  entity_id: "sub_premium_123",
  user_id: "5",
  amount: 3000,
  currency: "ARS",
  payment_method: "credit_card",
  payment_gateway: "mercadopago",
  callback_url: "http://localhost:5173/pago/resultado",
  webhook_url: "http://localhost:8083/webhooks/mercadopago"
}

Response: {
  id: "payment_abc123",
  transaction_id: "2959351968-6abe2156-8307-4dee-92cb-04b0012beaa2",
  metadata: {
    payment_url: "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=..."
  }
}

// 3. Redireccionar a MercadoPago
window.location.href = paymentResult.metadata.payment_url
```

### 3️⃣ Usuario Paga en MercadoPago
- Usuario ingresa datos de tarjeta
- MercadoPago procesa el pago
- MercadoPago redirige a callback_url con parámetros:

```
http://localhost:5173/pago/resultado?
  collection_id=123456789&
  collection_status=approved&
  payment_id=123456789&
  status=approved&
  external_reference=payment_abc123&
  preference_id=2959351968-6abe2156-8307-4dee-92cb-04b0012beaa2
```

### 4️⃣ Página de Resultado (`/pago/resultado`)
```javascript
// frontend/src/pages/PagoResultado.jsx

useEffect(() => {
  // 1. Leer parámetros de URL
  const mpStatus = searchParams.get('status')
  const collection_id = searchParams.get('collection_id')
  const external_reference = searchParams.get('external_reference')

  // 2. Sincronizar con backend
  const syncResponse = await fetch(
    `http://localhost:8083/payments/${external_reference}/sync`
  )
  const syncedPayment = await syncResponse.json()

  // 3. Determinar estado final
  if (syncedPayment.status === 'completed') {
    setStatus('success')
    toast.success('¡Pago exitoso!')
  }

  // 4. Mostrar detalles del pago
  setPaymentInfo({
    id: syncedPayment.id,
    amount: syncedPayment.amount,
    status: syncedPayment.status,
    transaction_id: syncedPayment.transaction_id
  })
}, [searchParams])
```

### 5️⃣ Estados Posibles

| Estado MP | Estado Backend | UI mostrada | Acción |
|-----------|---------------|-------------|--------|
| `approved` | `completed` | ✅ Pago exitoso | Navegar a /mi-suscripcion |
| `pending` | `pending` | ⏰ Pago en proceso | Navegar a /mi-suscripcion |
| `rejected` | `rejected` | ❌ Pago rechazado | Volver a /planes |

### 6️⃣ Webhooks (Asíncrono)
MercadoPago también envía notificaciones al backend:

```
POST http://localhost:8083/webhooks/mercadopago
{
  "type": "payment",
  "data": {
    "id": "123456789"
  }
}
```

El backend:
1. Verifica la firma (si está configurado WEBHOOK_SECRET)
2. Obtiene el pago de MercadoPago API
3. Actualiza el estado en la base de datos
4. Publica evento a RabbitMQ

---

## 🧪 Cómo Probar

### Opción 1: Flow completo manual

```bash
# 1. Levantar servicios
docker-compose up -d

# 2. Levantar frontend
cd frontend
npm run dev

# 3. Abrir navegador
# http://localhost:5173

# 4. Login con usuario existente o registrarse

# 5. Ir a Planes
# http://localhost:5173/planes

# 6. Seleccionar un plan → Suscribirse

# 7. En Checkout, seleccionar "Mercado Pago"

# 8. Click en "Pagar con Mercado Pago"

# 9. Se abre MercadoPago Checkout Pro
# Usar tarjeta de prueba:
# - Número: 4509 9535 6623 3704
# - Vencimiento: 11/25
# - CVV: 123
# - Nombre: APRO (para aprobado)

# 10. Completar pago

# 11. MercadoPago redirige a /pago/resultado

# 12. Ver resultado del pago
```

### Opción 2: Test con curl

```bash
# Crear pago directo
curl -X POST http://localhost:8083/payments/process \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "subscription",
    "entity_id": "sub_test_123",
    "user_id": "5",
    "amount": 3000.0,
    "currency": "ARS",
    "payment_method": "credit_card",
    "payment_gateway": "mercadopago",
    "callback_url": "http://localhost:5173/pago/resultado",
    "metadata": {
      "customer_email": "test@example.com",
      "customer_name": "Usuario Test"
    }
  }'

# Respuesta incluirá payment_url para abrir en navegador
```

---

## 📊 Endpoints de Payments API

### 1. Crear y Procesar Pago
```
POST /payments/process
Content-Type: application/json

{
  "entity_type": "subscription",
  "entity_id": "sub_123",
  "user_id": "5",
  "amount": 3000,
  "currency": "ARS",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "callback_url": "http://localhost:5173/pago/resultado"
}

Response 201:
{
  "id": "payment_abc123",
  "transaction_id": "pref_id_from_mp",
  "status": "pending",
  "metadata": {
    "payment_url": "https://mercadopago.com/checkout/..."
  }
}
```

### 2. Sincronizar Estado con MercadoPago
```
GET /payments/{payment_id}/sync

Response 200:
{
  "id": "payment_abc123",
  "status": "completed",
  "amount": 3000,
  "currency": "ARS",
  "transaction_id": "mp_payment_id"
}
```

### 3. Webhook de MercadoPago
```
POST /webhooks/mercadopago
Content-Type: application/json

{
  "type": "payment",
  "data": {
    "id": "123456789"
  }
}

Response 200: OK
```

---

## 🎨 UI de PagoResultado

### Estados y Diseño

#### ✅ Pago Exitoso (approved/completed)
```
┌────────────────────────────────────┐
│           ✅ (80px)                │
│     ¡Pago exitoso!                 │
│  Tu suscripción ha sido activada   │
│                                    │
│  ┌──────────────────────────┐     │
│  │  Detalles del pago       │     │
│  │  ID: payment_abc123      │     │
│  │  Monto: $3000 ARS        │     │
│  │  Estado: COMPLETED       │     │
│  └──────────────────────────┘     │
│                                    │
│  [Ver mi suscripción]              │
│                                    │
│  ¿Necesitás ayuda?                 │
│  0351-123-4567                     │
└────────────────────────────────────┘
```

#### ⏰ Pago Pendiente
- Ícono: ⏰
- Color: #ff9800 (naranja)
- Botón: "Ver mi suscripción"

#### ❌ Pago Rechazado
- Ícono: ❌
- Color: #f44336 (rojo)
- Botones: "Volver a planes" + "Elegir otro plan"

---

## 🔧 Configuración de Entorno

### Variables Requeridas

```bash
# .env (raíz del proyecto)
MERCADOPAGO_ACCESS_TOKEN=APP_USR-988447898124356-103117-055b2effa25fcae685ce1e460f2d1d1f-2959351968
MERCADOPAGO_PUBLIC_KEY=APP_USR-ece25914-5b84-4029-9146-5ace6356a8cf
MERCADOPAGO_WEBHOOK_SECRET=  # Opcional, para validar webhooks
```

### Docker Compose
El archivo `docker-compose.yml` ya está configurado para pasar estas variables al contenedor `payments-api`.

---

## ✅ Checklist de Integración

- [x] Backend: Gateway de MercadoPago implementado
- [x] Backend: Endpoints de payments-api funcionando
- [x] Backend: Webhooks configurados
- [x] Backend: Sync endpoint implementado
- [x] Configuración: Variables de entorno en .env
- [x] Configuración: Docker compose actualizado
- [x] Frontend: Checkout.jsx integrado con MP
- [x] Frontend: PagoResultado.jsx creado
- [x] Frontend: PagoResultado.css creado
- [x] Frontend: Route agregada a main.jsx
- [x] Frontend: Callback URL actualizado
- [x] Testing: Build de frontend exitoso
- [x] Testing: Pago de prueba exitoso

---

## 🚀 Próximos Pasos Opcionales

1. **Webhooks en Producción**
   - Configurar ngrok o exponer URL pública
   - Registrar webhook en panel de MercadoPago

2. **Suscripciones Recurrentes**
   - Usar MercadoPago Subscriptions API
   - Implementar auto-renovación de planes

3. **Más Métodos de Pago**
   - Agregar soporte para efectivo en Rapipago/PagoFácil
   - Transferencias bancarias

4. **Mejorar UX**
   - Modal de confirmación antes de pagar
   - Animación de loading durante redirect
   - Email de confirmación post-pago

---

## 📞 Soporte

Si hay problemas con la integración:

1. **Verificar logs del backend**
   ```bash
   docker-compose logs -f payments-api
   ```

2. **Verificar estado del pago**
   ```bash
   curl http://localhost:8083/payments/{payment_id}
   ```

3. **Sincronizar manualmente**
   ```bash
   curl http://localhost:8083/payments/{payment_id}/sync
   ```

4. **Verificar credenciales**
   - Panel de MercadoPago: https://www.mercadopago.com.ar/developers/panel
   - Verificar que el ACCESS_TOKEN corresponde a la cuenta correcta

---

## 📝 Notas Importantes

- Las credenciales actuales son de **DESARROLLO** (modo sandbox)
- Para producción, reemplazar con credenciales de producción
- El webhook_url debe ser accesible públicamente (usar ngrok en desarrollo)
- Los pagos de prueba no generan cargos reales
- Tarjetas de prueba: https://www.mercadopago.com.ar/developers/es/docs/testing/test-cards

---

**Última actualización:** 2025-11-25
**Estado:** ✅ Integración completa y funcional
