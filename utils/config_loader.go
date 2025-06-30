package utils

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads configuration from .env file and environment variables
// Environment variables take precedence over .env file values
func LoadConfig() error {
	// Try to load .env file (optional)
	_ = godotenv.Load()
	return nil
}

//封装环境变量获取
func GetConfig(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
