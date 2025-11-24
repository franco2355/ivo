package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Login
	loginBody := map[string]string{
		"username_or_email": "testuser",
		"password":          "password123",
	}
	body, _ := json.Marshal(loginBody)
	resp, _ := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(body))
	defer resp.Body.Close()

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	fmt.Printf("Token: %s\n\n", loginResp.Token)

	// Create subscription
	subBody := map[string]interface{}{
		"usuario_id":         "5",
		"plan_id":            "6923cc56ee6da85323ce5f48",
		"sucursal_origen_id": "1",
		"metodo_pago":        "cash",
		"auto_renovacion":    false,
	}
	body, _ = json.Marshal(subBody)
	fmt.Printf("Request body: %s\n\n", string(body))

	req, _ := http.NewRequest("POST", "http://localhost:8081/subscriptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)

	resp, _ = http.DefaultClient.Do(req)
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(bodyBytes))
}
