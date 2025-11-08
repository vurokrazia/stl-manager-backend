package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	SupabaseURL      string
	SupabaseAnonKey  string
	OpenAIAPIKey     string
	RedisAddr        string
	RedisUsername    string
	RedisPassword    string
	RedisDB          int
	ScanRootDir      string
	SupportedExts    []string
	APIKey           string
	Port             string
}

func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	cfg := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		SupabaseURL:     getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey: getEnv("SUPABASE_ANON_KEY", ""),
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		RedisAddr:       getEnv("REDIS_ADDR", ""),
		RedisUsername:   getEnv("REDIS_USERNAME", "default"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         redisDB,
		ScanRootDir:     getEnv("SCAN_ROOT_DIR", "E:\\Impresion3D"),
		SupportedExts:   parseExts(getEnv("SUPPORTED_EXTS", ".stl,.zip,.rar")),
		APIKey:          getEnv("API_KEY", "dev-secret-key"),
		Port:            getEnv("PORT", "8080"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.ScanRootDir == "" {
		return fmt.Errorf("SCAN_ROOT_DIR is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseExts(exts string) []string {
	parts := strings.Split(exts, ",")
	result := make([]string, 0, len(parts))
	for _, ext := range parts {
		trimmed := strings.TrimSpace(ext)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
