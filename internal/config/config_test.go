package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/config"
)

func clearEnvVars(t *testing.T) {
	t.Helper()

	envVars := []string{
		"SERVER_HOST",
		"SERVER_PORT",
		"SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT",
		"POSTGRES_DSN",
		"DB_MAX_OPEN_CONNS",
		"DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}

func TestLoadSuccess(t *testing.T) {
	tests := []struct {
		name                    string
		envVars                 map[string]string
		expectedHost            string
		expectedPort            int
		expectedReadTimeout     time.Duration
		expectedWriteTimeout    time.Duration
		expectedDSN             string
		expectedMaxOpenConns    int
		expectedMaxIdleConns    int
		expectedConnMaxLifetime time.Duration
	}{
		{
			name: "all values from environment",
			envVars: map[string]string{
				"SERVER_HOST":          "localhost",
				"SERVER_PORT":          "3000",
				"SERVER_READ_TIMEOUT":  "60s",
				"SERVER_WRITE_TIMEOUT": "60s",
				"POSTGRES_DSN":         "postgres://user:pass@localhost:5432/db",
				"DB_MAX_OPEN_CONNS":    "50",
				"DB_MAX_IDLE_CONNS":    "10",
				"DB_CONN_MAX_LIFETIME": "10m",
			},
			expectedHost:            "localhost",
			expectedPort:            3000,
			expectedReadTimeout:     60 * time.Second,
			expectedWriteTimeout:    60 * time.Second,
			expectedDSN:             "postgres://user:pass@localhost:5432/db",
			expectedMaxOpenConns:    50,
			expectedMaxIdleConns:    10,
			expectedConnMaxLifetime: 10 * time.Minute,
		},
		{
			name: "default values except required DSN",
			envVars: map[string]string{
				"POSTGRES_DSN": "postgres://user:pass@localhost:5432/db",
			},
			expectedHost:            "0.0.0.0",
			expectedPort:            8080,
			expectedReadTimeout:     30 * time.Second,
			expectedWriteTimeout:    30 * time.Second,
			expectedDSN:             "postgres://user:pass@localhost:5432/db",
			expectedMaxOpenConns:    25,
			expectedMaxIdleConns:    25,
			expectedConnMaxLifetime: 5 * time.Minute,
		},
		{
			name: "partial custom values",
			envVars: map[string]string{
				"SERVER_PORT":  "9000",
				"POSTGRES_DSN": "postgres://localhost/testdb",
			},
			expectedHost:            "0.0.0.0",
			expectedPort:            9000,
			expectedReadTimeout:     30 * time.Second,
			expectedWriteTimeout:    30 * time.Second,
			expectedDSN:             "postgres://localhost/testdb",
			expectedMaxOpenConns:    25,
			expectedMaxIdleConns:    25,
			expectedConnMaxLifetime: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			defer clearEnvVars(t)

			cfg, err := config.Load()

			require.NoError(t, err)
			assert.Equal(t, tt.expectedHost, cfg.Server.Host)
			assert.Equal(t, tt.expectedPort, cfg.Server.Port)
			assert.Equal(t, tt.expectedReadTimeout, cfg.Server.ReadTimeout)
			assert.Equal(t, tt.expectedWriteTimeout, cfg.Server.WriteTimeout)
			assert.Equal(t, tt.expectedDSN, cfg.Database.DSN)
			assert.Equal(t, tt.expectedMaxOpenConns, cfg.Database.MaxOpenConns)
			assert.Equal(t, tt.expectedMaxIdleConns, cfg.Database.MaxIdleConns)
			assert.Equal(t, tt.expectedConnMaxLifetime, cfg.Database.ConnMaxLifetime)
		})
	}
}

func TestLoadError(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedErr string
	}{
		{
			name:        "missing POSTGRES_DSN",
			envVars:     map[string]string{},
			expectedErr: "POSTGRES_DSN environment variable is required",
		},
		{
			name: "invalid SERVER_PORT",
			envVars: map[string]string{
				"SERVER_PORT":  "not-a-number",
				"POSTGRES_DSN": "postgres://localhost/db",
			},
			expectedErr: "invalid SERVER_PORT",
		},
		{
			name: "invalid SERVER_READ_TIMEOUT",
			envVars: map[string]string{
				"SERVER_READ_TIMEOUT": "invalid",
				"POSTGRES_DSN":        "postgres://localhost/db",
			},
			expectedErr: "invalid SERVER_READ_TIMEOUT",
		},
		{
			name: "invalid SERVER_WRITE_TIMEOUT",
			envVars: map[string]string{
				"SERVER_WRITE_TIMEOUT": "invalid",
				"POSTGRES_DSN":         "postgres://localhost/db",
			},
			expectedErr: "invalid SERVER_WRITE_TIMEOUT",
		},
		{
			name: "invalid DB_MAX_OPEN_CONNS",
			envVars: map[string]string{
				"DB_MAX_OPEN_CONNS": "not-a-number",
				"POSTGRES_DSN":      "postgres://localhost/db",
			},
			expectedErr: "invalid DB_MAX_OPEN_CONNS",
		},
		{
			name: "invalid DB_MAX_IDLE_CONNS",
			envVars: map[string]string{
				"DB_MAX_IDLE_CONNS": "not-a-number",
				"POSTGRES_DSN":      "postgres://localhost/db",
			},
			expectedErr: "invalid DB_MAX_IDLE_CONNS",
		},
		{
			name: "invalid DB_CONN_MAX_LIFETIME",
			envVars: map[string]string{
				"DB_CONN_MAX_LIFETIME": "invalid",
				"POSTGRES_DSN":         "postgres://localhost/db",
			},
			expectedErr: "invalid DB_CONN_MAX_LIFETIME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			defer clearEnvVars(t)

			_, err := config.Load()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestServerConfigAddressSuccess(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{
			name:     "default address",
			host:     "0.0.0.0",
			port:     8080,
			expected: "0.0.0.0:8080",
		},
		{
			name:     "localhost address",
			host:     "localhost",
			port:     3000,
			expected: "localhost:3000",
		},
		{
			name:     "custom host and port",
			host:     "192.168.1.100",
			port:     9000,
			expected: "192.168.1.100:9000",
		},
		{
			name:     "empty host",
			host:     "",
			port:     8080,
			expected: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverConfig := config.ServerConfig{
				Host: tt.host,
				Port: tt.port,
			}

			result := serverConfig.Address()

			assert.Equal(t, tt.expected, result)
		})
	}
}
