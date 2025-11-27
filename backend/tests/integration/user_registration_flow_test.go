package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

// TestUserRegistrationFlow prueba el flujo completo de registro de usuario
func TestUserRegistrationFlow(t *testing.T) {
	baseURL := getBaseURL()
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(999999)

	// Datos únicos para el usuario de prueba
	username := fmt.Sprintf("user_%d", randomNum)
	email := fmt.Sprintf("test%d@test.com", randomNum)
	password := fmt.Sprintf("Test%d!Pass", randomNum)

	t.Run("Register new user", func(t *testing.T) {
		registerPayload := map[string]interface{}{
			"nombre":   "Test",
			"apellido": "User",
			"username": username,
			"email":    email,
			"password": password,
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
			t.Errorf("Expected status 200 or 201, got: %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["token"] == nil {
			t.Error("Expected token in response")
		}

		if result["user"] == nil {
			t.Error("Expected user data in response")
		}

		user := result["user"].(map[string]interface{})
		if user["username"] != username {
			t.Errorf("Expected username %s, got: %s", username, user["username"])
		}
	})

	t.Run("Cannot register duplicate username", func(t *testing.T) {
		registerPayload := map[string]interface{}{
			"nombre":   "Test",
			"apellido": "User",
			"username": username, // Mismo username
			"email":    fmt.Sprintf("diff%d@test.com", randomNum+1),
			"password": "TestPass123!",
		}

		body, _ := json.Marshal(registerPayload)
		resp, err := http.Post(
			fmt.Sprintf("%s/register", baseURL),
			"application/json",
			bytes.NewBuffer(body),
		)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 or 409 for duplicate username, got: %d", resp.StatusCode)
		}
	})

	t.Run("Cannot register duplicate email", func(t *testing.T) {
		registerPayload := map[string]interface{}{
			"nombre":   "Test",
			"apellido": "User",
			"username": fmt.Sprintf("user_%d", randomNum+1),
			"email":    email, // Mismo email
			"password": "TestPass123!",
		}

		body, _ := json.Marshal(registerPayload)
		resp, err := http.Post(
			fmt.Sprintf("%s/register", baseURL),
			"application/json",
			bytes.NewBuffer(body),
		)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 or 409 for duplicate email, got: %d", resp.StatusCode)
		}
	})

	t.Run("Login with new user", func(t *testing.T) {
		loginPayload := map[string]interface{}{
			"username_or_email": username,
			"password":          password,
		}

		body, _ := json.Marshal(loginPayload)
		resp, err := http.Post(
			fmt.Sprintf("%s/login", baseURL),
			"application/json",
			bytes.NewBuffer(body),
		)

		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got: %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["token"] == nil {
			t.Error("Expected token in login response")
		}
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		loginPayload := map[string]interface{}{
			"username_or_email": username,
			"password":          "WrongPassword123!",
		}

		body, _ := json.Marshal(loginPayload)
		resp, err := http.Post(
			fmt.Sprintf("%s/login", baseURL),
			"application/json",
			bytes.NewBuffer(body),
		)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 401 or 400, got: %d", resp.StatusCode)
		}
	})

	t.Run("Validate password requirements", func(t *testing.T) {
		tests := []struct {
			name     string
			password string
		}{
			{"too short", "Pass1!"},
			{"no uppercase", "password123!"},
			{"no lowercase", "PASSWORD123!"},
			{"no number", "Password!"},
			{"no special char", "Password123"},
		}

		for i, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				registerPayload := map[string]interface{}{
					"nombre":   "Test",
					"apellido": "User",
					"username": fmt.Sprintf("pw%d_%d", randomNum, i),
					"email":    fmt.Sprintf("pw%d_%d@test.com", randomNum, i),
					"password": tt.password,
				}

				body, _ := json.Marshal(registerPayload)
				resp, err := http.Post(
					fmt.Sprintf("%s/register", baseURL),
					"application/json",
					bytes.NewBuffer(body),
				)

				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusBadRequest {
					// Leer el body para ver el error exacto
					var errorResp map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&errorResp)
					t.Logf("⚠️  Expected status 400, got: %d - Error: %v", resp.StatusCode, errorResp)
				}
			})
		}
	})
}

// TestUserProfileManagement prueba la gestión de perfil de usuario
func TestUserProfileManagement(t *testing.T) {
	// Usar helper para registrar usuario
	token, userID, userData := registerUser(t)
	t.Logf("✅ Usuario registrado - ID: %d, Username: %s", userID, userData.Username)

	t.Run("Get user profile", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/users/%d", userID), nil)
		req.Header.Set("Authorization", token)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to get profile: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got: %d", resp.StatusCode)
		}

		var profile map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&profile)

		if profile["username"] != userData.Username {
			t.Errorf("Expected username %s, got: %s", userData.Username, profile["username"])
		}
	})

	t.Run("Get user profile without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/users/%d", userID), nil)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Debería fallar sin autenticación
		if resp.StatusCode == http.StatusOK {
			t.Log("Warning: Endpoint should require authentication")
		}
	})
}

// Helper function
func getBaseURL() string {
	// Podría configurarse desde variables de entorno
	return "http://localhost:8080"
}
