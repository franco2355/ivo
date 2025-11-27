package integration

import (
	"testing"
)

// TestPlanUpgradeFlow valida el flujo completo de upgrade de plan
func TestPlanUpgradeFlow(t *testing.T) {
	t.Log("🚀 Iniciando test de integración: Plan Upgrade Flow")

	// ==================== PASO 1: Setup ====================
	t.Log("\n📝 PASO 1: Registrar usuario y admin")
	userToken, userID, userData := registerUser(t)
	t.Logf("✅ Usuario registrado - ID: %d, Username: %s", userID, userData.Username)

	t.Log("\n📝 PASO 2: Login como admin")
	adminToken, adminID := login(t, "admin", "admin123")
	t.Logf("✅ Admin logueado - ID: %d, Token: %.20s...", adminID, adminToken)

	// ==================== PASO 3: Crear suscripción con Plan Básico ====================
	t.Log("\n📝 PASO 3: Crear suscripción con Plan Básico (limitado)")
	subscriptionID := createSubscription(t, userToken, userID, PlanBasicoID) // Plan Básico
	t.Logf("✅ Suscripción creada - ID: %s", subscriptionID)

	// Verificar estado inicial
	subscription := getSubscription(t, userToken, subscriptionID)
	if subscription.Estado != "pendiente_pago" {
		t.Fatalf("❌ Estado incorrecto. Esperado: pendiente_pago, Obtenido: %s", subscription.Estado)
	}
	t.Logf("✅ Suscripción en estado: %s", subscription.Estado)

	// ==================== PASO 4: Crear y aprobar pago para Plan Básico ====================
	t.Log("\n📝 PASO 4: Crear pago en efectivo para Plan Básico")
	paymentID := createCashPayment(t, adminToken, userID, subscriptionID, 1000.0)
	t.Logf("✅ Pago creado - ID: %s", paymentID)

	t.Log("\n📝 PASO 5: Admin aprueba el pago y activa suscripción")
	updatePaymentStatus(t, adminToken, paymentID, "completed")
	activateSubscription(t, userToken, subscriptionID, paymentID)

	subscription = getSubscription(t, userToken, subscriptionID)
	if subscription.Estado != "activa" {
		t.Fatalf("❌ Suscripción no se activó. Estado: %s", subscription.Estado)
	}
	t.Logf("✅ Suscripción activada! Estado: %s", subscription.Estado)

	// ==================== PASO 7: Crear actividad Funcional e intentar inscribirse (debe fallar con Plan Básico) ====================
	t.Log("\n📝 PASO 7: Crear actividad Funcional e intentar inscribirse (NO permitida por Plan Básico)")
	funcionalActivity := createActivity(t, adminToken, "Funcional Test", "funcional", 1)
	funcionalID := int(funcionalActivity["id"].(float64))
	t.Logf("✅ Actividad Funcional creada - ID: %d", funcionalID)

	// Intentar inscribirse (debería fallar con Plan Básico)
	err := tryEnrollActivity(t, userToken, int(userID), funcionalID, true) // true = se espera error
	if err == nil {
		t.Fatal("❌ ERROR: La inscripción a Funcional debería haber sido bloqueada!")
	}
	t.Logf("✅ Inscripción bloqueada correctamente! Error: %v", err.Error())

	// ==================== PASO 8: Crear suscripción con Plan Premium ====================
	t.Log("\n📝 PASO 8: Upgrade a Plan Premium (acceso completo)")
	premiumSubID := createSubscription(t, userToken, userID, PlanPremiumID) // Plan Premium
	t.Logf("✅ Suscripción Premium creada - ID: %s", premiumSubID)

	// ==================== PASO 9: Pagar Plan Premium ====================
	t.Log("\n📝 PASO 9: Crear y aprobar pago para Plan Premium")
	premiumPaymentID := createCashPayment(t, adminToken, userID, premiumSubID, 3000.0)
	updatePaymentStatus(t, adminToken, premiumPaymentID, "completed")
	activateSubscription(t, userToken, premiumSubID, premiumPaymentID)

	premiumSub := getSubscription(t, userToken, premiumSubID)
	if premiumSub.Estado != "activa" {
		t.Fatalf("❌ Suscripción Premium no se activó. Estado: %s", premiumSub.Estado)
	}
	t.Log("✅ Suscripción Premium activada!")

	// ==================== PASO 10: Inscribirse a Funcional con Plan Premium (ahora debe funcionar) ====================
	t.Log("\n📝 PASO 10: Inscribirse a Funcional con Plan Premium")

	inscripcion := enrollActivity(t, userToken, int(userID), funcionalID)
	t.Logf("✅ Inscripción exitosa a Funcional con Plan Premium! ID: %d_%d", inscripcion.UsuarioID, inscripcion.ActividadID)

	// ==================== RESUMEN ====================
	t.Log("\n================================================================================")
	t.Log("🎉 TEST DE UPGRADE DE PLAN COMPLETADO EXITOSAMENTE!")
	t.Log("================================================================================")
	t.Log("✅ Suscripción con Plan Básico creada y activada")
	t.Log("✅ Inscripción a actividad premium bloqueada con Plan Básico")
	t.Log("✅ Upgrade a Plan Premium exitoso")
	t.Log("✅ Inscripción a actividad premium exitosa con Plan Premium")
	t.Log("================================================================================")
}
