package config

import (
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	GoogleClientID    string
	GoogleSecret      string
	GoogleRedirectURL string
	FrontendURL       string
	WhatsAppMode      string // "baileys" or "mock"
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "./reyna.db"),
		JWTSecret:         getEnv("JWT_SECRET", "reyna-dev-secret-change-in-prod"),
		GoogleClientID:    getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL: getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:5173"),
		WhatsAppMode:      getEnv("WHATSAPP_MODE", "mock"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
