# Autenticación JWT - Subscriptions API

## 📋 Resumen

Este microservicio utiliza **JWT (JSON Web Tokens)** para autenticar y autorizar usuarios.

## 🔑 Configuración

### Variables de Entorno

```env
JWT_SECRET=your-secret-key-change-in-production
```

**IMPORTANTE**: En producción, usa un secret complejo y seguro (mínimo 32 caracteres).

## 🛡️ Endpoints Protegidos

### Rutas Públicas (sin autenticación)
- `GET /healthz` - Health check
- `GET /plans` - Listar planes
- `GET /plans/:id` - Obtener plan por ID

### Rutas Protegidas (requieren autenticación)
- `POST /subscriptions` - Crear suscripción
- `GET /subscriptions/:id` - Obtener suscripción
- `GET /subscriptions/active/:user_id` - Obtener suscripción activa
- `PATCH /subscriptions/:id/status` - Actualizar estado
- `DELETE /subscriptions/:id` - Cancelar suscripción

### Rutas Admin (requieren rol "admin")
- `POST /plans` - Crear plan

## 📝 Estructura del JWT

### Claims del Token

```json
{
  "user_id": "123",
  "username": "john_doe",
  "role": "user",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Roles Disponibles

- **user**: Usuario regular (puede gestionar sus propias suscripciones)
- **admin**: Administrador (puede crear planes y gestionar todo)

## 🔧 Uso

### 1. Obtener Token

Los tokens JWT deben ser generados por el microservicio `users-api` al hacer login.

**Ejemplo de login en users-api:**
```bash
POST http://localhost:8080/auth/login
Content-Type: application/json

{
  "username": "john_doe",
  "password": "securepassword"
}
```

**Respuesta:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123",
    "username": "john_doe",
    "role": "user"
  }
}
```

### 2. Usar Token en Requests

Incluye el token en el header `Authorization`:

```bash
GET http://localhost:8081/subscriptions/673f4e20bcf86cd799439022
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 3. Ejemplos con cURL

**Crear suscripción (requiere autenticación):**
```bash
curl -X POST http://localhost:8081/subscriptions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "usuario_id": "123",
    "plan_id": "673f4e20bcf86cd799439011",
    "metodo_pago": "credit_card"
  }'
```

**Crear plan (requiere rol admin):**
```bash
curl -X POST http://localhost:8081/plans \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Premium",
    "precio_mensual": 100.00,
    "tipo_acceso": "completo",
    "duracion_dias": 30,
    "activo": true
  }'
```

## 🚨 Respuestas de Error

### 401 Unauthorized

**Sin token:**
```json
{
  "error": "Token de autorización requerido"
}
```

**Token inválido:**
```json
{
  "error": "Token inválido o expirado"
}
```

**Formato incorrecto:**
```json
{
  "error": "Formato de token inválido. Use: Bearer {token}"
}
```

### 403 Forbidden

**Permisos insuficientes:**
```json
{
  "error": "Acceso denegado: permisos insuficientes"
}
```

## 🧪 Testing con Token

### Generar Token de Prueba (solo desarrollo)

Puedes usar [jwt.io](https://jwt.io) para generar tokens de prueba:

**Payload:**
```json
{
  "user_id": "1",
  "username": "test_user",
  "role": "user",
  "exp": 9999999999,
  "iat": 1234567890
}
```

**Secret:** Usa el mismo valor de `JWT_SECRET` del .env

### Token de Admin de Prueba

**Payload:**
```json
{
  "user_id": "1",
  "username": "admin",
  "role": "admin",
  "exp": 9999999999,
  "iat": 1234567890
}
```

## 🔒 Seguridad

### Mejores Prácticas

1. **Usa HTTPS en producción** - Los tokens deben transmitirse por canales seguros
2. **Tokens de corta duración** - Establece `exp` (expiration) a 1-24 horas
3. **Rota el JWT_SECRET** - Cambia el secret periódicamente
4. **Valida todos los claims** - Verifica `exp`, `iat`, `user_id`, etc.
5. **No almacenes información sensible** - Los JWT son decodificables
6. **Implementa refresh tokens** - Para renovar tokens sin re-login

### Validación Automática

El middleware `JWTAuth` valida automáticamente:
- ✅ Firma del token (HMAC SHA256)
- ✅ Expiración (`exp` claim)
- ✅ Formato del token
- ✅ Integridad de los claims

## 🔄 Integración con users-api

Este microservicio **NO genera tokens**, solo los **valida**.

**Flujo completo:**

1. Usuario hace login en `users-api`
2. `users-api` genera y retorna JWT
3. Usuario usa JWT en requests a `subscriptions-api`
4. `subscriptions-api` valida JWT usando el mismo `JWT_SECRET`

**Requisitos:**
- `users-api` y `subscriptions-api` deben usar el **mismo JWT_SECRET**
- Los claims deben incluir: `user_id`, `username`, `role`

## 📚 Helpers Disponibles

### En Controllers

```go
import "github.com/yourusername/gym-management/subscriptions-api/internal/middleware"

func (c *SubscriptionController) GetSubscription(ctx *gin.Context) {
    // Obtener user_id del token
    userID, err := middleware.GetUserIDFromContext(ctx)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
        return
    }

    // Usar userID para validar permisos
    // ...
}
```

### Middleware Opcional

Para rutas donde la autenticación es **opcional**:

```go
router.GET("/public-endpoint", middleware.OptionalAuth(cfg.JWTSecret), handler)
```

## 🐛 Troubleshooting

### "Token inválido o expirado"

- Verifica que el token no haya expirado (`exp` claim)
- Asegúrate de usar el mismo `JWT_SECRET` en todos los servicios
- Verifica el formato: `Bearer {token}`

### "Acceso denegado: permisos insuficientes"

- Verifica que el usuario tenga el rol correcto
- Para crear planes se requiere rol `admin`

### "Token de autorización requerido"

- Asegúrate de incluir el header `Authorization`
- Formato correcto: `Authorization: Bearer YOUR_TOKEN`
