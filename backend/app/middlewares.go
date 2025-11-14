package app

import (
	"net/http"
	"proyecto-integrador/services"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func JWTValidationMiddle(ctx *gin.Context) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "debes especificar el header 'Authorization' con tu token"})
		return
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}

	tokenClaims, err := services.UsuarioService.GetClaimsFromToken(parts[1])
	if err != nil {
		log.Debug(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}

	idUser, _ := tokenClaims["id_usuario"].(float64)
	isAdmin, _ := tokenClaims["is_admin"].(bool)

	ctx.Set("id_usuario", uint(idUser))
	ctx.Set("is_admin", isAdmin)
	ctx.Next()
}

func IsAdminMiddle(ctx *gin.Context) {
	isAdmin, exists := ctx.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para realizar esta acción"})
		return
	}

	ctx.Next()
}

func MapMiddewares() {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
}
