# Tests de Integración - Gym Management System

Este directorio contiene tests de integración end-to-end para el sistema completo.

## 📋 Test Disponible

### `cash_payment_and_restrictions_test.go`

Test completo que verifica:
- ✅ Login de usuario y admin
- ✅ Creación de suscripción con Plan Básico (limitado)
- ✅ Creación de pago en efectivo
- ✅ Aprobación de pago por admin
- ✅ Activación automática de suscripción via RabbitMQ
- ✅ Bloqueo de inscripción a actividad NO permitida por el plan
- ✅ Inscripción exitosa a actividad permitida por el plan

## 🚀 Cómo ejecutar los tests

### Pre-requisitos

1. **Todos los servicios deben estar corriendo:**
```bash
docker-compose up -d
```

2. **Verificar que todos los servicios están saludables:**
```bash
docker-compose ps
```

Deberías ver:
- ✅ users-api (puerto 8080)
- ✅ subscriptions-api (puerto 8081)
- ✅ activities-api (puerto 8082)
- ✅ payments-api (puerto 8083)
- ✅ mysql (puerto 3307)
- ✅ mongodb (puerto 27017)
- ✅ rabbitmq (puerto 5672)

3. **El usuario `testuser` debe existir en la base de datos**

Si no existe, créalo desde el frontend o con este comando:
```bash
curl -X POST "http://localhost:8080/register" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Test",
    "apellido": "User",
    "username": "testuser",
    "email": "testuser@test.com",
    "password": "Test@123456",
    "dni": "12345678",
    "telefono": "1234567890"
  }'
```

### Ejecutar el test

```bash
# Desde el directorio backend/tests
cd backend/tests

# Ejecutar el test con output verbose
go test -v ./integration/

# O ejecutar el test específico
go test -v ./integration/ -run TestCashPaymentFlowWithPlanRestrictions
```

### Output esperado

```
=== RUN   TestCashPaymentFlowWithPlanRestrictions
🚀 Iniciando test de integración completo: Cash Payment + Plan Restrictions

📝 PASO 1: Login como usuario regular
✅ Usuario logueado - ID: 5, Token: eyJhbGciOiJIUzI1NiIsI...

📝 PASO 2: Login como admin
✅ Admin logueado - ID: 1, Token: eyJhbGciOiJIUzI1NiIsI...

📝 PASO 3: Crear suscripción con Plan Básico (limitado a yoga y spinning)
✅ Suscripción creada - ID: 6922a1b50fe2aa1a9f6314c5
✅ Suscripción en estado: pendiente_pago

📝 PASO 4: Crear pago en efectivo
✅ Pago en efectivo creado - ID: 6922a1d4c6155c16bc9b2cd8
✅ Pago en estado: pending, Gateway: cash

📝 PASO 5: Admin aprueba el pago en efectivo
✅ Pago aprobado por admin
✅ Pago actualizado a estado: completed

📝 PASO 6: Esperando activación automática de suscripción via RabbitMQ...
✅ Suscripción activada automáticamente! Estado: activa, PagoID: 6922a1d4c6155c16bc9b2cd8

📝 PASO 7: Verificar suscripción activa desde endpoint /active
✅ Suscripción activa verificada - Plan: Plan Básico

📝 PASO 8: Intentar inscribirse a Funcional (NO permitida por Plan Básico)
✅ Inscripción bloqueada correctamente! Error: tu plan 'Plan Básico' no incluye la categoría 'funcional'...

📝 PASO 9: Inscribirse a Yoga (permitida por Plan Básico)
✅ Inscripción exitosa a Yoga! UsuarioID: 5, ActividadID: 1

📝 PASO 10: Verificar lista de inscripciones del usuario
✅ Inscripción a Yoga encontrada en la lista (Total inscripciones: 1)

================================================================================
🎉 TEST COMPLETADO EXITOSAMENTE!
================================================================================
✅ Login de usuario y admin
✅ Creación de suscripción con Plan Básico (limitado)
✅ Creación de pago en efectivo
✅ Aprobación de pago por admin
✅ Activación automática de suscripción via RabbitMQ
✅ Bloqueo de inscripción a actividad NO permitida
✅ Inscripción exitosa a actividad permitida
================================================================================
--- PASS: TestCashPaymentFlowWithPlanRestrictions (5.23s)
PASS
ok      github.com/yourusername/gym-management/tests/integration        5.234s
```

## 🔧 Troubleshooting

### Error: "connection refused"
Los servicios no están corriendo. Ejecuta:
```bash
docker-compose up -d
docker-compose ps
```

### Error: "invalid credentials"
El usuario `testuser` no existe o la contraseña es incorrecta. Verifica o crea el usuario.

### Error: "plan no encontrado"
El Plan Básico con ID `6922595ffd37294158ce5f47` no existe en la base de datos.

Verifica con:
```bash
curl "http://localhost:8081/plans"
```

Y actualiza la constante `planBasicoID` en el test con el ID correcto.

### Test falla en PASO 6 (Activación automática)
RabbitMQ no está procesando los eventos correctamente. Verifica:
```bash
# Ver logs de subscriptions-api
docker logs gym-subscriptions-api --tail 50 | grep "📥\|payment.completed"

# Ver logs de payments-api
docker logs gym-payments-api --tail 50 | grep "📤\|payment.completed"
```

### Test falla en PASO 8 (Bloqueo de restricción)
Las restricciones de plan no están funcionando. Verifica:
```bash
# Ver logs de activities-api
docker logs gym-activities-api --tail 50
```

## 📊 Modificar el test

Para cambiar qué actividades se prueban, modifica las constantes al inicio del archivo:

```go
const (
    // IDs de actividades (ajustar según tu base de datos)
    yogaActivityID      = 1 // Permitida por Plan Básico
    spinningActivityID  = 2 // Permitida por Plan Básico
    funcionalActivityID = 3 // NO permitida por Plan Básico
)
```

## 🎯 Tests adicionales recomendados

Este test puede extenderse para verificar:
- [ ] Pagos con MercadoPago (auto-aprobación)
- [ ] Refunds de pagos
- [ ] Plan Premium (acceso completo a todas las actividades)
- [ ] Desinscripción de actividades
- [ ] Expiración de suscripciones
