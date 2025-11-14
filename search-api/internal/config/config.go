package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	SolrURL             string
	SolrCore            string
	RabbitMQURL         string
	RabbitMQExchange    string
	RabbitMQQueue       string
	MemcachedServers    []string
	CacheTTL            int
	LocalCacheTTL       int
	ActivitiesAPIURL    string
	SubscriptionsAPIURL string
	// MySQL config for fallback
	DBUser   string
	DBPass   string
	DBHost   string
	DBPort   string
	DBSchema string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	memcachedServers := strings.Split(getEnv("MEMCACHED_SERVERS", "localhost:11211"), ",")
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL", "60"))
	localCacheTTL, _ := strconv.Atoi(getEnv("LOCAL_CACHE_TTL", "30"))

	return &Config{
		Port:                getEnv("PORT", "8084"),
		SolrURL:             getEnv("SOLR_URL", "http://localhost:8983/solr/gym_activities"),
		SolrCore:            getEnv("SOLR_CORE", "gym_activities"),
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange:    getEnv("RABBITMQ_EXCHANGE", "gym_events"),
		RabbitMQQueue:       getEnv("RABBITMQ_QUEUE", "search_queue"),
		MemcachedServers:    memcachedServers,
		CacheTTL:            cacheTTL,
		LocalCacheTTL:       localCacheTTL,
		ActivitiesAPIURL:    getEnv("ACTIVITIES_API_URL", "http://localhost:8082"),
		SubscriptionsAPIURL: getEnv("SUBSCRIPTIONS_API_URL", "http://localhost:8081"),
		DBUser:              getEnv("DB_USER", "root"),
		DBPass:              getEnv("DB_PASS", "root"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "3306"),
		DBSchema:            getEnv("DB_SCHEMA", "gimnasio"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
