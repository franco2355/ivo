# Guía Rápida - Subscriptions API

## 🚀 Inicio Rápido

### 1. Instalar Dependencias

```bash
go mod tidy
```

Esto descargará:
- Gin (framework web)
- MongoDB driver
- RabbitMQ client
- JWT library
- Todas las dependencias transitivas

### 2. Configurar Variables de Entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
# Server
PORT=8081
GIN_MODE=debug

# MongoDB
MONGO_URI=mongodb://root:admin@localhost:27017
MONGO_DATABASE=gym_subscriptions

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=gym_events

# External APIs
USERS_API_URL=http://localhost:8080
PAYMENTS_API_URL=http://localhost:8083

# JWT
JWT_SECRET=my-super-secret-key-change-in-production
```

### 3. Levantar Infraestructura (Docker Compose)

```bash
# MongoDB
docker run -d \
  --name mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:latest

# RabbitMQ
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

**Opcional:** Si no tienes MongoDB/RabbitMQ, el servicio funcionará parcialmente:
- Sin MongoDB: ❌ El servicio no iniciará
- Sin RabbitMQ: ✅ Usa `NullEventPublisher` (eventos no se publican)

### 4. Ejecutar el Servicio

```bash
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
📦 Arquitectura: Controllers → Services → Repositories
💉 Dependency Injection: Activada
```

Si RabbitMQ no está disponible:
```
⚠️  Warning: No se pudo conectar a RabbitMQ: dial tcp...
⚠️  Usando NullEventPublisher como fallback
⚠️  Usando NullEventPublisher - Los eventos no serán publicados
```

---

## 🧪 Ejecutar Tests

### Todos los Tests
```bash
go test ./internal/services/... -v
```

### Tests con Cobertura
```bash
go test ./internal/services/... -cover
```

### Test Específico
```bash
# Solo tests de PlanService
go test ./internal/services/... -run TestPlanService -v

# Solo tests de CreateSubscription
go test ./internal/services/... -run TestSubscriptionService_CreateSubscription -v
```

### Salida Esperada (Éxito)
```
=== RUN   TestPlanService_CreatePlan
=== RUN   TestPlanService_CreatePlan/Crear_plan_exitosamente
=== RUN   TestPlanService_CreatePlan/Error_al_crear_plan_en_repositorio
--- PASS: TestPlanService_CreatePlan (0.00s)
    --- PASS: TestPlanService_CreatePlan/Crear_plan_exitosamente (0.00s)
    --- PASS: TestPlanService_CreatePlan/Error_al_crear_plan_en_repositorio (0.00s)
...
PASS
ok      github.com/yourusername/gym-management/subscriptions-api/internal/services
```

---

## 📝 Probar Endpoints

### 1. Health Check (público)
```bash
curl http://localhost:8081/healthz
```

**Respuesta:**
```json
{
  "status": "healthy",
  "service": "subscriptions-api"
}
```

### 2. Listar Planes (público)
```bash
curl http://localhost:8081/plans
```

### 3. Crear Plan (requiere admin)

**Sin token (falla):**
```bash
curl -X POST http://localhost:8081/plans \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Premium",
    "descripcion": "Acceso completo",
    "precio_mensual": 100.00,
    "tipo_acceso": "completo",
    "duracion_dias": 30,
    "activo": true,
    "actividades_permitidas": ["gym", "pool", "classes"]
  }'
```

**Respuesta:**
```json
{
  "error": "Token de autorización requerido"
}
```

**Con token de admin:**
```bash
# Primero genera un token en https://jwt.io
# Payload:
{
  "user_id": "1",
  "username": "admin",
  "role": "admin",
  "exp": 9999999999,
  "iat": 1234567890
}
# Secret: my-super-secret-key-change-in-production

# Luego usa el token:
curl -X POST http://localhost:8081/plans \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Premium",
    "precio_mensual": 100.00,
    "tipo_acceso": "completo",
    "duracion_dias": 30,
    "activo": true
  }'
```

### 4. Crear Suscripción (requiere autenticación)

```bash
# Genera token de usuario en jwt.io
# Payload:
{
  "user_id": "5",
  "username": "user123",
  "role": "user",
  "exp": 9999999999,
  "iat": 1234567890
}

curl -X POST http://localhost:8081/subscriptions \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "usuario_id": "5",
    "plan_id": "673f4e20bcf86cd799439011",
    "sucursal_origen_id": "sucursal-1",
    "metodo_pago": "credit_card",
    "auto_renovacion": true
  }'
```

**Nota:** Necesitas que `users-api` esté corriendo en `localhost:8080` para validar el usuario.

---

## 🐳 Docker

### Build
```bash
docker build -t subscriptions-api .
```

### Run
```bash
docker run -p 8081:8081 --env-file .env subscriptions-api
```

### Con Docker Compose (completo)

Crea `docker-compose.yml`:
```yaml
version: '3.8'

services:
  mongo:
    image: mongo:latest
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: admin

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"

  subscriptions-api:
    build: .
    ports:
      - "8081:8081"
    environment:
      MONGO_URI: mongodb://root:admin@mongo:27017
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/
      USERS_API_URL: http://users-api:8080
    depends_on:
      - mongo
      - rabbitmq
```

```bash
docker-compose up
```

---

## 🔍 Verificar Índices de MongoDB

### Usando mongo shell
```bash
docker exec -it mongo mongosh -u root -p admin

use gym_subscriptions

# Ver índices de planes
db.planes.getIndexes()

# Ver índices de suscripciones
db.suscripciones.getIndexes()
```

**Salida esperada:**
```javascript
// Deberías ver índices como:
[
  { v: 2, key: { _id: 1 }, name: '_id_' },
  { v: 2, key: { activo: 1 }, name: 'idx_planes_activo' },
  { v: 2, key: { nombre: 1 }, name: 'idx_planes_nombre' },
  ...
]
```

---

## 🔧 Troubleshooting

### Error: "go: command not found"
```bash
# Instalar Go
# Windows: https://go.dev/dl/
# Linux: sudo apt install golang-go
# Mac: brew install go
```

### Error: "MongoDB connection failed"
```bash
# Verificar que MongoDB está corriendo
docker ps | grep mongo

# Ver logs
docker logs mongo
```

### Error: "RabbitMQ connection failed"
```bash
# El servicio debería funcionar igual con NullEventPublisher
# Pero si necesitas RabbitMQ:
docker ps | grep rabbitmq
docker logs rabbitmq

# Interfaz web: http://localhost:15672
# Usuario: guest / Contraseña: guest
```

### Error: "usuario no válido" al crear suscripción
```bash
# Verificar que users-api está corriendo
curl http://localhost:8080/users/5

# Si no está corriendo, levántalo primero
```

---

## 📚 Documentación Adicional

- **README.md** - Arquitectura y conceptos clave
- **AUTH.md** - Guía completa de autenticación JWT
- **IMPLEMENTATION_SUMMARY.md** - Resumen de funcionalidades implementadas

---

## ✨ Checklist de Verificación

Antes de considerar el servicio funcionando, verifica:

- [ ] El servicio inicia sin errores
- [ ] MongoDB está conectado (logs: "✅ Conectado a MongoDB")
- [ ] Índices fueron creados (logs: "✅ Índices de... creados")
- [ ] RabbitMQ conectado O NullEventPublisher activo
- [ ] Health check responde: `curl http://localhost:8081/healthz`
- [ ] Tests pasan: `go test ./internal/services/... -v`
- [ ] Endpoints públicos funcionan sin token
- [ ] Endpoints protegidos requieren token

**Estado:** Si todos los checks pasan → ✅ **READY**
