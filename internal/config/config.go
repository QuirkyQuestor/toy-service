package config

import (
	"context"
	"os"
	"strings"
)

type Config struct {
	DATABASE_URL string
}

// Pulls DB credentials from AWS Secrets Manager - maybe later
// if unavailable, read env vars to connect to the DB
func GetDBConfig(stopFunc context.CancelFunc) *Config {
	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		return &Config{
			DATABASE_URL: dbURL,
		}
	}
	stopFunc()
	return nil
}
