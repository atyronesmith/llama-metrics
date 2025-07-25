package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_STRING",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_STRING",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "empty environment variable uses default",
			key:          "EMPTY_STRING",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := LoadEnvString(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("LoadEnvString() = %v, want %v", result, tt.expected)
			}

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

func TestLoadEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT",
			defaultValue: 5000,
			envValue:     "8080",
			expected:     8080,
		},
		{
			name:         "invalid integer falls back to default",
			key:          "TEST_INT_INVALID",
			defaultValue: 5000,
			envValue:     "invalid",
			expected:     5000,
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_INT",
			defaultValue: 5000,
			envValue:     "",
			expected:     5000,
		},
		{
			name:         "negative integer",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 5000,
			envValue:     "-1",
			expected:     -1,
		},
		{
			name:         "zero value",
			key:          "TEST_INT_ZERO",
			defaultValue: 5000,
			envValue:     "0",
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := LoadEnvInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("LoadEnvInt() = %v, want %v", result, tt.expected)
			}

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

func TestLoadEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "true value",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "false value",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "1 value",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "0 value",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "invalid value falls back to default",
			key:          "TEST_BOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true,
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_BOOL",
			defaultValue: false,
			envValue:     "",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := LoadEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("LoadEnvBool() = %v, want %v", result, tt.expected)
			}

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

func TestLoadEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "valid duration in seconds",
			key:          "TEST_DURATION_S",
			defaultValue: 30 * time.Second,
			envValue:     "60s",
			expected:     60 * time.Second,
		},
		{
			name:         "valid duration in minutes",
			key:          "TEST_DURATION_M",
			defaultValue: 30 * time.Second,
			envValue:     "1m",
			expected:     1 * time.Minute,
		},
		{
			name:         "valid duration in hours",
			key:          "TEST_DURATION_H",
			defaultValue: 30 * time.Second,
			envValue:     "2h",
			expected:     2 * time.Hour,
		},
		{
			name:         "invalid duration falls back to default",
			key:          "TEST_DURATION_INVALID",
			defaultValue: 30 * time.Second,
			envValue:     "invalid",
			expected:     30 * time.Second,
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_DURATION",
			defaultValue: 30 * time.Second,
			envValue:     "",
			expected:     30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := LoadEnvDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("LoadEnvDuration() = %v, want %v", result, tt.expected)
			}

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

func TestDefaultBaseConfig(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		envVars     map[string]string
		want        BaseConfig
	}{
		{
			name:        "default values",
			serviceName: "test-service",
			envVars:     map[string]string{},
			want: BaseConfig{
				ServiceName:  "test-service",
				Version:      "1.0.0",
				Port:         8080,
				MetricsPort:  8001,
				LogLevel:     "info",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				HealthPath:   "/health",
			},
		},
		{
			name:        "custom environment values",
			serviceName: "custom-service",
			envVars: map[string]string{
				"VERSION":       "2.0.0",
				"PORT":          "9000",
				"METRICS_PORT":  "9001",
				"LOG_LEVEL":     "debug",
				"READ_TIMEOUT":  "60s",
				"WRITE_TIMEOUT": "45s",
				"HEALTH_PATH":   "/healthz",
			},
			want: BaseConfig{
				ServiceName:  "custom-service",
				Version:      "2.0.0",
				Port:         9000,
				MetricsPort:  9001,
				LogLevel:     "debug",
				ReadTimeout:  60 * time.Second,
				WriteTimeout: 45 * time.Second,
				HealthPath:   "/healthz",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant environment variables
			envKeys := []string{"VERSION", "PORT", "METRICS_PORT", "LOG_LEVEL", "READ_TIMEOUT", "WRITE_TIMEOUT", "HEALTH_PATH"}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			result := DefaultBaseConfig(tt.serviceName)

			// Compare all fields
			if result.ServiceName != tt.want.ServiceName {
				t.Errorf("ServiceName = %v, want %v", result.ServiceName, tt.want.ServiceName)
			}
			if result.Version != tt.want.Version {
				t.Errorf("Version = %v, want %v", result.Version, tt.want.Version)
			}
			if result.Port != tt.want.Port {
				t.Errorf("Port = %v, want %v", result.Port, tt.want.Port)
			}
			if result.MetricsPort != tt.want.MetricsPort {
				t.Errorf("MetricsPort = %v, want %v", result.MetricsPort, tt.want.MetricsPort)
			}
			if result.LogLevel != tt.want.LogLevel {
				t.Errorf("LogLevel = %v, want %v", result.LogLevel, tt.want.LogLevel)
			}
			if result.ReadTimeout != tt.want.ReadTimeout {
				t.Errorf("ReadTimeout = %v, want %v", result.ReadTimeout, tt.want.ReadTimeout)
			}
			if result.WriteTimeout != tt.want.WriteTimeout {
				t.Errorf("WriteTimeout = %v, want %v", result.WriteTimeout, tt.want.WriteTimeout)
			}
			if result.HealthPath != tt.want.HealthPath {
				t.Errorf("HealthPath = %v, want %v", result.HealthPath, tt.want.HealthPath)
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestBaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: BaseConfig{
				ServiceName: "test-service",
				Port:        8080,
				MetricsPort: 8001,
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			config: BaseConfig{
				ServiceName: "",
				Port:        8080,
				MetricsPort: 8001,
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "invalid port - zero",
			config: BaseConfig{
				ServiceName: "test-service",
				Port:        0,
				MetricsPort: 8001,
			},
			wantErr: true,
			errMsg:  "invalid port number: 0",
		},
		{
			name: "invalid port - negative",
			config: BaseConfig{
				ServiceName: "test-service",
				Port:        -1,
				MetricsPort: 8001,
			},
			wantErr: true,
			errMsg:  "invalid port number: -1",
		},
		{
			name: "invalid port - too high",
			config: BaseConfig{
				ServiceName: "test-service",
				Port:        65536,
				MetricsPort: 8001,
			},
			wantErr: true,
			errMsg:  "invalid port number: 65536",
		},
		{
			name: "invalid metrics port",
			config: BaseConfig{
				ServiceName: "test-service",
				Port:        8080,
				MetricsPort: 0,
			},
			wantErr: true,
			errMsg:  "invalid metrics port number: 0",
		},
		{
			name: "same port and metrics port",
			config: BaseConfig{
				ServiceName: "test-service",
				Port:        8080,
				MetricsPort: 8080,
			},
			wantErr: true,
			errMsg:  "port and metrics port cannot be the same: 8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}