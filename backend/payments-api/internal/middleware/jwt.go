package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// JWTAuthMiddleware valida el token JWT en el header Authorization
// Nota: Este middleware NO valida si el usuario existe en la BD (eso lo hace users-api)
// Solo valida que el token sea válido y extrae los claims
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Obtener header Authorization
		auth := ctx.GetHeader("Authorization")
		fmt.Printf("[JWT DEBUG] Authorization header: %s\n", auth)
		if auth == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Validar formato "Bearer <token>"
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			fmt.Printf("[JWT DEBUG] Invalid header format. Parts: %v\n", parts)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			return
		}

		fmt.Printf("[JWT DEBUG] Token: %s\n", parts[1][:20]+"...")
		fmt.Printf("[JWT DEBUG] JWT Secret length: %d\n", len(jwtSecret))

		// Parsear y validar token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			// Validar el algoritmo
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			fmt.Printf("[JWT DEBUG] Token validation failed. Error: %v, Valid: %v\n", err, token.Valid)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"details": err.Error(),
			})
			return
		}

		// Extraer claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Guardar claims en contexto
		idUser, _ := claims["id_usuario"].(float64)
		isAdmin, _ := claims["is_admin"].(bool)
		username, _ := claims["username"].(string)

		ctx.Set("id_usuario", uint(idUser))
		ctx.Set("is_admin", isAdmin)
		ctx.Set("username", username)

		ctx.Next()
	}
}

// AdminOnlyMiddleware verifica que el usuario sea administrador
// Debe usarse DESPUÉS de JWTAuthMiddleware
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isAdmin, exists := ctx.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "You don't have permission to perform this action. Admin access required.",
			})
			return
		}

		ctx.Next()
	}
}
