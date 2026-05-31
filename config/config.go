package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	Environment      string
	PaystackSecret   string
}

func Load() *Config {
	godotenv.Load()

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" || jwtSecret == "change-me" {
		env := getEnv("ENVIRONMENT", "development")
		if env == "production" {
			log.Fatal("JWT_SECRET must be set to a strong secret in production")
		}
		log.Println("[warn] JWT_SECRET not set — using insecure default (development only)")
		jwtSecret = "dev-insecure-secret-change-me"
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:      jwtSecret,
		Environment:    getEnv("ENVIRONMENT", "development"),
		PaystackSecret: getEnv("PAYSTACK_SECRET_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
