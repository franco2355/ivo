# Guía de Implementación: Sistema de Gateways de Pago

## Tabla de Contenidos
- [Objetivo](#objetivo)
- [Pre-requisitos](#pre-requisitos)
- [Arquitectura a Implementar](#arquitectura-a-implementar)
- [Paso 1: Crear Interfaces y DTOs](#paso-1-crear-interfaces-y-dtos)
- [Paso 2: Implementar Factory](#paso-2-implementar-factory)
- [Paso 3: Implementar Mercado Pago Gateway](#paso-3-implementar-mercado-pago-gateway)
- [Paso 4: Crear Servicio Mejorado](#paso-4-crear-servicio-mejorado)
- [Paso 5: Actualizar Configuración](#paso-5-actualizar-configuración)
- [Paso 6: Crear Controlador de Webhooks](#paso-6-crear-controlador-de-webhooks)
- [Paso 7: Actualizar Main y Rutas](#paso-7-actualizar-main-y-rutas)
- [Paso 8: Testing](#paso-8-testing)
- [Checklist de Verificación](#checklist-de-verificación)
- [Troubleshooting](#troubleshooting)

---

## Objetivo

Implementar una arquitectura genérica y extensible que permita:
1. Integrar Mercado Pago como primera pasarela de pago
2. Facilitar agregar nuevas pasarelas (Stripe, PayPal, etc.) sin modificar código existente
3. Mantener la lógica de negocio desacoplada de implementaciones específicas
4. Soportar webhooks para recibir notificaciones asíncronas

---

## Pre-requisitos

### Conocimientos Necesarios
- Go básico (structs, interfaces, paquetes)
- HTTP/REST APIs
- Patrones: Strategy, Factory, Dependency Injection
- MongoDB (ya implementado en el proyecto)

### Herramientas
- Go 1.21+
- MongoDB (ya configurado)
- Cuenta de Mercado Pago (obtener credenciales en https://www.mercadopago.com.ar/developers)
- Postman o curl para testing

### Credenciales de Mercado Pago
1. Ir a https://www.mercadopago.com.ar/developers/panel
2. Crear aplicación o usar una existente
3. Obtener:
   - `Access Token` (empieza con `APP_USR-...` o `TEST-...`)
   - `Public Key` (empieza con `APP_USR-...` o `TEST-...`)
4. Guardar para usar en `.env`

---

## Arquitectura a Implementar

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
│  │         Payment Service Enhanced                    │          │
│  │  (Lógica de Negocio + Orquestación)                 │          │
│  │                                                      │          │
│  │  ┌────────────────────────────────────────────┐    │          │
│  │  │  Repository Interface                      │    │          │
│  │  │  (ya existe en el proyecto)                │    │          │
│  │  └────────────────────────────────────────────┘    │          │
│  │                                                      │          │
│  │  ┌────────────────────────────────────────────┐    │          │
│  │  │  Gateway Factory                           │    │          │
│  │  │  (Crea instancias de gateways)             │    │          │
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
│  │ Mercado Pago│      │   Stripe    │   │  PayPal  │ │ Mock  │ │
│  │  Gateway    │      │  Gateway    │   │ Gateway  │ │Gateway│ │
│  │  (IMPL)     │      │  (FUTURO)   │   │ (FUTURO) │ │(TEST) │ │
│  └──────┬──────┘      └──────┬──────┘   └────┬─────┘ └───┬───┘ │
│         │                    │               │           │      │
│         │ (HTTP/REST)        │               │           │      │
│         ▼                    ▼               ▼           ▼      │
├─────────────────────────────────────────────────────────────────┤
│  SERVICIOS EXTERNOS                                              │
│                                                                   │
│  [Mercado Pago API]    [Stripe API]    [PayPal API]    [Memory] │
└───────────────────────────────────────────────────────────────────┘
```

---

## Paso 1: Crear Interfaces y DTOs

### Archivo a crear: `internal/gateways/payment_gateway.go`

Este archivo define el contrato que todas las pasarelas deben cumplir.

**¿Qué acabamos de hacer?**
- ✅ Definimos el contrato `PaymentGateway` con 7 métodos
- ✅ Creamos DTOs genéricos que funcionan para cualquier pasarela
- ✅ Establecimos la base del patrón Strategy

---

## Paso 2: Implementar Factory

### Archivo a crear: `internal/gateways/factory.go`

El Factory se encarga de crear instancias de gateways según el nombre solicitado.

**¿Qué acabamos de hacer?**
- ✅ Centralizamos la creación de gateways
- ✅ Inyectamos automáticamente la configuración
- ✅ Facilitamos agregar nuevos gateways (solo un `case` nuevo)

---

## Paso 3: Implementar Mercado Pago Gateway

### Archivo a crear: `internal/gateways/mercadopago/mercadopago_gateway.go`

Esta es la implementación concreta de la interfaz `PaymentGateway` para Mercado Pago.

**¿Qué acabamos de hacer?**
- ✅ Implementamos todos los métodos de `PaymentGateway`
- ✅ Integramos con la API de Mercado Pago
- ✅ Mapeamos estados de MP a nuestros estados genéricos
- ✅ Manejamos errores HTTP

---

## Paso 4: Crear Servicio Mejorado

### Archivo a crear: `internal/services/payment_service_enhanced.go`

Este servicio orquesta la lógica de negocio usando repository y gateways.

**¿Qué acabamos de hacer?**
- ✅ Creamos servicio que orquesta repository + gateway
- ✅ Implementamos flujo completo: DB → Gateway → DB
- ✅ Agregamos métodos para reembolsos y sincronización

---

## Paso 5: Actualizar Configuración

### Modificar archivo existente: `internal/config/config.go`

Agregar configuración para gateways.

---

## Paso 6: Crear Controlador de Webhooks

### Archivo a crear: `internal/controllers/webhook_controller.go`

---

## Paso 7: Actualizar Main y Rutas

### Modificar archivo existente: `cmd/api/main.go`

---

## Paso 8: Testing

### 1. Verificar compilación

```bash
cd payments-api
go mod tidy
go build -o payments-api cmd/api/main.go
```

### 2. Iniciar servidor

```bash
./payments-api
```

### 3. Crear un pago de prueba

```bash
curl -X POST http://localhost:8080/payments/process \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type": "subscription",
    "entity_id": "sub_123",
    "user_id": "user_456",
    "amount": 100.00,
    "currency": "ARS",
    "payment_method": "credit_card",
    "payment_gateway": "mercadopago",
    "metadata": {
      "customer_email": "cliente@test.com",
      "customer_name": "Juan Perez"
    }
  }'
```

---

## Checklist de Verificación

Antes de dar por terminada la implementación, verifica:

- [ ] Todos los archivos compilan sin errores (`go build`)
- [ ] Las credenciales de Mercado Pago están en `.env`
- [ ] El servidor inicia correctamente
- [ ] Puedes crear un pago con `mock` gateway
- [ ] Puedes crear un pago con `mercadopago` gateway (con credenciales reales)
- [ ] El pago se guarda en MongoDB con `transaction_id`
- [ ] Puedes consultar el estado de un pago (`GET /payments/:id`)
- [ ] El endpoint de webhook responde 200 OK
- [ ] La estructura de carpetas coincide con la documentación
- [ ] El código sigue los principios SOLID (Open/Closed, Dependency Inversion)

---

## Troubleshooting

### Error: "gateway no soportado"
- Verificar que el nombre del gateway sea exacto: "mercadopago", "stripe", "mock"
- Verificar que el `switch` en `factory.go` tenga el case correspondiente

### Error: "credenciales inválidas"
- Verificar que `.env` tenga `MERCADOPAGO_ACCESS_TOKEN` y `MERCADOPAGO_PUBLIC_KEY`
- Verificar que los tokens sean de prueba (empiezan con `TEST-`)
- Ir a https://www.mercadopago.com.ar/developers/panel para obtener nuevos tokens

### Error: "error creando registro de pago"
- Verificar que MongoDB esté corriendo
- Verificar conexión: `MONGO_URI=mongodb://localhost:27017`

### El webhook no llega
- Los webhooks solo funcionan con URLs públicas (no `localhost`)
- Usar ngrok para testing local: `ngrok http 8080`
- Configurar la URL de webhook en Mercado Pago Developers

---

## Próximos Pasos (Opcional)

Una vez que esto funcione:

1. **Implementar Stripe**:
   - Crear `internal/gateways/stripe/stripe_gateway.go`
   - Agregar case en factory
   - Configurar credenciales en `.env`

2. **Mejorar Webhooks**:
   - Validar firmas de seguridad
   - Implementar reintentos
   - Agregar logging estructurado

3. **Testing Automatizado**:
   - Unit tests para cada gateway
   - Integration tests con mock gateway
   - End-to-end tests con API de Mercado Pago Sandbox

4. **Documentación API**:
   - Agregar Swagger/OpenAPI
   - Documentar ejemplos de requests/responses
