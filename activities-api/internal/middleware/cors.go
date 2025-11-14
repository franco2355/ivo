package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware maneja las políticas de CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.Request.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:5173"
		}
		// Permitir orígenes específicos
		allowedOrigins := map[string]bool{
			"http://localhost:5173": true,
			"http://localhost:3000": true,
		}

		if allowedOrigins[origin] {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Access-Control-Allow-Credentials", "true")
		} else {
			ctx.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		}

		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length")

		// Handle preflight requests
		if ctx.Request.Method == http.MethodOptions {
			ctx.Status(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
