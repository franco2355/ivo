package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware maneja las políticas de CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if ctx.Request.Method == http.MethodOptions {
			ctx.Status(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
