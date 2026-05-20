package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiID         int32
	ApiHash       string
	DbURL         string
	BotToken      string
	StringSession string
	GeminiKey     string
}

var Cfg *Config

func Load() (*Config, error) {
	_ = godotenv.Load()

	apiID, err := strconv.Atoi(getEnv("API_ID", ""))
	if err != nil {
		return nil, fmt.Errorf("API_ID must be a valid integer: %w", err)
	}

	cfg := &Config{
		ApiID:         int32(apiID),
		ApiHash:       requireEnv("API_HASH"),
		DbURL:         requireEnv("DB_URL"),
		BotToken:      requireEnv("TOKEN"),
		StringSession: getEnv("STRING_SESSION", ""),
		GeminiKey:     getEnv("API_KEY", ""),
	}

	if cfg.ApiHash == "" {
		return nil, fmt.Errorf("API_HASH is required")
	}
	if cfg.DbURL == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("TOKEN is required")
	}

	Cfg = cfg
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	return os.Getenv(key)
}
