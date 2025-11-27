package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthCheckEndpoint verifica el endpoint de health check
func TestHealthCheckEndpoint(t *testing.T) {
	// Simular handler simple de health check
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteStatus(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": "test-service",
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got: %s", response["status"])
	}
}

// TestJSONResponseEncoding verifica encoding de respuestas JSON
func TestJSONResponseEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "success",
			"data": map[string]interface{}{
				"id":   123,
				"name": "Test",
			},
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got: %s", contentType)
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)

	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	if response["message"] != "success" {
		t.Errorf("Expected message 'success', got: %s", response["message"])
	}
}

// TestRequestBodyParsing verifica parsing de request body
func TestRequestBodyParsing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid JSON",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"received": body,
		})
	})

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
	}{
		{
			name:           "valid JSON",
			payload:        `{"name": "test", "value": 123}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			payload:        `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			payload:        ``,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got: %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestAuthorizationHeaderParsing verifica parsing del header de autorización
func TestAuthorizationHeaderParsing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "missing authorization header",
			})
			return
		}

		// Verificar formato Bearer
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid authorization format",
			})
			return
		}

		token := authHeader[7:]

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token_received": token,
		})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid Bearer token",
			authHeader:     "Bearer valid_token_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing Bearer prefix",
			authHeader:     "valid_token_123",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "only Bearer",
			authHeader:     "Bearer",
			expectedStatus: http.StatusOK, // Token vacío pero formato correcto
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got: %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestCORSHeaders verifica headers CORS
func TestCORSHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular middleware CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	})

	t.Run("OPTIONS request for CORS preflight", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got: %d", w.Code)
		}

		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got: %s", allowOrigin)
		}

		allowMethods := w.Header().Get("Access-Control-Allow-Methods")
		if allowMethods == "" {
			t.Error("Expected Access-Control-Allow-Methods header")
		}
	})

	t.Run("Regular request with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got: %s", allowOrigin)
		}
	})
}

// TestErrorResponseFormat verifica formato de respuestas de error
func TestErrorResponseFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorType := r.URL.Query().Get("error")

		switch errorType {
		case "not_found":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "not_found",
				"message": "Resource not found",
			})
		case "bad_request":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "bad_request",
				"message": "Invalid request",
				"details": []string{"field 'name' is required"},
			})
		case "server_error":
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "internal_server_error",
				"message": "An unexpected error occurred",
			})
		default:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}
	})

	tests := []struct {
		name           string
		errorParam     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "not found error",
			errorParam:     "not_found",
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name:           "bad request error",
			errorParam:     "bad_request",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name:           "server error",
			errorParam:     "server_error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal_server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?error="+tt.errorParam, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got: %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			json.NewDecoder(w.Body).Decode(&response)

			if response["error"] != tt.expectedError {
				t.Errorf("Expected error '%s', got: %s", tt.expectedError, response["error"])
			}

			if response["message"] == nil {
				t.Error("Expected error message in response")
			}
		})
	}
}

// TestQueryParameterParsing verifica parsing de query parameters
func TestQueryParameterParsing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		page := query.Get("page")
		limit := query.Get("limit")
		filter := query.Get("filter")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
		})
	})

	req := httptest.NewRequest("GET", "/test?page=1&limit=10&filter=active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["page"] != "1" {
		t.Errorf("Expected page '1', got: %s", response["page"])
	}

	if response["limit"] != "10" {
		t.Errorf("Expected limit '10', got: %s", response["limit"])
	}

	if response["filter"] != "active" {
		t.Errorf("Expected filter 'active', got: %s", response["filter"])
	}
}
