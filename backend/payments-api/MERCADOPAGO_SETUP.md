# 🚀 Guía de Configuración de Mercado Pago

## ✅ Estado Actual

La integración con Mercado Pago está **completamente implementada** en el código, pero requiere configuración de credenciales.

### Implementación Completa:
- ✅ Gateway de Mercado Pago (`internal/gateways/mercadopago_gateway.go`)
- ✅ Checkout Pro (pagos únicos)
- ✅ Preapprovals (suscripciones recurrentes)
- ✅ Webhooks para notificaciones asíncronas
- ✅ Reembolsos
- ✅ Validación de credenciales
- ✅ Eventos RabbitMQ integrados

---

## 🔑 Paso 1: Obtener Credenciales de Mercado Pago

### Para Testing (GRATIS - Recomendado)

1. **Crear cuenta de desarrollador**:
   - Ve a: https://www.mercadopago.com.ar/developers
   - Haz clic en "Crear cuenta" o inicia sesión

2. **Acceder al panel de credenciales**:
   - URL: https://www.mercadopago.com.ar/developers/panel/credentials
   - Selecciona **"Credenciales de prueba"**

3. **Copiar credenciales**:
   - `Access Token` → Empieza con `TEST-` (ej: `TEST-1234567890-123456-abc...`)
   - `Public Key` → Empieza con `TEST-` (ej: `TEST-abc123-def4-56...`)

### Tarjetas de Prueba

```
✅ APROBADA (pago exitoso):
   Número: 5031 7557 3453 0604
   CVV: 123
   Vencimiento: 11/25
   Nombre: APRO

❌ RECHAZADA (pago fallido):
   Número: 5031 4332 1540 6351
   CVV: 123
   Vencimiento: 11/25
   Nombre: OTHE

⏳ PENDIENTE:
   Número: 5031 4332 1540 6351
   CVV: 123
   Vencimiento: 11/25
   Nombre: CONT
```

Documentación completa: https://www.mercadopago.com.ar/developers/es/docs/checkout-pro/additional-content/test-cards

---

## 🛠️ Paso 2: Configurar Variables de Entorno

### Editar el archivo `.env`

**Ubicación**: `backend/payments-api/.env`

Reemplaza `YOUR_ACCESS_TOKEN_HERE` y `YOUR_PUBLIC_KEY_HERE` con tus credenciales:

```bash
# Mercado Pago - CREDENCIALES DE PRUEBA
MERCADOPAGO_ACCESS_TOKEN=TEST-1234567890-123456-abcdef1234567890abcdef1234567890-123456789
MERCADOPAGO_PUBLIC_KEY=TEST-abcdef12-3456-7890-abcd-ef1234567890
MERCADOPAGO_WEBHOOK_SECRET=
```

**IMPORTANTE**:
- ❌ NO commitees el archivo `.env` con credenciales reales a Git
- ✅ El archivo `.env` ya está en `.gitignore`
- ✅ Usa `.env.example` como plantilla para otros desarrolladores

---

## 🧪 Paso 3: Probar la Integración

### 1. Levantar los servicios

```bash
docker-compose up -d
```

### 2. Verificar que payments-api esté funcionando

```bash
# Health check
curl http://localhost:8083/healthz

# Debería retornar:
{
  "status": "ok",
  "service": "payments-api",
  "checks": {
    "mongodb": "connected",
    "rabbitmq": "connected"
  }
}
```

### 3. Crear un pago de prueba con Mercado Pago

```bash
curl -X POST http://localhost:8083/payments/process \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "subscription",
    "entity_id": "sub_123",
    "user_id": "5",
    "amount": 1000.0,
    "currency": "ARS",
    "payment_method": "credit_card",
    "payment_gateway": "mercadopago",
    "callback_url": "http://localhost:5173/pagos/resultado",
    "metadata": {
      "customer_email": "test@example.com",
      "customer_name": "Usuario Test"
    }
  }'
```

**Respuesta esperada**:
```json
{
  "id": "67890abcdef1234567890abc",
  "entity_type": "subscription",
  "entity_id": "sub_123",
  "user_id": "5",
  "amount": 1000,
  "currency": "ARS",
  "status": "pending",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "transaction_id": "1234567-abc-def-123",
  "metadata": {
    "payment_url": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=1234567-abc...",
    "gateway_message": "Preferencia creada. Redirigir al usuario a payment_url",
    "init_point": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=..."
  },
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

### 4. Completar el pago

- Copia la URL de `metadata.payment_url`
- Ábrela en el navegador
- Completa el formulario con una **tarjeta de prueba**
- Mercado Pago redirigirá a tu `callback_url` con el resultado

---

## 🔔 Paso 4: Configurar Webhooks (Opcional)

Los webhooks permiten que Mercado Pago notifique cambios de estado automáticamente.

### 1. Configurar URL pública (Desarrollo)

Usa **ngrok** para exponer tu localhost:

```bash
# Instalar ngrok
# Windows: choco install ngrok
# Mac: brew install ngrok

# Exponer puerto 8083
ngrok http 8083
```

Ejemplo de salida:
```
Forwarding  https://abc123.ngrok.io -> http://localhost:8083
```

### 2. Configurar webhook en Mercado Pago

1. Ve a: https://www.mercadopago.com.ar/developers/panel/notifications/webhooks
2. Crea un nuevo webhook:
   - **URL**: `https://abc123.ngrok.io/webhooks/mercadopago`
   - **Eventos**: Selecciona "payment"
3. Guarda

### 3. Probar webhook

```bash
# Logs del payments-api
docker-compose logs -f payments-api

# Deberías ver:
[Webhook] Notificación recibida de Mercado Pago
[Webhook] Payment ID: 1234567890
[Webhook] Nuevo estado: approved
✅ Pago actualizado a 'completed'
```

---

## 📊 Endpoints Disponibles

### Pagos Únicos (Checkout Pro)

```bash
# Crear pago único
POST /payments/process
```

### Pagos Recurrentes (Suscripciones)

```bash
# Crear suscripción mensual
POST /payments/recurring
{
  "entity_type": "subscription",
  "entity_id": "sub_123",
  "user_id": "5",
  "amount": 3000.0,
  "currency": "ARS",
  "payment_method": "credit_card",
  "payment_gateway": "mercadopago",
  "frequency": 1,
  "frequency_type": "months"
}
```

### Consultar Estado

```bash
# Obtener pago
GET /payments/:id

# Sincronizar con gateway
GET /payments/:id/sync
```

### Reembolsos

```bash
# Reembolsar pago
POST /payments/:id/refund
{
  "amount": 1000.0
}
```

---

## 🐛 Troubleshooting

### Error: "Credenciales inválidas"

**Causa**: Access Token incorrecto o expirado

**Solución**:
1. Verifica que copiaste el Access Token completo
2. Asegúrate de usar credenciales de **"Credenciales de prueba"**
3. Regenera las credenciales si es necesario

### Error: "Gateway not supported"

**Causa**: El gateway factory no encuentra el gateway

**Solución**:
- Verifica que `payment_gateway: "mercadopago"` esté en minúsculas
- Revisa logs: `docker-compose logs payments-api`

### Webhook no se recibe

**Causa**: URL no accesible desde internet

**Solución**:
1. Usa ngrok para desarrollo local
2. En producción, usa HTTPS con dominio público
3. Verifica firewall/puertos abiertos

---

## 🔒 Seguridad

### Credenciales de Producción

**NUNCA** commits credenciales de producción en Git:

```bash
# ✅ Correcto
MERCADOPAGO_ACCESS_TOKEN=TEST-...  # Testing

# ❌ INCORRECTO
MERCADOPAGO_ACCESS_TOKEN=APP_USR-...  # Producción en código
```

### Mejores Prácticas

1. ✅ Usa variables de entorno
2. ✅ Archivo `.env` en `.gitignore`
3. ✅ Usa credenciales de TEST para desarrollo
4. ✅ Rota credenciales si se exponen
5. ✅ Valida webhooks con firma (webhook_secret)

---

## 📚 Documentación Oficial

- **Checkout Pro**: https://www.mercadopago.com.ar/developers/es/docs/checkout-pro/landing
- **Preapprovals**: https://www.mercadopago.com.ar/developers/es/docs/subscriptions/landing
- **Webhooks**: https://www.mercadopago.com.ar/developers/es/docs/your-integrations/notifications/webhooks
- **API Reference**: https://www.mercadopago.com.ar/developers/es/reference

---

## ✅ Checklist de Configuración

- [ ] Crear cuenta de desarrollador en Mercado Pago
- [ ] Obtener credenciales de prueba (Access Token + Public Key)
- [ ] Editar archivo `.env` con credenciales
- [ ] Reiniciar `payments-api`: `docker-compose restart payments-api`
- [ ] Probar health check: `curl http://localhost:8083/healthz`
- [ ] Crear pago de prueba: `POST /payments/process`
- [ ] Completar pago con tarjeta de prueba
- [ ] (Opcional) Configurar webhooks con ngrok

---

## 🎉 ¡Listo!

Una vez completados estos pasos, Mercado Pago estará 100% funcional en tu aplicación.

**Próximos pasos**:
- Integrar el frontend para mostrar el Checkout Pro
- Configurar página de éxito/error después del pago
- Implementar notificaciones por email al usuario
- Probar flujo completo: Suscripción → Pago → Activación
