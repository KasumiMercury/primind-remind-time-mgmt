package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
}

type LogConfig struct {
	Level string
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (*Config, error) {
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	readTimeout, err := time.ParseDuration(getEnv("SERVER_READ_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := time.ParseDuration(getEnv("SERVER_WRITE_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}

	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_DSN environment variable is required")
	}

	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         serverPort,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		Database: DatabaseConfig{
			DSN:             dsn,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
