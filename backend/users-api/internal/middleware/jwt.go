package middleware

import (
	"net/http"
	"strings"
	"users-api/internal/services"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware valida el token JWT en el header Authorization
func JWTAuthMiddleware(userService services.UsersService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Obtener header Authorization
		auth := ctx.GetHeader("Authorization")
		if auth == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Validar formato "Bearer <token>"
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			return
		}

		// Validar token
		tokenClaims, err := userService.ValidateToken(parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"details": err.Error(),
			})
			return
		}

		// Extraer claims y guardar en contexto
		idUser, _ := tokenClaims["id_usuario"].(float64)
		isAdmin, _ := tokenClaims["is_admin"].(bool)
		username, _ := tokenClaims["username"].(string)

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
