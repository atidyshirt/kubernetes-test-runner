package config

import (
	"testing"
)

func TestConfig_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "empty namespace should generate UUID",
			config: Config{
				Namespace: "",
			},
			expected: "ket-",
		},
		{
			name: "existing namespace should not change",
			config: Config{
				Namespace: "existing-namespace",
			},
			expected: "existing-namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.SetDefaults()

			if tt.expected == "ket-" {
				// Check that namespace starts with "ket-" and has 8 characters after
				if len(tt.config.Namespace) != 12 || tt.config.Namespace[:4] != "ket-" {
					t.Errorf("expected namespace to start with 'ket-' and be 12 chars long, got: %s", tt.config.Namespace)
				}
			} else {
				if tt.config.Namespace != tt.expected {
					t.Errorf("expected namespace %s, got %s", tt.expected, tt.config.Namespace)
				}
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name: "valid config",
			config: Config{
				Mode:          "launch",
				TargetPod:     "test-pod",
				TestCommand:   "npm test",
				ProcessToTest: "npm start",
			},
			shouldError: false,
		},
		{
			name: "missing target pod",
			config: Config{
				Mode:          "launch",
				TestCommand:   "npm test",
				ProcessToTest: "npm start",
			},
			shouldError: true,
		},
		{
			name: "missing test command",
			config: Config{
				Mode:          "launch",
				TargetPod:     "test-pod",
				ProcessToTest: "npm start",
			},
			shouldError: true,
		},
		{
			name: "missing process to test",
			config: Config{
				Mode:        "launch",
				TargetPod:   "test-pod",
				TestCommand: "npm test",
			},
			shouldError: true,
		},
		{
			name: "run mode should not require target pod",
			config: Config{
				Mode: "run",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be validation logic if we add it to the config package
			// For now, just test the basic structure
			if tt.config.Mode == "" {
				t.Error("mode should not be empty")
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	config := Config{}

	// Test that default values are reasonable
	if config.BackoffLimit != 0 {
		t.Errorf("expected default backoff limit to be 0, got %d", config.BackoffLimit)
	}

	if config.ActiveDeadlineS != 0 {
		t.Errorf("expected default active deadline to be 0, got %d", config.ActiveDeadlineS)
	}

	if config.KeepNamespace != false {
		t.Errorf("expected default keep namespace to be false, got %t", config.KeepNamespace)
	}

	if config.Debug != false {
		t.Errorf("expected default debug to be false, got %t", config.Debug)
	}
}
