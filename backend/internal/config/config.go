// Package config loads the settings from environment variables (.env in dev).
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds every runtime setting of the API.
type Config struct {
	Env            string // "development" or "production"
	Port           string // Render injects PORT
	MongoURI       string // secret: never hard-code or commit it
	MongoDatabase  string
	AllowedOrigins []string // frontend URLs allowed by CORS
}

// IsProduction reports whether the API runs in production mode (APP_ENV=production).
func (c Config) IsProduction() bool { return c.Env == "production" }

// Load reads the configuration. A local ".env" file is loaded if it exists;
// real environment variables always win over it (godotenv never overrides).
func Load() (Config, error) {
	_ = godotenv.Load() // a missing .env is fine (e.g. on Render)

	cfg := Config{
		Env:            getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		MongoURI:       os.Getenv("MONGODB_URI"),
		MongoDatabase:  getEnv("MONGODB_DATABASE", "task_manager"),
		AllowedOrigins: splitCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:5173")),
	}

	if cfg.MongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI is required (see .env.example)")
	}
	return cfg, nil
}

// getEnv returns the trimmed value of an environment variable, or the
// fallback when the variable is empty or not set.
func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// splitCSV turns "a, b,c" into ["a" "b" "c"], dropping empty items.
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
