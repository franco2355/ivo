package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	// La misma secret key del .env
	jwtSecret := "MWGxVtrMibweh78/ESDdjTn1xJoJZSJn8LjKAhCpIyo="

	// Crear claims
	claims := Claims{
		UserID:   "test-user-123",
		Username: "testuser",
		Role:     "admin", // Cambia a "user" si prefieres
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Expira en 24 horas
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Crear token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Printf("Error generando token: %v\n", err)
		return
	}

	fmt.Println("Token JWT generado:")
	fmt.Println(tokenString)
	fmt.Println("\nPuedes usarlo así:")
	fmt.Printf("curl -X POST http://localhost:8081/plans \\\n")
	fmt.Printf("  -H \"Authorization: Bearer %s\" \\\n", tokenString)
	fmt.Printf("  -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("  -d '{...}'\n")
}
