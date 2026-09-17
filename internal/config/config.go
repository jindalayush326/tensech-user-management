package config

import (
	"os"
	"time"
)

type Config struct {
	Port        string
	MongoURI    string
	MongoDBName string
	RedisAddr   string
	RedisPass   string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
	CacheTTL    time.Duration
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "user_management"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		JWTSecret:   getEnv("JWT_SECRET", "supersecretkey"),
		JWTIssuer:   getEnv("JWT_ISSUER", "user-management-service"),
		JWTAudience: getEnv("JWT_AUDIENCE", "user-management-clients"),
		CacheTTL:    getDurationEnv("CACHE_TTL", 5*time.Minute),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return defaultValue
	}

	return duration
}