# Guía Completa de Testing - Subscriptions API

## 📋 Índice
1. [Preparación del Entorno](#preparación-del-entorno)
2. [Health Check](#health-check)
3. [Testing de Planes](#testing-de-planes)
4. [Testing de Suscripciones](#testing-de-suscripciones)
5. [Testing de Paginación](#testing-de-paginación)
6. [Testing de Autenticación JWT](#testing-de-autenticación-jwt)
7. [Testing Automatizado](#testing-automatizado)

---

## 🚀 Preparación del Entorno

### 1. Levantar Dependencias

```bash
# MongoDB
docker run -d --name mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:latest

# RabbitMQ
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management

# Verificar que estén corriendo
docker ps
```

### 2. Configurar .env

Crea `.env` en la raíz:
```env
PORT=8081
MONGO_URI=mongodb://root:admin@localhost:27017
MONGO_DATABASE=gym_subscriptions
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=gym_events
USERS_API_URL=http://localhost:8080
JWT_SECRET=my-super-secret-key-for-testing
```

### 3. Iniciar el Servicio

```bash
# Instalar dependencias
go mod tidy

# Ejecutar
go run cmd/api/main.go
```

**Salida esperada:**
```
✅ Conectado a MongoDB exitosamente
📊 Creando índices de MongoDB...
✅ Índices de planes creados
✅ Índices de suscripciones creados
✅ Conectado a RabbitMQ (Exchange: gym_events)
🚀 Subscriptions API corriendo en puerto 8081
```

---

## 🩺 Health Check

### Verificar Estado del Servicio

```bash
curl http://localhost:8081/healthz
```

**Respuesta (todo OK):**
```json
{
  "status": "healthy",
  "service": "subscriptions-api",
  "checks": {
    "mongodb": "healthy",
    "rabbitmq": "healthy"
  },
  "uptime": "5m23.456789s",
  "version": "1.0.0"
}
```

**Respuesta (MongoDB caído):**
```json
{
  "status": "degraded",
  "service": "subscriptions-api",
  "checks": {
    "mongodb": "unhealthy",
    "rabbitmq": "healthy"
  },
  "uptime": "10m12.456789s",
  "version": "1.0.0"
}
```

**HTTP Status Code:**
- `200 OK` - Todo funcional
- `503 Service Unavailable` - Alguna dependencia caída

---

## 📦 Testing de Planes

### 1. Listar Planes (Público - Sin Autenticación)

```bash
curl http://localhost:8081/plans
```

**Respuesta (sin planes):**
```json
{
  "plans": [],
  "total": 0,
  "page": 1,
  "page_size": 10,
  "total_pages": 0
}
```

### 2. Crear Plan (Requiere Admin)

**Generar Token de Admin:**
Ve a [https://jwt.io](https://jwt.io)

**Payload:**
```json
{
  "user_id": "1",
  "username": "admin",
  "role": "admin",
  "exp": 9999999999,
  "iat": 1700000000
}
```

**Secret:** `my-super-secret-key-for-testing`

**Token generado:** `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInVzZXJuYW1lIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTcwMDAwMDAwMH0.qLj7-Qx4_kQw9Zt5YNZ5xYqWZ5qWZ5qWZ5qWZ5qWZ5g`

**Request:**
```bash
curl -X POST http://localhost:8081/plans \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInVzZXJuYW1lIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTcwMDAwMDAwMH0.Yo0Dqhvt8rLpBqBXqNQHOaUz9KSI-3VQXfL9KRvQdvg" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Premium",
    "descripcion": "Acceso completo al gimnasio",
    "precio_mensual": 100.00,
    "tipo_acceso": "completo",
    "duracion_dias": 30,
    "activo": true,
    "actividades_permitidas": ["gym", "pool", "classes", "sauna"]
  }'
```

**Respuesta:**
```json
{
  "id": "673f4e20bcf86cd799439011",
  "nombre": "Plan Premium",
  "descripcion": "Acceso completo al gimnasio",
  "precio_mensual": 100.00,
  "tipo_acceso": "completo",
  "duracion_dias": 30,
  "activo": true,
  "actividades_permitidas": ["gym", "pool", "classes", "sauna"],
  "created_at": "2025-11-11T10:30:00Z",
  "updated_at": "2025-11-11T10:30:00Z"
}
```

### 3. Crear Más Planes (para probar paginación)

```bash
# Plan Basic
curl -X POST http://localhost:8081/plans \
  -H "Authorization: Bearer TU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Basic",
    "descripcion": "Acceso básico",
    "precio_mensual": 50.00,
    "tipo_acceso": "limitado",
    "duracion_dias": 30,
    "activo": true,
    "actividades_permitidas": ["gym"]
  }'

# Plan Gold
curl -X POST http://localhost:8081/plans \
  -H "Authorization: Bearer TU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Gold",
    "descripcion": "Acceso premium plus",
    "precio_mensual": 150.00,
    "tipo_acceso": "completo",
    "duracion_dias": 30,
    "activo": true,
    "actividades_permitidas": ["gym", "pool", "classes", "sauna", "spa"]
  }'
```

### 4. Obtener Plan por ID

```bash
curl http://localhost:8081/plans/673f4e20bcf86cd799439011
```

---

## 🎫 Testing de Suscripciones

### 1. Crear Suscripción (Requiere Usuario Autenticado)

**Generar Token de Usuario:**

**Payload:**
```json
{
  "user_id": "5",
  "username": "user123",
  "role": "user",
  "exp": 9999999999,
  "iat": 1700000000
}
```

**Secret:** `my-super-secret-key-for-testing`

**Request:**
```bash
curl -X POST http://localhost:8081/subscriptions \
  -H "Authorization: Bearer TU_TOKEN_USUARIO" \
  -H "Content-Type: application/json" \
  -d '{
    "usuario_id": "5",
    "plan_id": "673f4e20bcf86cd799439011",
    "sucursal_origen_id": "sucursal-1",
    "metodo_pago": "credit_card",
    "auto_renovacion": true,
    "notas": "Cliente preferencial"
  }'
```

**⚠️ Importante:** Para que funcione, `users-api` debe estar corriendo en `localhost:8080` para validar el usuario.

**Respuesta:**
```json
{
  "id": "673f5a30bcf86cd799439022",
  "usuario_id": "5",
  "plan_id": "673f4e20bcf86cd799439011",
  "plan_nombre": "Plan Premium",
  "sucursal_origen_id": "sucursal-1",
  "fecha_inicio": "2025-11-11T10:35:00Z",
  "fecha_vencimiento": "2025-12-11T10:35:00Z",
  "estado": "pendiente_pago",
  "metodo_pago_preferido": "credit_card",
  "auto_renovacion": true,
  "notas": "Cliente preferencial",
  "historial_renovaciones": [],
  "created_at": "2025-11-11T10:35:00Z",
  "updated_at": "2025-11-11T10:35:00Z"
}
```

### 2. Obtener Suscripción por ID

```bash
curl http://localhost:8081/subscriptions/673f5a30bcf86cd799439022 \
  -H "Authorization: Bearer TU_TOKEN_USUARIO"
```

### 3. Obtener Suscripción Activa de Usuario

```bash
curl http://localhost:8081/subscriptions/active/5 \
  -H "Authorization: Bearer TU_TOKEN_USUARIO"
```

### 4. Actualizar Estado de Suscripción

```bash
curl -X PATCH http://localhost:8081/subscriptions/673f5a30bcf86cd799439022/status \
  -H "Authorization: Bearer TU_TOKEN_USUARIO" \
  -H "Content-Type: application/json" \
  -d '{
    "estado": "activa",
    "pago_id": "payment-123"
  }'
```

**Estados válidos:** `activa`, `vencida`, `cancelada`, `pendiente_pago`

### 5. Cancelar Suscripción

```bash
curl -X DELETE http://localhost:8081/subscriptions/673f5a30bcf86cd799439022 \
  -H "Authorization: Bearer TU_TOKEN_USUARIO"
```

---

## 📄 Testing de Paginación

### Listar Planes con Paginación

```bash
# Página 1, 10 resultados
curl "http://localhost:8081/plans?page=1&page_size=10"

# Página 2, 5 resultados
curl "http://localhost:8081/plans?page=2&page_size=5"

# Filtrar solo activos
curl "http://localhost:8081/plans?activo=true&page=1&page_size=10"

# Ordenar por precio ascendente
curl "http://localhost:8081/plans?sort_by=precio_mensual&sort_desc=false"

# Ordenar por precio descendente
curl "http://localhost:8081/plans?sort_by=precio_mensual&sort_desc=true"

# Ordenar por nombre
curl "http://localhost:8081/plans?sort_by=nombre"

# Combinación
curl "http://localhost:8081/plans?activo=true&page=1&page_size=5&sort_by=precio_mensual&sort_desc=true"
```

**Respuesta:**
```json
{
  "plans": [
    { "id": "...", "nombre": "Plan Gold", "precio_mensual": 150 },
    { "id": "...", "nombre": "Plan Premium", "precio_mensual": 100 },
    { "id": "...", "nombre": "Plan Basic", "precio_mensual": 50 }
  ],
  "total": 3,
  "page": 1,
  "page_size": 5,
  "total_pages": 1
}
```

**Notas:**
- `page` por defecto: 1
- `page_size` por defecto: 10
- `page_size` máximo: 100
- `sort_by` por defecto: `created_at`
- Los resultados **realmente** están paginados en MongoDB (usa skip/limit)

---

## 🔐 Testing de Autenticación JWT

### 1. Sin Token (401 Unauthorized)

```bash
curl -X POST http://localhost:8081/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"usuario_id": "5", "plan_id": "123"}'
```

**Respuesta:**
```json
{
  "error": "Token de autorización requerido"
}
```

### 2. Token Inválido (401 Unauthorized)

```bash
curl -X POST http://localhost:8081/subscriptions \
  -H "Authorization: Bearer token_invalido" \
  -H "Content-Type: application/json" \
  -d '{"usuario_id": "5", "plan_id": "123"}'
```

**Respuesta:**
```json
{
  "error": "Token inválido o expirado"
}
```

### 3. Usuario Intentando Crear Plan (403 Forbidden)

```bash
curl -X POST http://localhost:8081/plans \
  -H "Authorization: Bearer TOKEN_DE_USUARIO_ROLE_USER" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Plan Test", ...}'
```

**Respuesta:**
```json
{
  "error": "Acceso denegado: permisos insuficientes"
}
```

### 4. Formato Incorrecto (401 Unauthorized)

```bash
# Sin "Bearer"
curl -X POST http://localhost:8081/subscriptions \
  -H "Authorization: eyJhbGciOiJ..." \
  -H "Content-Type: application/json" \
  -d '{"usuario_id": "5", ...}'
```

**Respuesta:**
```json
{
  "error": "Formato de token inválido. Use: Bearer {token}"
}
```

---

## 🧪 Testing Automatizado

### Tests Unitarios

```bash
# Ejecutar todos los tests
go test ./internal/services/... -v

# Con cobertura
go test ./internal/services/... -cover

# Ver cobertura detallada
go test ./internal/services/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Test específico
go test ./internal/services/... -run TestPlanService_CreatePlan -v
```

### Tests de Integración con cURL Script

Crea `test-api.sh`:
```bash
#!/bin/bash

API_URL="http://localhost:8081"
ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
USER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

echo "🩺 Testing Health Check..."
curl -s $API_URL/healthz | jq .

echo "📦 Creating Plan..."
PLAN_RESPONSE=$(curl -s -X POST $API_URL/plans \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nombre":"Plan Test","precio_mensual":100,"tipo_acceso":"completo","duracion_dias":30,"activo":true}')

PLAN_ID=$(echo $PLAN_RESPONSE | jq -r '.id')
echo "Plan created: $PLAN_ID"

echo "🎫 Creating Subscription..."
curl -s -X POST $API_URL/subscriptions \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"usuario_id\":\"5\",\"plan_id\":\"$PLAN_ID\",\"metodo_pago\":\"credit_card\"}" | jq .

echo "✅ Tests completed!"
```

```bash
chmod +x test-api.sh
./test-api.sh
```

---

## 📱 Testing con Postman/Insomnia

### Importar Colección

Crea `subscriptions-api.postman_collection.json`:
```json
{
  "info": {
    "name": "Subscriptions API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/healthz",
          "host": ["{{base_url}}"],
          "path": ["healthz"]
        }
      }
    },
    {
      "name": "List Plans",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/plans?page=1&page_size=10",
          "host": ["{{base_url}}"],
          "path": ["plans"],
          "query": [
            {"key": "page", "value": "1"},
            {"key": "page_size", "value": "10"}
          ]
        }
      }
    },
    {
      "name": "Create Plan (Admin)",
      "request": {
        "method": "POST",
        "header": [
          {"key": "Authorization", "value": "Bearer {{admin_token}}"},
          {"key": "Content-Type", "value": "application/json"}
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"nombre\": \"Plan Premium\",\n  \"precio_mensual\": 100.00,\n  \"tipo_acceso\": \"completo\",\n  \"duracion_dias\": 30,\n  \"activo\": true\n}"
        },
        "url": {
          "raw": "{{base_url}}/plans",
          "host": ["{{base_url}}"],
          "path": ["plans"]
        }
      }
    }
  ],
  "variable": [
    {"key": "base_url", "value": "http://localhost:8081"},
    {"key": "admin_token", "value": ""},
    {"key": "user_token", "value": ""}
  ]
}
```

**Variables de entorno en Postman:**
- `base_url`: `http://localhost:8081`
- `admin_token`: Tu token de admin
- `user_token`: Tu token de usuario

---

## 🐛 Troubleshooting

### Error: "dial tcp [::1]:27017: connect: connection refused"
**Solución:** MongoDB no está corriendo
```bash
docker start mongo
# o
docker run -d --name mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=admin mongo
```

### Error: "usuario no válido"
**Solución:** `users-api` no está corriendo o el usuario no existe
```bash
# Verificar users-api
curl http://localhost:8080/users/5

# Si no existe, créalo primero en users-api
```

### Health Check retorna "degraded"
**Solución:** Verifica el campo `checks` para ver qué dependencia falló
```json
{
  "checks": {
    "mongodb": "unhealthy",  // ← MongoDB caído
    "rabbitmq": "healthy"
  }
}
```

---

## ✅ Checklist de Testing Completo

- [ ] Health check retorna status "healthy"
- [ ] MongoDB status: "healthy"
- [ ] RabbitMQ status: "healthy" o "disabled"
- [ ] Crear plan sin token retorna 401
- [ ] Crear plan con token de usuario retorna 403
- [ ] Crear plan con token de admin funciona
- [ ] Listar planes funciona sin token
- [ ] Paginación funciona (diferentes páginas retornan diferentes resultados)
- [ ] Ordenamiento funciona (sort_by)
- [ ] Crear suscripción requiere token
- [ ] Obtener suscripción requiere token
- [ ] Actualizar estado funciona
- [ ] Cancelar suscripción funciona
- [ ] Tests unitarios pasan: `go test ./... -v`

🎉 **Si todos los checks pasan, tu API está completamente funcional!**
