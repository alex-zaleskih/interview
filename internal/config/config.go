package config

import (
	"os"
	"time"
)

// Config holds application settings. All values can be overridden with environment variables.
type Config struct {
	HTTPAddr      string
	MongoURI      string
	MongoDatabase string
	SessionSecret string
	SessionTTL    time.Duration
	WebDir        string
}

func Load() Config {
	return Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DB", "shop"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-secret-change-me"),
		SessionTTL:    7 * 24 * time.Hour,
		WebDir:        getEnv("WEB_DIR", "web"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
