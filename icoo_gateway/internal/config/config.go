package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	StorageDriver   string
	DatabaseURL     string
	SQLitePath      string
}

func Load(workdir string) (Config, error) {
	if err := loadDotEnv(filepath.Join(workdir, ".env")); err != nil {
		return Config{}, err
	}
	cfg := Config{
		Host:            strings.TrimSpace(os.Getenv("GATEWAY_HOST")),
		Port:            intFromEnv("GATEWAY_PORT", 18080),
		ReadTimeout:     durationFromEnv("GATEWAY_READ_TIMEOUT_SECONDS", 10*time.Second),
		WriteTimeout:    durationFromEnv("GATEWAY_WRITE_TIMEOUT_SECONDS", 15*time.Second),
		ShutdownTimeout: durationFromEnv("GATEWAY_SHUTDOWN_TIMEOUT_SECONDS", 10*time.Second),
		StorageDriver:   strings.TrimSpace(os.Getenv("GATEWAY_STORAGE_DRIVER")),
		DatabaseURL:     strings.TrimSpace(os.Getenv("GATEWAY_DATABASE_URL")),
		SQLitePath:      strings.TrimSpace(os.Getenv("GATEWAY_SQLITE_PATH")),
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.StorageDriver == "" {
		cfg.StorageDriver = "memory"
	}
	if cfg.SQLitePath == "" {
		cfg.SQLitePath = filepath.Join(workdir, "data", "icoo_gateway.db")
	}
	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

func intFromEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	seconds := intFromEnv(key, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}
