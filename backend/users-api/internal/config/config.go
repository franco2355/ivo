package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	MySQL      MySQLConfig
	JWT        JWTConfig
	RateLimit  RateLimitConfig
	Environment string
}

type MySQLConfig struct {
	User   string
	Pass   string
	Host   string
	Port   string
	Schema string
}

type JWTConfig struct {
	Secret string
}

type RateLimitConfig struct {
	LoginAttempts     int
	LoginWindow       int // minutes
	RegisterAttempts  int
	RegisterWindow    int // minutes
	PublicRPM         int // requests per minute
}

func Load() Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file")
	}

	return Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "local"),
		MySQL: MySQLConfig{
			User:   getEnv("DB_USER", "root"),
			Pass:   getEnv("DB_PASS", ""),
			Host:   getEnv("DB_HOST", "localhost"),
			Port:   getEnv("DB_PORT", "3306"),
			Schema: getEnv("DB_SCHEMA", "proyecto_integrador"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "my-secret-key"),
		},
		RateLimit: RateLimitConfig{
			LoginAttempts:    getEnvInt("RATE_LIMIT_LOGIN_ATTEMPTS", 10000),
			LoginWindow:      getEnvInt("RATE_LIMIT_LOGIN_WINDOW", 15),
			RegisterAttempts: getEnvInt("RATE_LIMIT_REGISTER_ATTEMPTS", 10000),
			RegisterWindow:   getEnvInt("RATE_LIMIT_REGISTER_WINDOW", 10),
			PublicRPM:        getEnvInt("RATE_LIMIT_PUBLIC_RPM", 10000),
		},
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
