# Users API - Gimnasio Management System

Microservicio de gestión de usuarios con autenticación JWT.

## Características

- ✅ Registro de usuarios
- ✅ Login con JWT
- ✅ Validaciones de email y password strength
- ✅ Endpoint para validar existencia de usuarios (usado por otros microservicios)
- ✅ Roles (normal, admin)
- ✅ Patrón Repository con interfaces
- ✅ Separación DTO/DAO/Domain
- ✅ Dependency Injection
- ✅ MySQL con GORM
- ✅ Docker support

## Tecnologías

- Go 1.22
- Gin (web framework)
- GORM (ORM)
- MySQL 8.0
- JWT (autenticación)
- Docker

## Estructura del Proyecto

```
users-api/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuración
│   ├── domain/
│   │   └── user.go                # Domain models
│   ├── dao/
│   │   └── User.go                # DAO models (MySQL)
│   ├── repository/
│   │   └── users_mysql.go         # Repository implementation
│   ├── services/
│   │   └── users.go               # Business logic
│   ├── controllers/
│   │   └── users.go               # HTTP handlers
│   └── middleware/
│       ├── cors.go
│       └── jwt.go
├── .env.example
├── Dockerfile
├── go.mod
└── README.md
```

## Instalación y Ejecución

### Con Docker (recomendado)

```bash
# Desde la raíz del proyecto
docker-compose up --build users-api
```

### Local

1. **Crear archivo .env**

```bash
cp .env.example .env
```

Editar `.env` con tus credenciales de MySQL.

2. **Instalar dependencias**

```bash
go mod download
```

3. **Ejecutar**

```bash
go run cmd/api/main.go
```

La API estará disponible en `http://localhost:8080`

## Endpoints

### Públicos (sin autenticación)

#### POST /register

Registra un nuevo usuario.

**Request:**
```json
{
  "nombre": "Juan",
  "apellido": "Pérez",
  "username": "juanperez",
  "email": "juan@example.com",
  "password": "Password123",
  "sucursal_origen_id": 1
}
```

**Validaciones de password:**
- Mínimo 8 caracteres
- Al menos una mayúscula
- Al menos una minúscula
- Al menos un número

**Response 201:**
```json
{
  "user": {
    "id": 1,
    "nombre": "Juan",
    "apellido": "Pérez",
    "username": "juanperez",
    "email": "juan@example.com",
    "is_admin": false,
    "sucursal_origen_id": 1,
    "fecha_registro": "2025-01-19T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### POST /login

Autentica un usuario.

**Request:**
```json
{
  "username_or_email": "juanperez",
  "password": "Password123"
}
```

**Response 200:**
```json
{
  "user": {
    "id": 1,
    "nombre": "Juan",
    "apellido": "Pérez",
    "username": "juanperez",
    "email": "juan@example.com",
    "is_admin": false
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Protegidos (requieren JWT)

Incluir header: `Authorization: Bearer <token>`

#### GET /users/:id

Obtiene un usuario por ID. Usado por otros microservicios para validar existencia.

**Response 200:**
```json
{
  "id": 1,
  "nombre": "Juan",
  "apellido": "Pérez",
  "username": "juanperez",
  "email": "juan@example.com",
  "is_admin": false,
  "sucursal_origen_id": 1,
  "fecha_registro": "2025-01-19T10:00:00Z"
}
```

#### GET /users

Lista todos los usuarios. **Solo admin**.

**Response 200:**
```json
{
  "users": [
    {
      "id": 1,
      "nombre": "Juan",
      "apellido": "Pérez",
      "username": "juanperez",
      "email": "juan@example.com",
      "is_admin": false
    }
  ],
  "count": 1
}
```

### Health Check

#### GET /healthz

Verifica el estado del servicio.

**Response 200:**
```json
{
  "status": "ok",
  "service": "users-api",
  "version": "1.0.0"
}
```

## Variables de Entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `PORT` | Puerto del servidor | `8080` |
| `DB_USER` | Usuario de MySQL | `root` |
| `DB_PASS` | Contraseña de MySQL | `` |
| `DB_HOST` | Host de MySQL | `localhost` |
| `DB_PORT` | Puerto de MySQL | `3306` |
| `DB_SCHEMA` | Base de datos | `proyecto_integrador` |
| `JWT_SECRET` | Secret para JWT | `my-secret-key` |

## Arquitectura

### Patrón de Capas

```
HTTP Request → Controller → Service → Repository → MySQL
                    ↓
               Validaciones
                    ↓
             Generación JWT
                    ↓
                Response
```

### Dependency Injection

```go
// En main.go
repo := repository.NewMySQLUsersRepository(cfg.MySQL)
service := services.NewUsersService(repo, cfg.JWT.Secret)
controller := controllers.NewUsersController(service)
```

### Separación de Modelos

- **Domain (`domain.User`)**: Modelo de negocio, independiente de BD
- **DAO (`dao.User`)**: Modelo de BD con tags de GORM
- **DTO (`domain.UserRegister`, `domain.UserLogin`)**: Modelos de request/response

## Testing

```bash
# Ejecutar tests (cuando estén implementados)
go test ./...
```

## Notas de Desarrollo

- **Soft Delete**: Los usuarios eliminados no se borran físicamente (GORM soft delete)
- **Password Hashing**: SHA256 (matching con proyecto original)
- **JWT Expiration**: 30 minutos
- **CORS**: Habilitado para todos los orígenes (configurar en producción)

## Integración con Otros Microservicios

Otros microservicios pueden validar la existencia de un usuario:

```go
// Ejemplo desde subscriptions-api
resp, err := http.Get("http://users-api:8080/users/123")
if err != nil {
    return errors.New("error validating user")
}
defer resp.Body.Close()

if resp.StatusCode == 404 {
    return errors.New("user not found")
}
```

## Licencia

MIT
