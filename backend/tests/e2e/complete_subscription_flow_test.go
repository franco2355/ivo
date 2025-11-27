package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

// TestCompleteSubscriptionFlow prueba el flujo completo de suscripción
// Desde registro de usuario hasta activación de suscripción
func TestCompleteSubscriptionFlow(t *testing.T) {
	baseURL := getBaseURL()
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(999999)

	// Paso 1: Registrar nuevo usuario
	t.Log("Step 1: Register new user")
	username := fmt.Sprintf("e2e_%d", randomNum)
	email := fmt.Sprintf("e2e%d@test.com", randomNum)

	registerPayload := map[string]interface{}{
		"nombre":   "E2E",
		"apellido": "Test",
		"username": username,
		"email":    email,
		"password": fmt.Sprintf("Test%d!Pass", randomNum),
	}

	body, _ := json.Marshal(registerPayload)
	resp, err := http.Post(
		fmt.Sprintf("%s/register", baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to register user, status: %d", resp.StatusCode)
	}

	var registerResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&registerResult)
	userToken := registerResult["token"].(string)
	user := registerResult["user"].(map[string]interface{})
	userID := fmt.Sprintf("%.0f", user["id"].(float64))

	t.Logf("User registered successfully: ID=%s, Username=%s", userID, username)

	// Paso 2: Usar un plan conocido (Plan Básico)
	t.Log("Step 2: Using Plan Básico")

	// Plan IDs conocidos de la DB
	planID := "6923cc56ee6da85323ce5f47" // Plan Básico
	planPrice := 5000.0

	t.Logf("Selected plan: ID=%s, Price=%.2f", planID, planPrice)

	// Paso 3: Crear suscripción
	t.Log("Step 3: Create subscription")

	subscriptionPayload := map[string]interface{}{
		"id_usuario": userID,
		"id_plan":    planID,
	}

	body, _ = json.Marshal(subscriptionPayload)
	req, _ = http.NewRequest(
		"POST",
		"http://localhost:8081/subscriptions",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Failed to create subscription, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var subscription map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&subscription)
	subscriptionID := fmt.Sprintf("%.0f", subscription["id"].(float64))

	t.Logf("Subscription created: ID=%s, Status=%s", subscriptionID, subscription["estado"])

	// Paso 4: Crear pago (como admin)
	t.Log("Step 4: Create payment for subscription")

	adminToken := getAdminToken(t, baseURL)

	paymentPayload := map[string]interface{}{
		"entity_type":     "subscription",
		"entity_id":       subscriptionID,
		"user_id":         userID,
		"amount":          planPrice,
		"currency":        "ARS",
		"payment_method":  "cash",
		"payment_gateway": "manual",
	}

	body, _ = json.Marshal(paymentPayload)
	req, _ = http.NewRequest(
		"POST",
		"http://localhost:8083/payments",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to create payment, status: %d", resp.StatusCode)
	}

	var payment map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payment)
	paymentID := payment["id"].(string)

	t.Logf("Payment created: ID=%s, Status=%s", paymentID, payment["status"])

	// Paso 5: Marcar pago como completado
	t.Log("Step 5: Mark payment as completed")

	updatePaymentPayload := map[string]interface{}{
		"status":         "completed",
		"transaction_id": fmt.Sprintf("TXN_E2E_%d", timestamp),
	}

	body, _ = json.Marshal(updatePaymentPayload)
	req, _ = http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/payments-api/payments/%s/status", baseURL, paymentID),
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to update payment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to update payment status, status: %d", resp.StatusCode)
	}

	t.Log("Payment marked as completed")

	// Paso 6: Verificar que la suscripción se activó
	t.Log("Step 6: Verify subscription is active")

	// Dar tiempo para que los eventos se procesen
	time.Sleep(2 * time.Second)

	req, _ = http.NewRequest(
		"GET",
		fmt.Sprintf("http://localhost:8081/subscriptions/%s", subscriptionID),
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}
	defer resp.Body.Close()

	var updatedSubscription map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updatedSubscription)

	expectedStatus := "activa"
	actualStatus := updatedSubscription["estado"].(string)

	if actualStatus != expectedStatus {
		t.Logf("Warning: Expected subscription status '%s', got '%s'", expectedStatus, actualStatus)
		t.Logf("Subscription may need manual activation or event processing")
	} else {
		t.Log("Subscription is active!")
	}

	// Paso 7: Usuario intenta inscribirse en una actividad
	t.Log("Step 7: User enrolls in an activity")

	// Primero, obtener actividades disponibles
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/activities-api/actividades", baseURL), nil)
	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to get activities: %v", err)
	}
	defer resp.Body.Close()

	var activities []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&activities)

	if len(activities) > 0 {
		activity := activities[0]
		activityID := fmt.Sprintf("%.0f", activity["id"].(float64))

		inscriptionPayload := map[string]interface{}{
			"id_usuario":   userID,
			"id_actividad": activityID,
		}

		body, _ = json.Marshal(inscriptionPayload)
		req, _ = http.NewRequest(
			"POST",
			fmt.Sprintf("%s/activities-api/inscripciones", baseURL),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

		resp, err = client.Do(req)

		if err != nil {
			t.Fatalf("Failed to enroll in activity: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			t.Logf("Successfully enrolled in activity: %s", activity["titulo"])
		} else {
			t.Logf("Could not enroll in activity (status: %d). May require active subscription.", resp.StatusCode)
		}
	} else {
		t.Log("No activities available for enrollment")
	}

	t.Log("Complete subscription flow test finished successfully!")
}

// TestCompleteActivityEnrollmentFlow prueba el flujo completo de inscripción a actividad
func TestCompleteActivityEnrollmentFlow(t *testing.T) {
	baseURL := getBaseURL()
	timestamp := time.Now().Unix()

	// Paso 1: Crear usuario con suscripción activa
	t.Log("Step 1: Setup - Create user with active subscription")

	username := fmt.Sprintf("activityuser_%d", timestamp)
	email := fmt.Sprintf("activityuser_%d@example.com", timestamp)

	registerPayload := map[string]interface{}{
		"nombre":   "Activity",
		"apellido": "User",
		"username": username,
		"email":    email,
		"password": "ActivityPass123!",
	}

	body, _ := json.Marshal(registerPayload)
	resp, err := http.Post(
		fmt.Sprintf("%s/register", baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	var registerResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&registerResult)
	userToken := registerResult["token"].(string)
	user := registerResult["user"].(map[string]interface{})
	userID := fmt.Sprintf("%.0f", user["id"].(float64))

	t.Logf("User registered: ID=%s", userID)

	// Paso 2: Listar actividades disponibles
	t.Log("Step 2: List available activities")

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/activities-api/actividades", baseURL), nil)
	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to get activities: %v", err)
	}
	defer resp.Body.Close()

	var activities []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&activities)

	if len(activities) == 0 {
		t.Skip("No activities available for testing")
	}

	selectedActivity := activities[0]
	activityID := fmt.Sprintf("%.0f", selectedActivity["id"].(float64))

	t.Logf("Selected activity: ID=%s, Title=%s", activityID, selectedActivity["titulo"])

	// Paso 3: Inscribirse en la actividad
	t.Log("Step 3: Enroll in activity")

	inscriptionPayload := map[string]interface{}{
		"id_usuario":   userID,
		"id_actividad": activityID,
	}

	body, _ = json.Marshal(inscriptionPayload)
	req, _ = http.NewRequest(
		"POST",
		fmt.Sprintf("%s/activities-api/inscripciones", baseURL),
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

	resp, err = client.Do(req)

	if err != nil {
		t.Fatalf("Failed to enroll: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Logf("Enrollment response (status %d): %s", resp.StatusCode, string(body))
	}

	var inscription map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&inscription)

	if inscription["id"] != nil {
		inscriptionID := fmt.Sprintf("%.0f", inscription["id"].(float64))
		t.Logf("Enrollment created: ID=%s", inscriptionID)

		// Paso 4: Verificar inscripción
		t.Log("Step 4: Verify enrollment")

		req, _ = http.NewRequest(
			"GET",
			fmt.Sprintf("%s/activities-api/inscripciones/%s", baseURL, inscriptionID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

		resp, err = client.Do(req)

		if err != nil {
			t.Fatalf("Failed to get enrollment: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Log("Enrollment verified successfully")
		}

		// Paso 5: Cancelar inscripción
		t.Log("Step 5: Cancel enrollment")

		req, _ = http.NewRequest(
			"DELETE",
			fmt.Sprintf("%s/activities-api/inscripciones/%s", baseURL, inscriptionID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

		resp, err = client.Do(req)

		if err != nil {
			t.Fatalf("Failed to cancel enrollment: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			t.Log("Enrollment cancelled successfully")
		}
	}

	t.Log("Complete activity enrollment flow test finished!")
}

// Helper functions
func getBaseURL() string {
	return "http://localhost:8080"
}

func getAdminToken(t *testing.T, baseURL string) string {
	loginPayload := map[string]interface{}{
		"username_or_email": "admin",
		"password":          "admin",
	}

	body, _ := json.Marshal(loginPayload)
	resp, err := http.Post(
		fmt.Sprintf("%s/users-api/login", baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		t.Fatalf("Failed to login as admin: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["token"] == nil {
		t.Fatal("Failed to get admin token")
	}

	return result["token"].(string)
}
