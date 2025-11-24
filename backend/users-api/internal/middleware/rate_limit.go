package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter maneja el límite de requests por IP
type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	max     int           // máximo de requests
	window  time.Duration // ventana de tiempo
}

// ClientInfo almacena la info de requests por cliente
type ClientInfo struct {
	requests  []time.Time
	lastClean time.Time
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(maxRequests int, windowMinutes int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		max:     maxRequests,
		window:  time.Duration(windowMinutes) * time.Minute,
	}

	// Cleanup goroutine para limpiar clientes viejos cada 5 minutos
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Allow verifica si un cliente puede hacer un request
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Obtener o crear info del cliente
	client, exists := rl.clients[clientIP]
	if !exists {
		client = &ClientInfo{
			requests:  []time.Time{},
			lastClean: now,
		}
		rl.clients[clientIP] = client
	}

	// Filtrar requests dentro de la ventana de tiempo
	validRequests := []time.Time{}
	for _, reqTime := range client.requests {
		if now.Sub(reqTime) < rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Si está dentro del límite, permitir
	if len(validRequests) < rl.max {
		client.requests = append(validRequests, now)
		client.lastClean = now
		return true
	}

	// Actualizar lista de requests (aunque no permitimos este)
	client.requests = validRequests
	return false
}

// cleanup elimina clientes que no han hecho requests recientemente
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, client := range rl.clients {
		// Si no ha hecho requests en 2x la ventana, eliminar
		if now.Sub(client.lastClean) > (rl.window * 2) {
			delete(rl.clients, ip)
		}
	}
}

// Middleware para rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware crea un middleware de rate limiting genérico
func RateLimitMiddleware(maxRequests int, windowMinutes int) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, windowMinutes)
	return limiter.Middleware()
}
