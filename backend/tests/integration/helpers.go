package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

// API URLs
const (
	usersAPIURL         = "http://localhost:8080"
	subscriptionsAPIURL = "http://localhost:8081"
	activitiesAPIURL    = "http://localhost:8082"
	paymentsAPIURL      = "http://localhost:8083"
)

// Plan IDs - Configurados en MongoDB
const (
	PlanBasicoID    = "6923cc56ee6da85323ce5f47"
	PlanPremiumID   = "6923cc56ee6da85323ce5f48"
	PlanEstudianteID = "6923cc56ee6da85323ce5f49"
)

// LoginResponse estructura de respuesta del login
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
	} `json:"user"`
}

// SubscriptionResponse estructura de respuesta de suscripción
type SubscriptionResponse struct {
	ID               string `json:"id"`
	UsuarioID        string `json:"usuario_id"`
	PlanID           string `json:"plan_id"`
	PlanNombre       string `json:"plan_nombre"`
	Estado           string `json:"estado"`
	PagoID           string `json:"pago_id"`
	FechaInicio      string `json:"fecha_inicio"`
	FechaVencimiento string `json:"fecha_vencimiento"`
}

// PaymentResponse estructura de respuesta de pago
type PaymentResponse struct {
	ID             string  `json:"id"`
	EntityType     string  `json:"entity_type"`
	EntityID       string  `json:"entity_id"`
	UserID         string  `json:"user_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Status         string  `json:"status"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentGateway string  `json:"payment_gateway"`
}

// InscripcionResponse estructura de respuesta de inscripción
type InscripcionResponse struct {
	ID          int    `json:"id"`
	UsuarioID   int    `json:"usuario_id"`
	ActividadID int    `json:"actividad_id"`
	IsActiva    bool   `json:"is_activa"`
	FechaInicio string `json:"fecha_inicio"`
}

// ErrorResponse estructura de respuesta de error
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

// RegisterRequest estructura para registro de usuario
type RegisterRequest struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// generateRandomUser genera un usuario aleatorio con datos válidos
func generateRandomUser() RegisterRequest {
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(999999)

	// Username entre 3 y 30 caracteres
	username := fmt.Sprintf("user_%d", randomNum)
	email := fmt.Sprintf("test%d@test.com", randomNum)

	// Generar password que cumple todas las validaciones:
	// - Al menos 8 caracteres
	// - Al menos una mayúscula
	// - Al menos una minúscula
	// - Al menos un número
	// - Al menos un carácter especial
	password := fmt.Sprintf("Test%d!Pass", randomNum)

	return RegisterRequest{
		Nombre:   "Test",
		Apellido: "User",
		Username: username,
		Email:    email,
		Password: password,
	}
}

// registerUser registra un nuevo usuario y retorna token y userID
func registerUser(t *testing.T) (string, int, RegisterRequest) {
	userData := generateRandomUser()

	body, _ := json.Marshal(userData)
	resp, err := http.Post(usersAPIURL+"/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("❌ Error registrando usuario: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Registro falló - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("❌ Error decodificando respuesta de registro: %v", err)
	}

	t.Logf("✅ Usuario registrado - Username: %s, ID: %d", userData.Username, loginResp.User.ID)

	return "Bearer " + loginResp.Token, loginResp.User.ID, userData
}

// registerAndLoginUser crea un usuario nuevo y retorna sus credenciales
func registerAndLoginUser(t *testing.T) (string, int) {
	token, userID, _ := registerUser(t)
	return token, userID
}

// Helper function: login - compatible con ambos tipos (int y uint)
// No verifica el rol, acepta cualquier usuario
func login(t *testing.T, username, password string) (string, uint) {
	reqBody := map[string]string{
		"username_or_email": username,
		"password":          password,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(usersAPIURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("❌ Error en login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Login falló - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("❌ Error decodificando respuesta de login: %v", err)
	}

	return "Bearer " + loginResp.Token, uint(loginResp.User.ID)
}

// Helper function: loginUser - función original
func loginUser(t *testing.T, username, password string, isAdmin bool) (string, int) {
	reqBody := map[string]string{
		"username_or_email": username,
		"password":          password,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(usersAPIURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("❌ Error en login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Login falló - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("❌ Error decodificando respuesta de login: %v", err)
	}

	if loginResp.User.IsAdmin != isAdmin {
		t.Fatalf("❌ Usuario no tiene el rol esperado. IsAdmin: %v, Esperado: %v", loginResp.User.IsAdmin, isAdmin)
	}

	return "Bearer " + loginResp.Token, loginResp.User.ID
}

// Helper function: createSubscription
func createSubscription(t *testing.T, token string, userID interface{}, planID string) string {
	var userIDStr string
	switch v := userID.(type) {
	case int:
		userIDStr = fmt.Sprintf("%d", v)
	case uint:
		userIDStr = fmt.Sprintf("%d", v)
	case string:
		userIDStr = v
	default:
		userIDStr = fmt.Sprintf("%v", v)
	}

	reqBody := map[string]interface{}{
		"usuario_id":         userIDStr,
		"plan_id":            planID,
		"sucursal_origen_id": "1",
		"metodo_pago":        "cash",
		"auto_renovacion":    false,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", subscriptionsAPIURL+"/subscriptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error creando suscripción: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error creando suscripción - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var sub SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		t.Fatalf("❌ Error decodificando respuesta de suscripción: %v", err)
	}

	return sub.ID
}

// Helper function: getSubscription
func getSubscription(t *testing.T, token, subscriptionID string) SubscriptionResponse {
	req, _ := http.NewRequest("GET", subscriptionsAPIURL+"/subscriptions/"+subscriptionID, nil)
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error obteniendo suscripción: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error obteniendo suscripción - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var sub SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		t.Fatalf("❌ Error decodificando suscripción: %v", err)
	}

	return sub
}

// Helper function: getActiveSubscription
func getActiveSubscription(t *testing.T, token string, userID int) SubscriptionResponse {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/subscriptions/active/%d", subscriptionsAPIURL, userID), nil)
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error obteniendo suscripción activa: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error obteniendo suscripción activa - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var sub SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		t.Fatalf("❌ Error decodificando suscripción activa: %v", err)
	}

	return sub
}

// Helper function: createCashPayment
func createCashPayment(t *testing.T, token string, userID interface{}, subscriptionID string, amount float64) string {
	var userIDStr string
	switch v := userID.(type) {
	case int:
		userIDStr = fmt.Sprintf("%d", v)
	case uint:
		userIDStr = fmt.Sprintf("%d", v)
	case string:
		userIDStr = v
	default:
		userIDStr = fmt.Sprintf("%v", v)
	}

	reqBody := map[string]interface{}{
		"user_id":         userIDStr,
		"amount":          amount,
		"currency":        "ARS",
		"entity_type":     "subscription",
		"entity_id":       subscriptionID,
		"payment_gateway": "cash",
		"payment_method":  "cash",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", paymentsAPIURL+"/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error creando pago: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error creando pago - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var payment PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		t.Fatalf("❌ Error decodificando respuesta de pago: %v", err)
	}

	return payment.ID
}

// Helper function: getPayment
func getPayment(t *testing.T, paymentID string) PaymentResponse {
	resp, err := http.Get(paymentsAPIURL + "/payments/" + paymentID)
	if err != nil {
		t.Fatalf("❌ Error obteniendo pago: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error obteniendo pago - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var payment PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		t.Fatalf("❌ Error decodificando pago: %v", err)
	}

	return payment
}

// Helper function: updatePaymentStatus (alias de approvePayment)
func updatePaymentStatus(t *testing.T, adminToken, paymentID, status string) {
	reqBody := map[string]string{
		"status": status,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PATCH", paymentsAPIURL+"/payments/"+paymentID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error actualizando estado de pago: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error actualizando estado de pago - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}
}

// Helper function: approvePayment
func approvePayment(t *testing.T, adminToken, paymentID string) {
	updatePaymentStatus(t, adminToken, paymentID, "completed")
}

// Helper function: tryEnrollActivity
func tryEnrollActivity(t *testing.T, token string, userID, activityID int, expectError bool) error {
	reqBody := map[string]int{
		"usuario_id":   userID,
		"actividad_id": activityID,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", activitiesAPIURL+"/inscripciones", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error en inscripción: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if expectError {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil // No error when we expected one
		}
		var errResp ErrorResponse
		json.Unmarshal(bodyBytes, &errResp)
		return fmt.Errorf(errResp.Error)
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.Unmarshal(bodyBytes, &errResp)
		t.Fatalf("❌ Error inscribiéndose a actividad - Status: %d, Error: %s", resp.StatusCode, errResp.Error)
	}

	return nil
}

// Helper function: enrollActivity
func enrollActivity(t *testing.T, token string, userID, activityID int) InscripcionResponse {
	reqBody := map[string]int{
		"usuario_id":   userID,
		"actividad_id": activityID,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", activitiesAPIURL+"/inscripciones", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error en inscripción: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(bodyBytes, &errResp)
		t.Fatalf("❌ Error inscribiéndose a actividad - Status: %d, Error: %s, Details: %s", resp.StatusCode, errResp.Error, errResp.Details)
	}

	var inscripcion InscripcionResponse
	if err := json.NewDecoder(resp.Body).Decode(&inscripcion); err != nil {
		t.Fatalf("❌ Error decodificando inscripción: %v", err)
	}

	return inscripcion
}

// Helper function: unenrollActivity
func unenrollActivity(t *testing.T, token string, userID, activityID int) {
	reqBody := map[string]int{
		"usuario_id":   userID,
		"actividad_id": activityID,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("DELETE", activitiesAPIURL+"/inscripciones", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error en desinscripción: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(bodyBytes, &errResp)
		t.Fatalf("❌ Error desinscribiéndose de actividad - Status: %d, Error: %s, Details: %s", resp.StatusCode, errResp.Error, errResp.Details)
	}
}

// Helper function: listInscripciones
func listInscripciones(t *testing.T, token string, userID int) []InscripcionResponse {
	req, _ := http.NewRequest("GET", activitiesAPIURL+"/inscripciones", nil)
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error listando inscripciones: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error listando inscripciones - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var inscripciones []InscripcionResponse
	if err := json.NewDecoder(resp.Body).Decode(&inscripciones); err != nil {
		t.Fatalf("❌ Error decodificando lista de inscripciones: %v", err)
	}

	return inscripciones
}

// Helper function: activateSubscription - Activa manualmente una suscripción
func activateSubscription(t *testing.T, token, subscriptionID, paymentID string) {
	reqBody := map[string]interface{}{
		"estado":   "activa",
		"pago_id":  paymentID,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PATCH", subscriptionsAPIURL+"/subscriptions/"+subscriptionID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error activando suscripción: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error activando suscripción - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}
}

// Helper function: createActivity - Crea una nueva actividad
func createActivity(t *testing.T, adminToken, titulo, categoria string, sucursalID int) map[string]interface{} {
	activityReq := map[string]interface{}{
		"titulo":         titulo,
		"descripcion":    "Actividad de prueba para tests",
		"cupo":           10,
		"dia":            "Lunes",
		"horario_inicio": "10:00",
		"horario_final":  "11:00",
		"foto_url":       "https://images.unsplash.com/photo-1544367567-0f2fcb009e0b",
		"instructor":     "Test Instructor",
		"categoria":      categoria,
		"sucursal_id":    sucursalID,
	}
	body, _ := json.Marshal(activityReq)

	req, _ := http.NewRequest("POST", activitiesAPIURL+"/actividades", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ Error creando actividad: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("❌ Error creando actividad - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var activity map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		t.Fatalf("❌ Error decodificando actividad creada: %v", err)
	}

	return activity
}
