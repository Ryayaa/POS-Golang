package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	JWTSecret string
	DBDSN     string // ganti dari DBpath ke DBDSN
}

func Load() (*Config, error) {
	godotenv.Load()

	config := &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "your_secret_key"),
		DBDSN:     getEnv("DB_DSN", "root:@tcp(localhost:3306)/pos_db?parseTime=true"),
	}
	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
