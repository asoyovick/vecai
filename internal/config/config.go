// Package config loads runtime configuration from the environment.
package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port          string
	DatabasePath  string
	MigrationsDir string
	StaticDir     string
	UploadsDir    string
}

func Load() *Config {
	c := &Config{
		Port:          envOr("PORT", "8080"),
		DatabasePath:  envOr("DATABASE_PATH", filepath.Join("data", "vecai.db")),
		MigrationsDir: envOr("MIGRATIONS_DIR", "migrations"),
		StaticDir:     envOr("STATIC_DIR", "dist"),
	}
	if v := os.Getenv("UPLOADS_DIR"); v != "" {
		c.UploadsDir = v
	} else {
		c.UploadsDir = filepath.Join(filepath.Dir(c.DatabasePath), "uploads")
	}
	return c
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
