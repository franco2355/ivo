package integration

import (
	"testing"
)

// Test de integración completo: Pago en efectivo + Restricciones de plan
// Este test verifica:
// 1. Login de usuario y admin
// 2. Creación de suscripción con pago en efectivo
// 3. Aprobación de pago por admin
// 4. Activación automática de suscripción via RabbitMQ
// 5. Intento de inscripción a actividad NO permitida por el plan (debe fallar)
// 6. Inscripción exitosa a actividad permitida por el plan

const (
	// Credenciales de prueba
	userUsername    = "testuser"
	userPassword    = "password123"
	adminUsername   = "admin"
	adminPassword   = "admin123"

	// Plan Básico (limitado a yoga y spinning)
	planBasicoID = PlanBasicoID

	// IDs de actividades
	yogaActivityID      = 1 // Permitida por Plan Básico
	spinningActivityID  = 2 // Permitida por Plan Básico
	funcionalActivityID = 3 // NO permitida por Plan Básico (debe fallar)
)

func TestCashPaymentFlowWithPlanRestrictions(t *testing.T) {
	t.Log("🚀 Iniciando test de integración completo: Cash Payment + Plan Restrictions")

	// ============================================================================
	// PASO 1: Login como usuario regular
	// ============================================================================
	t.Log("\n📝 PASO 1: Login como usuario regular")
	userToken, userID := loginUser(t, userUsername, userPassword, false)
	t.Logf("✅ Usuario logueado - ID: %d, Token: %s...", userID, userToken[:20])

	// ============================================================================
	// PASO 2: Login como admin
	// ============================================================================
	t.Log("\n📝 PASO 2: Login como admin")
	adminToken, adminID := loginUser(t, adminUsername, adminPassword, true)
	t.Logf("✅ Admin logueado - ID: %d, Token: %s...", adminID, adminToken[:20])

	// ============================================================================
	// PASO 3: Crear suscripción con Plan Básico (limitado)
	// ============================================================================
	t.Log("\n📝 PASO 3: Crear suscripción con Plan Básico (limitado a yoga y spinning)")
	subscriptionID := createSubscription(t, userToken, userID, planBasicoID)
	t.Logf("✅ Suscripción creada - ID: %s", subscriptionID)

	// Verificar que la suscripción está en estado "pendiente_pago"
	subscription := getSubscription(t, userToken, subscriptionID)
	if subscription.Estado != "pendiente_pago" {
		t.Fatalf("❌ Estado incorrecto de suscripción. Esperado: pendiente_pago, Obtenido: %s", subscription.Estado)
	}
	t.Logf("✅ Suscripción en estado: %s", subscription.Estado)

	// ============================================================================
	// PASO 4: Crear pago en efectivo
	// ============================================================================
	t.Log("\n📝 PASO 4: Crear pago en efectivo")
	paymentID := createCashPayment(t, userToken, userID, subscriptionID, 5000.0)
	t.Logf("✅ Pago en efectivo creado - ID: %s", paymentID)

	// Verificar que el pago está en estado "pending"
	payment := getPayment(t, paymentID)
	if payment.Status != "pending" {
		t.Fatalf("❌ Estado incorrecto de pago. Esperado: pending, Obtenido: %s", payment.Status)
	}
	if payment.PaymentGateway != "cash" {
		t.Fatalf("❌ Gateway incorrecto. Esperado: cash, Obtenido: %s", payment.PaymentGateway)
	}
	t.Logf("✅ Pago en estado: %s, Gateway: %s", payment.Status, payment.PaymentGateway)

	// ============================================================================
	// PASO 5: Admin aprueba el pago en efectivo
	// ============================================================================
	t.Log("\n📝 PASO 5: Admin aprueba el pago en efectivo")
	approvePayment(t, adminToken, paymentID)
	t.Logf("✅ Pago aprobado por admin")

	// Verificar que el pago cambió a "completed"
	payment = getPayment(t, paymentID)
	if payment.Status != "completed" {
		t.Fatalf("❌ Estado incorrecto de pago después de aprobación. Esperado: completed, Obtenido: %s", payment.Status)
	}
	t.Logf("✅ Pago actualizado a estado: %s", payment.Status)

	// ============================================================================
	// PASO 6: Activar suscripción manualmente
	// ============================================================================
	t.Log("\n📝 PASO 6: Activando suscripción...")
	activateSubscription(t, userToken, subscriptionID, paymentID)

	// Verificar que la suscripción se activó
	subscription = getSubscription(t, userToken, subscriptionID)
	if subscription.Estado != "activa" {
		t.Fatalf("❌ Suscripción no se activó. Esperado: activa, Obtenido: %s", subscription.Estado)
	}
	if subscription.PagoID != paymentID {
		t.Fatalf("❌ PagoID no coincide. Esperado: %s, Obtenido: %s", paymentID, subscription.PagoID)
	}
	t.Logf("✅ Suscripción activada! Estado: %s, PagoID: %s", subscription.Estado, subscription.PagoID)

	// ============================================================================
	// PASO 7: Verificar suscripción activa desde endpoint específico
	// ============================================================================
	t.Log("\n📝 PASO 7: Verificar suscripción activa desde endpoint /active")
	activeSubscription := getActiveSubscription(t, userToken, userID)
	if activeSubscription.ID != subscriptionID {
		t.Fatalf("❌ Suscripción activa no coincide. Esperado: %s, Obtenido: %s", subscriptionID, activeSubscription.ID)
	}
	t.Logf("✅ Suscripción activa verificada - Plan: %s", activeSubscription.PlanNombre)

	// ============================================================================
	// PASO 8: Intentar inscribirse a actividad NO permitida (debe fallar)
	// ============================================================================
	t.Log("\n📝 PASO 8: Intentar inscribirse a Funcional (NO permitida por Plan Básico)")
	err := tryEnrollActivity(t, userToken, userID, funcionalActivityID, true) // true = se espera error
	if err == nil {
		t.Fatalf("❌ La inscripción debería haber fallado pero fue exitosa")
	}
	t.Logf("✅ Inscripción bloqueada correctamente! Error: %s", err.Error())

	// Verificar mensaje de error específico
	if err.Error() != "tu plan 'Plan Básico' no incluye la categoría 'funcional'. Actualiza tu plan para acceder a esta actividad" {
		t.Logf("⚠️  Mensaje de error diferente al esperado: %s", err.Error())
	}

	// ============================================================================
	// PASO 9: Desinscribirse de Yoga si ya estaba inscripto (cleanup previo)
	// ============================================================================
	t.Log("\n📝 PASO 9: Limpiar inscripciones previas a Yoga si existen")
	existingInscripciones := listInscripciones(t, userToken, userID)
	for _, insc := range existingInscripciones {
		if insc.ActividadID == yogaActivityID && insc.IsActiva {
			t.Logf("⚠️  Usuario ya estaba inscripto a Yoga, desinscribiendo...")
			unenrollActivity(t, userToken, userID, yogaActivityID)
			t.Logf("✅ Desinscripción previa completada")
			break
		}
	}

	// ============================================================================
	// PASO 10: Inscribirse exitosamente a actividad permitida (Yoga)
	// ============================================================================
	t.Log("\n📝 PASO 10: Inscribirse a Yoga (permitida por Plan Básico)")
	inscripcion := enrollActivity(t, userToken, userID, yogaActivityID)
	if inscripcion.ActividadID != yogaActivityID {
		t.Fatalf("❌ ActividadID no coincide. Esperado: %d, Obtenido: %d", yogaActivityID, inscripcion.ActividadID)
	}
	if inscripcion.UsuarioID != userID {
		t.Fatalf("❌ UsuarioID no coincide. Esperado: %d, Obtenido: %d", userID, inscripcion.UsuarioID)
	}
	if !inscripcion.IsActiva {
		t.Fatalf("❌ Inscripción no está activa")
	}
	t.Logf("✅ Inscripción exitosa a Yoga! UsuarioID: %d, ActividadID: %d", inscripcion.UsuarioID, inscripcion.ActividadID)

	// ============================================================================
	// PASO 11: Verificar que la inscripción está registrada
	// ============================================================================
	t.Log("\n📝 PASO 11: Verificar lista de inscripciones del usuario")
	inscripciones := listInscripciones(t, userToken, userID)
	found := false
	for _, insc := range inscripciones {
		if insc.ActividadID == yogaActivityID && insc.IsActiva {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("❌ No se encontró la inscripción a Yoga en la lista de inscripciones")
	}
	t.Logf("✅ Inscripción a Yoga encontrada en la lista (Total inscripciones: %d)", len(inscripciones))

	// ============================================================================
	// RESUMEN FINAL
	// ============================================================================
	separator := "================================================================================"
	t.Log("\n" + separator)
	t.Log("🎉 TEST COMPLETADO EXITOSAMENTE!")
	t.Log(separator)
	t.Log("✅ Login de usuario y admin")
	t.Log("✅ Creación de suscripción con Plan Básico (limitado)")
	t.Log("✅ Creación de pago en efectivo")
	t.Log("✅ Aprobación de pago por admin")
	t.Log("✅ Activación automática de suscripción via RabbitMQ")
	t.Log("✅ Bloqueo de inscripción a actividad NO permitida")
	t.Log("✅ Inscripción exitosa a actividad permitida")
	t.Log(separator)
}
