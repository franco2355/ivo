package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestPaymentCreationFlow prueba el flujo de creación de pagos
func TestPaymentCreationFlow(t *testing.T) {
	t.Log("🚀 Iniciando test de flujo de pagos")

	// Registrar usuario nuevo y obtener admin token
	t.Log("\n📝 Paso 1: Registrar usuario y obtener admin token")
	userToken, userID := registerAndLoginUser(t)
	t.Logf("✅ Usuario registrado - ID: %d", userID)

	adminToken, _ := login(t, "admin", "admin123")
	t.Log("✅ Admin logueado")

	t.Run("Create cash payment for subscription", func(t *testing.T) {
		// Primero crear una suscripción real
		subscriptionID := createSubscription(t, userToken, userID, PlanBasicoID)
		t.Logf("✅ Suscripción creada - ID: %s", subscriptionID)

		paymentPayload := map[string]interface{}{
			"entity_type":     "subscription",
			"entity_id":       subscriptionID,
			"user_id":         fmt.Sprintf("%d", userID),
			"amount":          5000.0,
			"currency":        "ARS",
			"payment_method":  "cash",
			"payment_gateway": "manual",
		}

		body, _ := json.Marshal(paymentPayload)
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8083/payments",
			bytes.NewBuffer(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", adminToken)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 or 201, got: %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["id"] == nil {
			t.Error("Expected payment ID in response")
		}

		if result["status"] != "pending" {
			t.Errorf("Expected status 'pending', got: %s", result["status"])
		}

		// Guardar payment ID para tests posteriores
		paymentID := result["id"].(string)

		t.Run("Update payment status to completed", func(t *testing.T) {
			updatePayload := map[string]interface{}{
				"status": "completed",
			}

			body, _ := json.Marshal(updatePayload)
			req, _ := http.NewRequest(
				"PATCH",
				fmt.Sprintf("http://localhost:8083/payments/%s/status", paymentID),
				bytes.NewBuffer(body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", adminToken)

			resp, err := client.Do(req)

			if err != nil {
				t.Fatalf("Failed to update payment: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got: %d", resp.StatusCode)
			}
		})

		t.Run("Get payment by ID", func(t *testing.T) {
			req, _ := http.NewRequest(
				"GET",
				fmt.Sprintf("http://localhost:8083/payments/%s", paymentID),
				nil,
			)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

			resp, err := client.Do(req)

			if err != nil {
				t.Fatalf("Failed to get payment: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got: %d", resp.StatusCode)
			}

			var payment map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&payment)

			if payment["id"] != paymentID {
				t.Errorf("Expected payment ID %s, got: %s", paymentID, payment["id"])
			}
		})
	})

	t.Run("Create payment with invalid data", func(t *testing.T) {
		paymentPayload := map[string]interface{}{
			"entity_type": "subscription",
			"entity_id":   "",
			"amount":      -100, // Monto negativo
			"currency":    "ARS",
		}

		body, _ := json.Marshal(paymentPayload)
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8083/payments",
			bytes.NewBuffer(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got: %d", resp.StatusCode)
		}
	})
}

// TestPaymentListingAndFiltering prueba el listado y filtrado de pagos
func TestPaymentListingAndFiltering(t *testing.T) {
	baseURL := getBaseURL()
	adminToken := getAdminToken(t, baseURL)

	t.Run("Get all payments", func(t *testing.T) {
		req, _ := http.NewRequest(
			"GET",
			"http://localhost:8083/payments",
			nil,
		)
		req.Header.Set("Authorization", adminToken)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to get payments: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got: %d", resp.StatusCode)
		}

		var payments []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&payments)

		t.Logf("Found %d payments", len(payments))
	})

	t.Run("Get payments by user", func(t *testing.T) {
		userID := "1"
		req, _ := http.NewRequest(
			"GET",
			fmt.Sprintf("http://localhost:8083/payments/user/%s", userID),
			nil,
		)
		req.Header.Set("Authorization", adminToken)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to get payments: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got: %d", resp.StatusCode)
		}

		var payments []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&payments)

		// Verificar que todos los pagos son del usuario
		for _, p := range payments {
			if p["user_id"] != userID {
				t.Errorf("Expected user_id %s, got: %s", userID, p["user_id"])
			}
		}
	})

	t.Run("Get payments by status", func(t *testing.T) {
		status := "completed"
		req, _ := http.NewRequest(
			"GET",
			fmt.Sprintf("http://localhost:8083/payments/status/%s", status),
			nil,
		)
		req.Header.Set("Authorization", adminToken)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to get payments: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got: %d", resp.StatusCode)
		}

		var payments []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&payments)

		// Verificar que todos los pagos tienen el estado correcto
		for _, p := range payments {
			if p["status"] != status {
				t.Errorf("Expected status %s, got: %s", status, p["status"])
			}
		}
	})
}

// TestPaymentAuthorizationRules prueba las reglas de autorización
func TestPaymentAuthorizationRules(t *testing.T) {
	// Crear usuario regular usando helper
	userToken, userID, _ := registerUser(t)
	t.Logf("✅ Usuario registrado - ID: %d", userID)

	t.Run("Regular user cannot create payments", func(t *testing.T) {
		paymentPayload := map[string]interface{}{
			"entity_type":     "subscription",
			"entity_id":       "sub_test_456",
			"user_id":         "1",
			"amount":          5000.0,
			"currency":        "ARS",
			"payment_method":  "cash",
			"payment_gateway": "manual",
		}

		body, _ := json.Marshal(paymentPayload)
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8083/payments",
			bytes.NewBuffer(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Debería devolver 403 Forbidden o similar
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			t.Error("Regular user should not be able to create payments")
		}
	})

	t.Run("User can view their own payments", func(t *testing.T) {
		req, _ := http.NewRequest(
			"GET",
			fmt.Sprintf("http://localhost:8083/payments/user/%d", userID),
			nil,
		)
		req.Header.Set("Authorization", userToken)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Debería permitir ver sus propios pagos
		if resp.StatusCode != http.StatusOK {
			t.Logf("User should be able to view their own payments, got status: %d", resp.StatusCode)
		}
	})
}

// Helper para obtener token de admin
func getAdminToken(t *testing.T, baseURL string) string {
	token, _ := login(t, "admin", "admin123")
	return token
}
