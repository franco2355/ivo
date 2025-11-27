# Guía Completa de Testing - Gym Management System

## 📚 Tabla de Contenidos

1. [Estructura de Tests](#estructura-de-tests)
2. [Tipos de Tests](#tipos-de-tests)
3. [Ejecutar Tests](#ejecutar-tests)
4. [Escribir Nuevos Tests](#escribir-nuevos-tests)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)
7. [Cobertura de Tests](#cobertura-de-tests)

---

## 📁 Estructura de Tests

```
backend/
├── tests/
│   ├── unit/                           # Tests unitarios compartidos
│   │   ├── inscripciones_service_test.go
│   │   ├── plan_service_test.go
│   │   └── controllers_test.go
│   │
│   ├── integration/                    # Tests de integración
│   │   ├── user_registration_flow_test.go
│   │   ├── payment_workflow_test.go
│   │   ├── activity_capacity_test.go
│   │   ├── cash_payment_and_restrictions_test.go
│   │   ├── jwt_security_test.go
│   │   ├── plan_upgrade_test.go
│   │   ├── rate_limiting_test.go
│   │   ├── search_api_test.go
│   │   ├── solr_search_test.go
│   │   ├── subscription_cancellation_test.go
│   │   ├── subscription_expiration_test.go
│   │   └── unsubscribe_resubscribe_test.go
│   │
│   ├── e2e/                            # Tests end-to-end
│   │   └── complete_subscription_flow_test.go
│   │
│   ├── mocks/                          # Mocks reutilizables
│   ├── README.md                       # Docs de tests de integración
│   └── TEST_GUIDE.md                   # Esta guía
│
└── {service}/internal/services/        # Tests específicos de cada servicio
    ├── actividades_test.go
    ├── users_test.go
    ├── payment_service_test.go
    ├── subscription_service_test.go
    ├── plan_service_test.go
    ├── cache_service_test.go
    └── search_service_test.go
```

---

## 🎯 Tipos de Tests

### 1. Tests Unitarios (Unit Tests)

**Objetivo**: Probar lógica de negocio de forma aislada.

**Características**:
- ✅ No requieren servicios externos (DB, APIs, RabbitMQ)
- ✅ Usan mocks para todas las dependencias
- ✅ Muy rápidos (< 1 segundo cada uno)
- ✅ Alta cobertura de casos edge

**Ubicación**:
- `backend/tests/unit/` - Tests compartidos
- `backend/{service}/internal/services/*_test.go` - Tests del servicio

**Ejemplo**:
```go
func TestCreateUser_Success(t *testing.T) {
    // Arrange - Configurar mocks
    mockRepo := &MockUsersRepository{
        CreateFunc: func(ctx context.Context, user domain.User) (domain.User, error) {
            user.ID = 1
            return user, nil
        },
    }
    service := NewUsersService(mockRepo, "secret")

    // Act - Ejecutar acción
    user, err := service.Create(ctx, input)

    // Assert - Verificar resultado
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if user.ID != 1 {
        t.Errorf("Expected ID 1, got: %d", user.ID)
    }
}
```

**Ejecutar**:
```bash
# Todos los tests unitarios
go test ./backend/tests/unit/... ./backend/*/internal/services/... -v

# De un servicio específico
go test ./backend/users-api/internal/services/... -v

# Con cobertura
go test ./backend/users-api/internal/services/... -cover
```

---

### 2. Tests de Integración (Integration Tests)

**Objetivo**: Validar interacción entre componentes.

**Características**:
- ✅ Prueban endpoints HTTP reales
- ✅ Requieren servicios externos (MySQL, MongoDB, RabbitMQ, Solr)
- ✅ Verifican comunicación entre microservicios
- ✅ Validan eventos y mensajería
- ⚠️ Más lentos (2-10 segundos cada uno)

**Ubicación**: `backend/tests/integration/`

**Ejemplo**:
```go
func TestUserRegistrationFlow(t *testing.T) {
    baseURL := "http://localhost:8080"

    // Registrar usuario
    resp, err := http.Post(
        baseURL + "/users-api/register",
        "application/json",
        bytes.NewBuffer(payload),
    )

    // Verificar respuesta
    if resp.StatusCode != http.StatusCreated {
        t.Errorf("Expected 201, got: %d", resp.StatusCode)
    }
}
```

**Ejecutar**:
```bash
# ⚠️ IMPORTANTE: Levantar servicios primero
docker-compose up -d

# Ejecutar tests de integración
go test ./backend/tests/integration/... -v

# Ejecutar un test específico
go test ./backend/tests/integration/ -run TestUserRegistrationFlow -v
```

---

### 3. Tests End-to-End (E2E Tests)

**Objetivo**: Validar flujos completos de usuario.

**Características**:
- ✅ Prueban escenarios reales de usuario
- ✅ Involucran múltiples microservicios
- ✅ Verifican TODO el stack tecnológico
- ⚠️ Los más lentos (5-30 segundos cada uno)
- ⚠️ Más frágiles (dependen de más componentes)

**Ubicación**: `backend/tests/e2e/`

**Ejemplo - Flujo completo de suscripción**:
```go
func TestCompleteSubscriptionFlow(t *testing.T) {
    // 1. Registrar usuario
    // 2. Listar planes disponibles
    // 3. Crear suscripción
    // 4. Crear pago
    // 5. Aprobar pago
    // 6. Verificar suscripción activada
    // 7. Inscribirse en actividad
}
```

**Ejecutar**:
```bash
# ⚠️ IMPORTANTE: Sistema completo debe estar corriendo
docker-compose up -d

# Ejecutar tests E2E
go test ./backend/tests/e2e/... -v

# Con timeout extendido
go test ./backend/tests/e2e/... -timeout 10m -v
```

---

## 🚀 Ejecutar Tests

### Prerequisitos

#### Para Tests Unitarios
✅ Ninguno - no requieren servicios externos

#### Para Tests de Integración y E2E
```bash
# 1. Levantar todos los servicios
docker-compose up -d

# 2. Verificar que estén saludables
docker-compose ps

# Deberías ver todos con estado "Up":
# - users-api (8080)
# - subscriptions-api (8081)
# - activities-api (8082)
# - payments-api (8083)
# - search-api (8084)
# - mysql (3307)
# - mongodb (27017)
# - rabbitmq (5672, 15672)
# - solr (8983)
```

### Comandos Útiles

```bash
# ============================================
# TESTS UNITARIOS (rápidos, sin dependencias)
# ============================================

# Todos los tests unitarios
go test ./backend/tests/unit/... ./backend/*/internal/services/...

# De un servicio específico
go test ./backend/users-api/internal/services/...
go test ./backend/activities-api/internal/services/...
go test ./backend/payments-api/internal/services/...

# Con verbose
go test ./backend/tests/unit/... -v

# Con cobertura
go test ./backend/users-api/internal/services/... -cover

# Generar reporte HTML de cobertura
go test ./backend/users-api/internal/services/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# ============================================
# TESTS DE INTEGRACIÓN (requieren servicios)
# ============================================

# Todos los tests de integración
go test ./backend/tests/integration/... -v

# Un test específico
go test ./backend/tests/integration/ -run TestUserRegistrationFlow -v
go test ./backend/tests/integration/ -run TestPaymentCreationFlow -v

# Tests que coincidan con patrón
go test ./backend/tests/integration/ -run ".*Payment.*" -v

# ============================================
# TESTS E2E (requieren sistema completo)
# ============================================

# Todos los tests E2E
go test ./backend/tests/e2e/... -v

# Con timeout personalizado (tests lentos)
go test ./backend/tests/e2e/... -timeout 10m -v

# ============================================
# TODOS LOS TESTS
# ============================================

# Ejecutar TODO (puede tomar varios minutos)
go test ./backend/... -v

# Con paralelización (cuidado con tests que modifican estado)
go test ./backend/tests/unit/... -parallel 4

# Con timeout global
go test ./backend/... -timeout 15m -v
```

---

## ✍️ Escribir Nuevos Tests

### Convenciones de Nombres

#### Archivos
```
{component}_test.go             # Tests unitarios
{feature}_flow_test.go          # Tests de integración
complete_{flow}_test.go         # Tests E2E
```

#### Funciones
```go
// Tests unitarios
Test{Component}_{Scenario}
// Ejemplos:
TestCreateUser_Success
TestCreateUser_ValidationError
TestLogin_WrongPassword

// Tests de integración
Test{Feature}{Flow}
// Ejemplos:
TestUserRegistrationFlow
TestPaymentCreationFlow

// Tests E2E
TestComplete{Flow}
// Ejemplos:
TestCompleteSubscriptionFlow
TestCompleteActivityEnrollmentFlow
```

### Patrón AAA (Arrange-Act-Assert)

```go
func TestCreatePlan_Success(t *testing.T) {
    // ============================================
    // ARRANGE - Configurar el entorno de prueba
    // ============================================
    mockRepo := &MockPlanRepository{
        CreateFunc: func(ctx context.Context, plan Plan) (Plan, error) {
            plan.ID = 1
            return plan, nil
        },
    }
    service := NewPlanService(mockRepo)

    input := PlanCreate{
        Nombre: "Plan Mensual",
        Precio: 5000,
    }

    // ============================================
    // ACT - Ejecutar la acción a probar
    // ============================================
    result, err := service.Create(context.Background(), input)

    // ============================================
    // ASSERT - Verificar el resultado
    // ============================================
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }

    if result.ID != 1 {
        t.Errorf("Expected ID 1, got: %d", result.ID)
    }

    if result.Nombre != "Plan Mensual" {
        t.Errorf("Expected name 'Plan Mensual', got: %s", result.Nombre)
    }
}
```

### Table-Driven Tests

Útil para probar múltiples casos:

```go
func TestPasswordValidation(t *testing.T) {
    tests := []struct {
        name        string
        password    string
        expectError bool
        errorMsg    string
    }{
        {
            name:        "valid password",
            password:    "ValidPass123!",
            expectError: false,
        },
        {
            name:        "too short",
            password:    "Pass1!",
            expectError: true,
            errorMsg:    "password must be at least 8 characters",
        },
        {
            name:        "no uppercase",
            password:    "password123!",
            expectError: true,
            errorMsg:    "password must contain at least one uppercase letter",
        },
        {
            name:        "no number",
            password:    "Password!",
            expectError: true,
            errorMsg:    "password must contain at least one number",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validatePassword(tt.password)

            if tt.expectError {
                if err == nil {
                    t.Fatalf("Expected error, got nil")
                }
                if err.Error() != tt.errorMsg {
                    t.Errorf("Expected error '%s', got: '%s'", tt.errorMsg, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("Expected no error, got: %v", err)
                }
            }
        })
    }
}
```

### Subtests

Organizar tests relacionados:

```go
func TestUserManagement(t *testing.T) {
    t.Run("Create user", func(t *testing.T) {
        // Test crear usuario
    })

    t.Run("Get user by ID", func(t *testing.T) {
        // Test obtener usuario
    })

    t.Run("Update user", func(t *testing.T) {
        // Test actualizar usuario
    })

    t.Run("Delete user", func(t *testing.T) {
        // Test eliminar usuario
    })
}
```

---

## 🏆 Best Practices

### 1. Tests Independientes

```go
// ❌ MAL - Tests dependen uno del otro
func TestCreateUser(t *testing.T) {
    user = createUser() // Variable global
}

func TestUpdateUser(t *testing.T) {
    updateUser(user) // Depende de TestCreateUser
}

// ✅ BIEN - Cada test es independiente
func TestCreateUser(t *testing.T) {
    user := createUser()
    // Test completo aquí
}

func TestUpdateUser(t *testing.T) {
    user := createUser() // Crear su propio usuario
    updateUser(user)
    // Test completo aquí
}
```

### 2. Cleanup Apropiado

```go
func TestWithResources(t *testing.T) {
    // Crear recurso
    conn := openDBConnection()

    // Registrar cleanup
    t.Cleanup(func() {
        conn.Close()
    })

    // Usar el recurso
    // ...
}
```

### 3. Mensajes de Error Claros

```go
// ❌ MAL
if result != expected {
    t.Error("failed")
}

// ✅ BIEN
if result != expected {
    t.Errorf("Expected result %v, but got %v", expected, result)
}

// ✅ MEJOR
if result != expected {
    t.Errorf("User creation failed: Expected ID %d, got %d. "+
             "This might indicate an issue with the auto-increment sequence.",
             expected, result)
}
```

### 4. Skip Cuando Sea Apropiado

```go
func TestNewFeature(t *testing.T) {
    if !featureEnabled {
        t.Skip("Feature not yet implemented")
    }

    // Test logic
}

func TestIntegrationWithExternalAPI(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Test logic
}
```

### 5. Helper Functions

```go
// Crear helpers para operaciones comunes
func createTestUser(t *testing.T) *User {
    t.Helper() // Marca esta función como helper

    user := &User{
        Username: fmt.Sprintf("test_%d", time.Now().Unix()),
        Email:    fmt.Sprintf("test_%d@example.com", time.Now().Unix()),
    }

    return user
}

func TestSomething(t *testing.T) {
    user := createTestUser(t) // Usar helper
    // ...
}
```

---

## 🔧 Troubleshooting

### Error: "connection refused"

**Problema**: Los servicios no están corriendo.

**Solución**:
```bash
docker-compose up -d
docker-compose ps  # Verificar que todos estén "Up"
```

### Error: "database not found"

**Problema**: La base de datos no se inicializó correctamente.

**Solución**:
```bash
# Recrear todo
docker-compose down -v  # -v elimina volúmenes
docker-compose up -d
```

### Tests timeout

**Problema**: Los tests tardan demasiado.

**Solución**:
```bash
# Aumentar timeout
go test ./backend/tests/e2e/... -timeout 10m -v
```

### Error: "invalid credentials" en tests de integración

**Problema**: El usuario admin no existe o tiene contraseña diferente.

**Solución**:
```bash
# Verificar que existe admin/admin
# O actualizar las credenciales en el test
```

### Tests fallan aleatoriamente

**Problema**: Condiciones de carrera o estado compartido.

**Solución**:
- Asegurar que cada test es independiente
- No compartir variables globales
- Usar datos únicos (timestamps en nombres)

### RabbitMQ no procesa eventos

**Problema**: Los consumidores no están escuchando.

**Solución**:
```bash
# Ver logs de servicios
docker logs gym-subscriptions-api --tail 50
docker logs gym-payments-api --tail 50

# Verificar RabbitMQ
docker exec gym-rabbitmq rabbitmqctl list_queues
```

---

## 📊 Cobertura de Tests

### Por Servicio

| Servicio | Tests Unitarios | Tests Integración | Tests E2E | Cobertura |
|----------|----------------|-------------------|-----------|-----------|
| **Users API** | ✅ Alta (17 tests) | ✅ Media (3 flujos) | ✅ Alta | ~85% |
| **Activities API** | ✅ Alta (10 tests) | ✅ Media (4 flujos) | ✅ Alta | ~80% |
| **Subscriptions API** | ✅ Media (5 tests) | ✅ Alta (6 flujos) | ✅ Alta | ~75% |
| **Payments API** | ✅ Alta (20 tests) | ✅ Alta (3 flujos) | ✅ Alta | ~82% |
| **Search API** | ✅ Media (4 tests) | ✅ Alta (2 flujos) | ⚠️ Baja | ~65% |

### Áreas Bien Cubiertas ✅

- Validación de usuarios y autenticación (JWT)
- Creación y gestión de pagos (cash y MercadoPago)
- Flujos completos de suscripción
- Restricciones de planes
- Rate limiting
- Capacidad de actividades
- Búsqueda con Solr

### Áreas que Necesitan Más Tests ⚠️

- Webhooks de gateways de pago externos
- Escenarios de fallo de red
- Reembolsos parciales
- Manejo de concurrencia avanzada
- Tests de carga/stress

### Generar Reporte de Cobertura

```bash
# Cobertura global
go test ./backend/... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Cobertura por servicio
go test ./backend/users-api/... -coverprofile=users-coverage.out
go tool cover -html=users-coverage.out -o users-coverage.html

# Abrir en navegador
start users-coverage.html  # Windows
open users-coverage.html   # Mac
xdg-open users-coverage.html  # Linux
```

---

## 🎓 Recursos Adicionales

### Documentación de Go Testing
- [Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify (assertions library)](https://github.com/stretchr/testify)

### Ejemplos en el Proyecto
- Tests unitarios básicos: `backend/users-api/internal/services/users_test.go`
- Tests de integración: `backend/tests/integration/user_registration_flow_test.go`
- Tests E2E: `backend/tests/e2e/complete_subscription_flow_test.go`

---

**Última actualización**: 2025-11-27
**Mantenido por**: Backend Team
