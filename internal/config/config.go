package config

import (
	"os"
)

type Config struct {
	MongoURI   string
	MongoDB    string
	JWTSecret  string
	Env        string
	ServerPort string
}

func Load() *Config {
	return &Config{
		MongoURI:   getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:    getEnv("MONGO_DB", "userdb"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me"),
		Env:        getEnv("ENV", "development"),
		ServerPort: getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
