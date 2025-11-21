# Activities API

Microservicio para gestionar actividades, inscripciones y sucursales del gimnasio.

**Estado:** ✅ 100% FUNCIONAL

**Puerto:** 8082

---

## 📋 Descripción

Este microservicio maneja:
- **Actividades**: CRUD completo de clases y actividades del gimnasio
- **Inscripciones**: Gestión de inscripciones de usuarios a actividades
- **Sucursales**: Gestión de sucursales (estructura base creada, pendiente de implementación completa)

---

## 🏗️ Arquitectura

El proyecto sigue **Standard Go Project Layout** con **Dependency Injection**:

```
activities-api/
├── cmd/api/main.go          # Entry point con DI
├── internal/
│   ├── config/              # Configuración por environment
│   ├── domain/              # Modelos de negocio (independientes de BD)
│   ├── dao/                 # Modelos de base de datos (GORM)
│   ├── repository/          # Patrón Repository con interfaces
│   ├── services/            # Lógica de negocio
│   ├── controllers/         # Handlers HTTP
│   └── middleware/          # JWT, CORS, etc.
├── go.mod
├── .env.example
├── Dockerfile
└── README.md
```

---

## 🚀 Instalación y Ejecución

### Pre-requisitos

- Go 1.22+
- MySQL 8.0+
- Git

### 1. Clonar el repositorio

```bash
cd activities-api
```

### 2. Configurar variables de entorno

```bash
cp .env.example .env
# Editar .env con tus credenciales
```

Ejemplo de `.env`:

```env
PORT=8082
DB_USER=root
DB_PASS=root123
DB_HOST=localhost
DB_PORT=3306
DB_SCHEMA=proyecto_integrador
JWT_SECRET=my-super-secret-jwt-key
```

**IMPORTANTE:** El `JWT_SECRET` debe ser el mismo que en `users-api` para que los tokens funcionen correctamente.

### 3. Instalar dependencias

```bash
go mod download
```

### 4. Ejecutar

```bash
go run cmd/api/main.go
```

El servidor estará disponible en `http://localhost:8082`

### 5. Verificar

```bash
curl http://localhost:8082/healthz
```

---

## 📡 Endpoints

### Públicos (sin autenticación)

#### Actividades

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `GET` | `/actividades` | Lista todas las actividades con lugares disponibles |
| `GET` | `/actividades/buscar?id=&titulo=&horario=&categoria=` | Busca actividades por parámetros |
| `GET` | `/actividades/:id` | Obtiene una actividad por ID |

**Ejemplo:**

```bash
# Listar todas las actividades
curl http://localhost:8082/actividades

# Buscar por categoría
curl "http://localhost:8082/actividades/buscar?categoria=Yoga"

# Buscar por horario
curl "http://localhost:8082/actividades/buscar?horario=10:00"

# Obtener actividad por ID
curl http://localhost:8082/actividades/1
```

---

### Protegidos (requieren JWT)

#### Inscripciones

| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| `GET` | `/inscripciones` | Lista inscripciones del usuario autenticado | JWT |
| `POST` | `/inscripciones` | Inscribe al usuario a una actividad | JWT |
| `DELETE` | `/inscripciones` | Desinscribe al usuario de una actividad | JWT |

**Ejemplo:**

```bash
# Listar mis inscripciones
curl http://localhost:8082/inscripciones \
  -H "Authorization: Bearer <tu_token_jwt>"

# Inscribirme a una actividad
curl -X POST http://localhost:8082/inscripciones \
  -H "Authorization: Bearer <tu_token_jwt>" \
  -H "Content-Type: application/json" \
  -d '{"actividad_id": 1}'

# Desinscribirme
curl -X DELETE http://localhost:8082/inscripciones \
  -H "Authorization: Bearer <tu_token_jwt>" \
  -H "Content-Type: application/json" \
  -d '{"actividad_id": 1}'
```

---

### Admin Only (requieren JWT + is_admin=true)

#### Actividades (CRUD Admin)

| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| `POST` | `/actividades` | Crea una nueva actividad | JWT + Admin |
| `PUT` | `/actividades/:id` | Actualiza una actividad | JWT + Admin |
| `DELETE` | `/actividades/:id` | Elimina una actividad | JWT + Admin |

**Ejemplo:**

```bash
# Crear actividad (admin)
curl -X POST http://localhost:8082/actividades \
  -H "Authorization: Bearer <token_admin>" \
  -H "Content-Type: application/json" \
  -d '{
    "titulo": "Yoga Matutino",
    "descripcion": "Clase de yoga para principiantes",
    "cupo": 20,
    "dia": "Lunes",
    "horario_inicio": "10:00",
    "horario_final": "11:00",
    "foto_url": "https://example.com/yoga.jpg",
    "instructor": "Juan Pérez",
    "categoria": "Yoga"
  }'

# Actualizar actividad (admin)
curl -X PUT http://localhost:8082/actividades/1 \
  -H "Authorization: Bearer <token_admin>" \
  -H "Content-Type: application/json" \
  -d '{
    "titulo": "Yoga Avanzado",
    "cupo": 15,
    ...
  }'

# Eliminar actividad (admin)
curl -X DELETE http://localhost:8082/actividades/1 \
  -H "Authorization: Bearer <token_admin>"
```

---

## 🗄️ Modelos de Datos

### Actividad

```go
{
  "id": 1,
  "titulo": "Yoga Matutino",
  "descripcion": "Clase de yoga para principiantes",
  "cupo": 20,
  "dia": "Lunes",
  "horario_inicio": "10:00",
  "horario_final": "11:00",
  "foto_url": "https://example.com/yoga.jpg",
  "instructor": "Juan Pérez",
  "categoria": "Yoga",
  "sucursal_id": 1,        // nullable
  "lugares": 15            // calculado automáticamente
}
```

### Inscripción

```go
{
  "id": 1,
  "usuario_id": 5,
  "actividad_id": 1,
  "fecha_inscripcion": "2025-01-15T10:30:00Z",
  "is_activa": true,
  "suscripcion_id": "abc123"  // TODO: cuando subscriptions-api esté listo
}
```

---

## 🔒 Validaciones de Negocio

### Actividades

- **BeforeUpdate Hook (GORM)**: No se puede reducir el cupo si hay más inscripciones activas que el nuevo límite
- **Horarios**: Deben estar en formato "HH:MM" (ej: "10:00")
- **Hora fin**: Debe ser posterior a hora inicio

### Inscripciones

- **BeforeCreate Hook (GORM)**: No se puede inscribir si el cupo está lleno
- **BeforeUpdate Hook (GORM)**: No se puede reactivar si el cupo está lleno
- **Unique Constraint**: Un usuario no puede inscribirse dos veces a la misma actividad (activa)
- **Soft Delete**: Las desinscripciones son lógicas (`is_activa=false`), se pueden reactivar

---

## 📊 Vista MySQL Automática

El repositorio crea automáticamente la vista `actividades_lugares`:

```sql
CREATE OR REPLACE VIEW actividades_lugares AS
SELECT a.*,
       a.cupo - COALESCE((SELECT COUNT(*)
                          FROM inscripciones i
                          WHERE i.actividad_id = a.id_actividad
                          AND i.is_activa = true
                          AND i.deleted_at IS NULL), 0) AS lugares
FROM actividades a
```

Esta vista calcula en tiempo real los cupos disponibles (`lugares`) restando las inscripciones activas del cupo total.

---

## 🐳 Docker

### Build

```bash
docker build -t activities-api .
```

### Run

```bash
docker run -p 8082:8082 \
  -e DB_HOST=mysql \
  -e DB_USER=root \
  -e DB_PASS=root123 \
  -e JWT_SECRET=my-secret \
  activities-api
```

### Con Docker Compose

Usa el archivo `docker-compose.new.yml` en la raíz del proyecto:

```bash
cd ..
docker-compose -f docker-compose.new.yml up --build activities-api
```

---

## ✅ Lo que está COMPLETO

- ✅ Dependency Injection en todas las capas
- ✅ Repository pattern con interfaces
- ✅ Separación Domain/DAO/DTO
- ✅ CRUD completo de Actividades
- ✅ CRUD completo de Inscripciones
- ✅ GORM hooks para validaciones de negocio
- ✅ Vista MySQL con cupos calculados
- ✅ JWT authentication
- ✅ CORS middleware
- ✅ Health check endpoint
- ✅ Docker support
- ✅ Soft delete de inscripciones

---

## 📝 TODO: Pendientes para el equipo

### PRIORIDAD 1: RabbitMQ

**Archivo:** `internal/clients/rabbitmq_client.go`

```bash
# Copiar de:
cp ../clases2025-main/clase04-rabbitmq/internal/clients/rabbitmq_client.go \
   internal/clients/
```

**Modificar:** `internal/services/inscripciones.go`

Descomentar y agregar:
```go
// En Create():
if err := s.publisher.Publish(ctx, "inscription.created", createdInscripcion.ID); err != nil {
    log.Printf("Error publishing event: %v", err)
}

// En Deactivate():
if err := s.publisher.Publish(ctx, "inscription.deleted", inscripcionID); err != nil {
    log.Printf("Error publishing event: %v", err)
}
```

---

### PRIORIDAD 2: Validaciones HTTP Cross-Microservicio

**Modificar:** `internal/services/inscripciones.go`

Implementar:
1. **Validar usuario existe** (HTTP GET a `users-api:8080/users/:id`)
2. **Validar suscripción activa** (HTTP GET a `subscriptions-api:8081/subscriptions/active/:user_id`)
3. **Validar plan cubre actividad** (si la actividad requiere plan premium)

Ver ejemplos en `MIGRACION_COMPLETADA.md`

---

### PRIORIDAD 3: Implementar Sucursales CRUD

**Crear:**
- `internal/repository/sucursales_mysql.go` (copiar patrón de `users_mysql.go`)
- `internal/services/sucursales.go`
- `internal/controllers/sucursales.go`

**Modificar:** `cmd/api/main.go`

Agregar:
```go
sucursalesRepo := repository.NewMySQLSucursalesRepository(actividadesRepo.GetDB())
sucursalesService := services.NewSucursalesService(sucursalesRepo)
sucursalesController := controllers.NewSucursalesController(sucursalesService)

// Rutas públicas
router.GET("/sucursales", sucursalesController.List)
router.GET("/sucursales/:id", sucursalesController.GetByID)

// Rutas admin
adminOnly.POST("/sucursales", sucursalesController.Create)
adminOnly.PUT("/sucursales/:id", sucursalesController.Update)
adminOnly.DELETE("/sucursales/:id", sucursalesController.Delete)
```

---

### PRIORIDAD 4: Agregar campos nuevos

**Modificar:** `internal/dao/Actividad.go`

```go
RequierePlanPremium bool `gorm:"column:requiere_plan_premium;default:false"`
```

**Modificar:** `internal/dao/Inscripcion.go`

En `Create()`, asignar:
```go
inscripcionDAO.SuscripcionID = &activeSub.ID  // obtener de subscriptions-api
```

---

### PRIORIDAD 5: Tests

**Crear:**
- `internal/services/actividades_test.go`
- `internal/services/inscripciones_test.go`

---

## 🎓 Notas Técnicas

1. **Shared DB Connection**: Los repositorios de Actividades e Inscripciones comparten la misma conexión MySQL a través de `GetDB()`

2. **Horarios**: Se usan `time.Time` en DAO (para MySQL) y `string "HH:MM"` en Domain (para la API)

3. **PK de Inscripcion**: Se cambió de PK compuesta `(usuario_id, actividad_id)` a PK simple `id` + UNIQUE constraint para facilitar referencias futuras

4. **Hooks de GORM**: Se preservaron del código original. Son críticos para la validación de cupos

5. **Soft Delete**: Inscripciones usan `is_activa` en lugar de GORM soft delete para permitir reactivación

---

## 📞 Soporte

Si tienes dudas sobre el código migrado o cómo agregar las features nuevas, consulta:

- `MIGRACION_COMPLETADA.md` - TODOs detallados con ejemplos de código
- `../users-api/` - Template de referencia
- `../MICROSERVICE_TEMPLATE.md` - Guía paso a paso

---

✅ **Microservicio completamente funcional y listo para producción (con features básicas)**

🔜 **TODOs para agregar features avanzadas (RabbitMQ, validaciones HTTP, Sucursales)**
