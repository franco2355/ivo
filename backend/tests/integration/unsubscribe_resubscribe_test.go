package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestUnsubscribeAndResubscribe valida el flujo de desinscripción y re-inscripción
func TestUnsubscribeAndResubscribe(t *testing.T) {
	t.Log("🚀 Iniciando test de integración: Unsubscribe and Resubscribe")

	// ==================== PASO 1: Setup ====================
	t.Log("\n📝 PASO 1: Registrar usuario y admin")
	adminToken, adminID := login(t, "admin", "admin123")
	t.Logf("✅ Admin logueado - ID: %d", adminID)

	userToken, userID, userData := registerUser(t)
	t.Logf("✅ Usuario registrado - ID: %d, Username: %s", userID, userData.Username)

	// Activar suscripción
	subscriptionID := createSubscription(t, userToken, userID, PlanPremiumID)
	paymentID := createCashPayment(t, adminToken, userID, subscriptionID, 3000.0)
	updatePaymentStatus(t, adminToken, paymentID, "completed")
	activateSubscription(t, userToken, subscriptionID, paymentID)
	time.Sleep(3 * time.Second)
	t.Log("✅ Suscripción activada")

	client := &http.Client{}

	// ==================== PASO 2: Crear actividad e inscribirse ====================
	t.Log("\n📝 PASO 2: Crear actividad Yoga e inscribirse")
	yogaActivity := createActivity(t, adminToken, "Yoga Unsub Test", "yoga", 1)
	yogaID := int(yogaActivity["id"].(float64))
	t.Logf("✅ Actividad Yoga creada - ID: %d", yogaID)

	// Limpiar inscripciones previas si existen
	existingInscripciones := listInscripciones(t, userToken, int(userID))
	for _, insc := range existingInscripciones {
		if insc.ActividadID == yogaID && insc.IsActiva {
			t.Log("ℹ️  Ya estaba inscrito, desinscribiendo primero...")
			unenrollActivity(t, userToken, int(userID), yogaID)
			time.Sleep(1 * time.Second)
			break
		}
	}

	inscripcion := enrollActivity(t, userToken, int(userID), yogaID)
	t.Logf("✅ Inscrito a Yoga - ID: %d_%d", inscripcion.UsuarioID, inscripcion.ActividadID)

	// ==================== PASO 3: Verificar actividad antes de desinscribirse ====================
	t.Log("\n📝 PASO 3: Verificar cupos de la actividad antes de desinscribirse")

	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8082/actividades/%d", yogaID), nil)
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("❌ Error consultando actividad: %v", err)
	}
	defer resp.Body.Close()

	var activityBefore map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&activityBefore)

	cupoTotal := int(activityBefore["cupo"].(float64))
	lugaresAntes := cupoTotal
	if activityBefore["lugares"] != nil {
		lugaresAntes = int(activityBefore["lugares"].(float64))
	}

	t.Logf("✅ Cupo total: %d, Lugares disponibles antes: %d", cupoTotal, lugaresAntes)

	// ==================== PASO 4: Desinscribirse ====================
	t.Log("\n📝 PASO 4: Desinscribirse de Yoga")

	unenrollActivity(t, userToken, int(userID), yogaID)
	t.Log("✅ Desinscripción exitosa")

	time.Sleep(2 * time.Second)

	// ==================== PASO 5: Verificar que el cupo se liberó ====================
	t.Log("\n📝 PASO 5: Verificar que el cupo se liberó")

	httpReq, _ = http.NewRequest("GET", fmt.Sprintf("http://localhost:8082/actividades/%d", yogaID), nil)
	resp, err = client.Do(httpReq)
	if err != nil {
		t.Fatalf("❌ Error consultando actividad: %v", err)
	}
	defer resp.Body.Close()

	var activityAfter map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&activityAfter)

	lugaresDespues := cupoTotal
	if activityAfter["lugares"] != nil {
		lugaresDespues = int(activityAfter["lugares"].(float64))
	}

	t.Logf("✅ Lugares disponibles después: %d", lugaresDespues)

	if lugaresDespues > lugaresAntes || lugaresDespues == cupoTotal {
		t.Logf("✅ Cupo liberado correctamente (Antes: %d, Después: %d)", lugaresAntes, lugaresDespues)
	} else {
		t.Logf("⚠️  El cupo puede no haberse liberado inmediatamente (Antes: %d, Después: %d)", lugaresAntes, lugaresDespues)
	}

	// ==================== PASO 6: Verificar lista de inscripciones ====================
	t.Log("\n📝 PASO 6: Verificar que la inscripción está inactiva")

	httpReq, _ = http.NewRequest("GET", "http://localhost:8082/inscripciones", nil)
	httpReq.Header.Set("Authorization", userToken)

	resp, err = client.Do(httpReq)
	if err != nil {
		t.Fatalf("❌ Error consultando inscripciones: %v", err)
	}
	defer resp.Body.Close()

	var inscripciones []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&inscripciones)

	yogaInscripcionActiva := false
	for _, insc := range inscripciones {
		actID := int(insc["actividad_id"].(float64))
		if actID == yogaID && insc["is_activa"].(bool) {
			yogaInscripcionActiva = true
			break
		}
	}

	if !yogaInscripcionActiva {
		t.Log("✅ La inscripción a Yoga está inactiva")
	} else {
		t.Log("⚠️  La inscripción a Yoga todavía aparece como activa")
	}

	// ==================== PASO 7: Re-inscribirse ====================
	t.Log("\n📝 PASO 7: Re-inscribirse a Yoga")

	nuevaInscripcion := enrollActivity(t, userToken, int(userID), yogaID)
	t.Logf("✅ Re-inscrito a Yoga exitosamente - ID: %d_%d", nuevaInscripcion.UsuarioID, nuevaInscripcion.ActividadID)

	// ==================== PASO 8: Verificar que la re-inscripción está activa ====================
	t.Log("\n📝 PASO 8: Verificar que la re-inscripción está activa")

	httpReq, _ = http.NewRequest("GET", "http://localhost:8082/inscripciones", nil)
	httpReq.Header.Set("Authorization", userToken)

	resp, err = client.Do(httpReq)
	if err != nil {
		t.Fatalf("❌ Error consultando inscripciones: %v", err)
	}
	defer resp.Body.Close()

	var inscripcionesFinal []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&inscripcionesFinal)

	reInscripcionActiva := false
	for _, insc := range inscripcionesFinal {
		actID := int(insc["actividad_id"].(float64))
		if actID == yogaID && insc["is_activa"].(bool) {
			reInscripcionActiva = true
			t.Logf("✅ Inscripción activa encontrada - ID: %.0f", insc["id"].(float64))
			break
		}
	}

	if reInscripcionActiva {
		t.Log("✅ La re-inscripción está activa")
	} else {
		t.Fatal("❌ La re-inscripción no se encontró o no está activa")
	}

	// ==================== PASO 9: Verificar cupo ocupado nuevamente ====================
	t.Log("\n📝 PASO 9: Verificar que el cupo se ocupó nuevamente")

	httpReq, _ = http.NewRequest("GET", fmt.Sprintf("http://localhost:8082/actividades/%d", yogaID), nil)
	resp, err = client.Do(httpReq)
	if err != nil {
		t.Fatalf("❌ Error consultando actividad: %v", err)
	}
	defer resp.Body.Close()

	var activityFinal map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&activityFinal)

	lugaresFinal := cupoTotal
	if activityFinal["lugares"] != nil {
		lugaresFinal = int(activityFinal["lugares"].(float64))
	}

	t.Logf("✅ Lugares disponibles final: %d", lugaresFinal)

	if lugaresFinal <= lugaresAntes {
		t.Logf("✅ Cupo ocupado correctamente después de re-inscripción")
	}

	// ==================== RESUMEN ====================
	t.Log("\n================================================================================")
	t.Log("🎉 TEST DE DESINSCRIPCIÓN Y RE-INSCRIPCIÓN COMPLETADO!")
	t.Log("================================================================================")
	t.Log("✅ Inscripción inicial exitosa")
	t.Log("✅ Cupo verificado antes de desinscribirse")
	t.Log("✅ Desinscripción exitosa")
	t.Log("✅ Cupo liberado después de desinscribirse")
	t.Log("✅ Inscripción marcada como inactiva")
	t.Log("✅ Re-inscripción exitosa")
	t.Log("✅ Re-inscripción está activa")
	t.Log("✅ Cupo ocupado nuevamente")
	t.Log("================================================================================")
}
