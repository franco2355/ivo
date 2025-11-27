package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestPaymentSecurityEndpoints prueba todos los endpoints de pagos con validaciones de seguridad
func TestPaymentSecurityEndpoints(t *testing.T) {
	t.Log("🔒 INICIANDO TEST DE SEGURIDAD DE ENDPOINTS DE PAGOS")

	// ==================== SETUP ====================
	t.Log("\n📝 PASO 1: Crear usuarios de prueba")
	user1Token, user1ID, user1Data := registerUser(t)
	t.Logf("✅ Usuario 1 registrado - ID: %d, Username: %s", user1ID, user1Data.Username)

	user2Token, user2ID, user2Data := registerUser(t)
	t.Logf("✅ Usuario 2 registrado - ID: %d, Username: %s", user2ID, user2Data.Username)

	adminToken, adminID := login(t, "admin", "admin123")
	t.Logf("✅ Admin logueado - ID: %d", adminID)

	// Crear suscripciones para generar pagos
	t.Log("\n📝 PASO 2: Crear suscripciones y pagos para pruebas")
	sub1ID := createSubscription(t, user1Token, user1ID, PlanBasicoID)
	t.Logf("✅ Suscripción User1 creada - ID: %s", sub1ID)

	sub2ID := createSubscription(t, user2Token, user2ID, PlanBasicoID)
	t.Logf("✅ Suscripción User2 creada - ID: %s", sub2ID)

	// Admin crea pagos en efectivo para ambos usuarios
	payment1ID := createCashPayment(t, adminToken, user1ID, sub1ID, 1000.0)
	t.Logf("✅ Pago User1 creado - ID: %s", payment1ID)

	payment2ID := createCashPayment(t, adminToken, user2ID, sub2ID, 1000.0)
	t.Logf("✅ Pago User2 creado - ID: %s", payment2ID)

	// ==================== TEST 1: POST /payments ====================
	t.Run("POST /payments - Create payment", func(t *testing.T) {
		t.Log("\n🧪 TEST 1: POST /payments")

		t.Run("Sin autenticación debe fallar", func(t *testing.T) {
			paymentData := map[string]interface{}{
				"user_id":       fmt.Sprintf("%d", user1ID),
				"entity_type":   "subscription",
				"entity_id":     sub1ID,
				"amount":        1000.0,
				"payment_method": "cash",
			}
			body, _ := json.Marshal(paymentData)

			req, _ := http.NewRequest("POST", "http://localhost:8083/payments", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			// No se envía token

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("❌ Sin auth debería ser 401, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Sin auth rechazado correctamente (401)")
			}
		})

		t.Run("Usuario autenticado puede crear pago", func(t *testing.T) {
			// Ya probado en helpers - solo verificar que funciona
			testSubID := createSubscription(t, user1Token, user1ID, PlanBasicoID)
			testPaymentID := createCashPayment(t, user1Token, user1ID, testSubID, 1000.0)
			if testPaymentID == "" {
				t.Error("❌ Usuario autenticado no pudo crear pago")
			} else {
				t.Logf("✅ Usuario autenticado creó pago: %s", testPaymentID)
			}
		})
	})

	// ==================== TEST 2: GET /payments ====================
	t.Run("GET /payments - List all payments", func(t *testing.T) {
		t.Log("\n🧪 TEST 2: GET /payments")

		t.Run("Sin autenticación debe fallar", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8083/payments", nil)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("❌ Sin auth debería ser 401, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Sin auth rechazado correctamente (401)")
			}
		})

		t.Run("Usuario regular no puede ver todos los pagos", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8083/payments", nil)
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("❌ Usuario regular debería recibir 403, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Usuario regular bloqueado correctamente (403)")
			}
		})

		t.Run("Admin puede ver todos los pagos", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8083/payments", nil)
			req.Header.Set("Authorization", adminToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("❌ Admin debería recibir 200, obtenido: %d", resp.StatusCode)
			} else {
				var payments []map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&payments)
				t.Logf("✅ Admin puede ver todos los pagos (%d pagos)", len(payments))
			}
		})
	})

	// ==================== TEST 3: GET /payments/:id ====================
	t.Run("GET /payments/:id - Get specific payment", func(t *testing.T) {
		t.Log("\n🧪 TEST 3: GET /payments/:id")

		t.Run("Sin autenticación debe fallar", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/%s", payment1ID), nil)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("❌ Sin auth debería ser 401, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Sin auth rechazado correctamente (401)")
			}
		})

		t.Run("Usuario puede ver su propio pago", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/%s", payment1ID), nil)
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				var errorResp map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&errorResp)
				t.Errorf("❌ Usuario debería ver su pago (200), obtenido: %d - %v", resp.StatusCode, errorResp)
			} else {
				t.Log("✅ Usuario puede ver su propio pago")
			}
		})

		t.Run("Usuario NO puede ver pago de otro usuario", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/%s", payment2ID), nil)
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("❌ Debería ser 403 al intentar ver pago ajeno, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Acceso a pago ajeno bloqueado correctamente (403)")
			}
		})

		t.Run("Admin puede ver cualquier pago", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/%s", payment1ID), nil)
			req.Header.Set("Authorization", adminToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("❌ Admin debería ver cualquier pago (200), obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Admin puede ver cualquier pago")
			}
		})
	})

	// ==================== TEST 4: GET /payments/user/:id ====================
	t.Run("GET /payments/user/:id - Get user payments", func(t *testing.T) {
		t.Log("\n🧪 TEST 4: GET /payments/user/:id")

		t.Run("Sin autenticación debe fallar", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/user/%d", user1ID), nil)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("❌ Sin auth debería ser 401, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Sin auth rechazado correctamente (401)")
			}
		})

		t.Run("Usuario puede ver sus propios pagos", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/user/%d", user1ID), nil)
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				var errorResp map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&errorResp)
				t.Errorf("❌ Usuario debería ver sus pagos (200), obtenido: %d - %v", resp.StatusCode, errorResp)
			} else {
				var payments []map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&payments)
				t.Logf("✅ Usuario puede ver sus propios pagos (%d pagos)", len(payments))
			}
		})

		t.Run("Usuario NO puede ver pagos de otro usuario", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/user/%d", user2ID), nil)
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("❌ Debería ser 403 al intentar ver pagos ajenos, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Acceso a pagos ajenos bloqueado correctamente (403)")
			}
		})

		t.Run("Admin puede ver pagos de cualquier usuario", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8083/payments/user/%d", user1ID), nil)
			req.Header.Set("Authorization", adminToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("❌ Admin debería ver pagos de cualquier usuario (200), obtenido: %d", resp.StatusCode)
			} else {
				var payments []map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&payments)
				t.Logf("✅ Admin puede ver pagos de cualquier usuario (%d pagos)", len(payments))
			}
		})
	})

	// ==================== TEST 5: GET /payments/status/:status ====================
	t.Run("GET /payments/status/:status - Filter by status", func(t *testing.T) {
		t.Log("\n🧪 TEST 5: GET /payments/status/:status")

		t.Run("Sin autenticación debe fallar", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8083/payments/status/pending", nil)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("❌ Sin auth debería ser 401, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Sin auth rechazado correctamente (401)")
			}
		})

		t.Run("Usuario regular no puede filtrar por status", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8083/payments/status/pending", nil)
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("❌ Usuario regular debería recibir 403, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Usuario regular bloqueado correctamente (403)")
			}
		})

		t.Run("Admin puede filtrar por status", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8083/payments/status/pending", nil)
			req.Header.Set("Authorization", adminToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				var errorResp map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&errorResp)
				t.Errorf("❌ Admin debería filtrar por status (200), obtenido: %d - %v", resp.StatusCode, errorResp)
			} else {
				var payments []map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&payments)
				t.Logf("✅ Admin puede filtrar por status (%d pagos pendientes)", len(payments))
			}
		})
	})

	// ==================== TEST 6: PATCH /payments/:id/status ====================
	t.Run("PATCH /payments/:id/status - Update payment status", func(t *testing.T) {
		t.Log("\n🧪 TEST 6: PATCH /payments/:id/status")

		t.Run("Sin autenticación debe fallar", func(t *testing.T) {
			updateData := map[string]string{"status": "completed"}
			body, _ := json.Marshal(updateData)

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:8083/payments/%s/status", payment1ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("❌ Sin auth debería ser 401, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Sin auth rechazado correctamente (401)")
			}
		})

		t.Run("Usuario regular no puede actualizar status", func(t *testing.T) {
			updateData := map[string]string{"status": "completed"}
			body, _ := json.Marshal(updateData)

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:8083/payments/%s/status", payment1ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", user1Token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("❌ Usuario regular debería recibir 403, obtenido: %d", resp.StatusCode)
			} else {
				t.Log("✅ Usuario regular bloqueado correctamente (403)")
			}
		})

		t.Run("Admin puede actualizar status", func(t *testing.T) {
			updateData := map[string]string{"status": "completed"}
			body, _ := json.Marshal(updateData)

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:8083/payments/%s/status", payment1ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", adminToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error en request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				var errorResp map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&errorResp)
				t.Errorf("❌ Admin debería actualizar status (200), obtenido: %d - %v", resp.StatusCode, errorResp)
			} else {
				t.Log("✅ Admin puede actualizar status")
			}
		})
	})

	// ==================== RESUMEN ====================
	t.Log("\n================================================================================")
	t.Log("🎉 TEST DE SEGURIDAD DE ENDPOINTS DE PAGOS COMPLETADO!")
	t.Log("================================================================================")
	t.Log("✅ POST /payments - Autenticación requerida")
	t.Log("✅ GET /payments - Solo admin")
	t.Log("✅ GET /payments/:id - Solo propietario o admin")
	t.Log("✅ GET /payments/user/:id - Solo propietario o admin")
	t.Log("✅ GET /payments/status/:status - Solo admin")
	t.Log("✅ PATCH /payments/:id/status - Solo admin")
	t.Log("================================================================================")
}
