package launcher

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"testrunner/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLauncher_DryRunMode(t *testing.T) {
	cfg := config.Config{
		Mode:            "launch",
		ProjectRoot:     "test-project",
		Image:           "node:18-alpine",
		Debug:           false,
		TestCommand:     "npm test",
		KeepTestRunner:  false,
		BackoffLimit:    1,
		ActiveDeadlineS: 1800,
		WorkspacePath:   "/workspace",
		DryRun:          true,
	}

	// Capture stdout to verify output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	// Run in dry run mode
	err := Run(cfg)
	require.NoError(t, err)

	// Close pipe and read output
	w.Close()
	var output bytes.Buffer
	output.ReadFrom(r)
	outputStr := output.String()

	// Verify dry run messages
	assert.Contains(t, outputStr, "DRY RUN MODE - Generating manifests without applying")
	assert.Contains(t, outputStr, "DRY RUN COMPLETE - No resources were created")

	// Verify that manifests were generated
	assert.Contains(t, outputStr, "metadata:")
	assert.Contains(t, outputStr, "name: default")         // ServiceAccount
	assert.Contains(t, outputStr, "name: ket-test-runner") // Role
	assert.Contains(t, outputStr, "npm test")              // Test command
	assert.Contains(t, outputStr, "image: node:18-alpine") // Image
}

func TestLauncher_DryRunMode_ComplexCommand(t *testing.T) {
	cfg := config.Config{
		Mode:            "launch",
		ProjectRoot:     ".",
		Image:           "python:3.11",
		Debug:           false,
		TestCommand:     "pytest --verbose --cov=src --cov-report=html",
		KeepTestRunner:  false,
		BackoffLimit:    3,
		ActiveDeadlineS: 3600,
		WorkspacePath:   "/workspace",
		DryRun:          true,
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := Run(cfg)
	require.NoError(t, err)

	w.Close()
	var output bytes.Buffer
	output.ReadFrom(r)
	outputStr := output.String()

	// Verify complex command was preserved
	assert.Contains(t, outputStr, "pytest --verbose --cov=src --cov-report=html")

	// Verify configuration was applied
	assert.Contains(t, outputStr, "image: python:3.11")
	assert.Contains(t, outputStr, "backoffLimit: 3")
	assert.Contains(t, outputStr, "activeDeadlineSeconds: 3600")
}

func TestLauncher_DryRunMode_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
	}{
		{
			name: "empty test command",
			cfg: config.Config{
				Mode:        "launch",
				ProjectRoot: "test-project",
				Image:       "alpine:latest",
				TestCommand: "",
				DryRun:      true,
			},
			expectError: false,
		},
		{
			name: "special characters in test command",
			cfg: config.Config{
				Mode:        "launch",
				ProjectRoot: "test-project",
				Image:       "alpine:latest",
				TestCommand: "echo 'test with spaces' && npm test",
				DryRun:      true,
			},
			expectError: false,
		},
		{
			name: "very long test command",
			cfg: config.Config{
				Mode:        "launch",
				ProjectRoot: "test-project",
				Image:       "alpine:latest",
				TestCommand: strings.Repeat("echo test && ", 100) + "echo done",
				DryRun:      true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = originalStdout }()

			err := Run(tt.cfg)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify output contains the test command
			w.Close()
			var output bytes.Buffer
			output.ReadFrom(r)
			outputStr := output.String()

			// For very long commands, check that the key parts are present
			if strings.Contains(tt.cfg.TestCommand, "echo test") {
				assert.Contains(t, outputStr, "echo test")
				assert.Contains(t, outputStr, "echo done")
			} else {
				assert.Contains(t, outputStr, tt.cfg.TestCommand)
			}
			assert.Contains(t, outputStr, "DRY RUN COMPLETE - No resources were created")
		})
	}
}

func TestLauncher_DryRunMode_ProjectRootVariations(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		expectedJob string
	}{
		{
			name:        "simple directory",
			projectRoot: "my-project",
			expectedJob: "ket-my-project",
		},
		{
			name:        "nested path",
			projectRoot: "path/to/my-project",
			expectedJob: "ket-my-project",
		},
		{
			name:        "current directory",
			projectRoot: ".",
			expectedJob: "ket-", // Will be filled by os.Getwd()
		},
		{
			name:        "absolute path",
			projectRoot: "/absolute/path/to/project",
			expectedJob: "ket-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				Mode:        "launch",
				ProjectRoot: tt.projectRoot,
				Image:       "alpine:latest",
				TestCommand: "echo test",
				DryRun:      true,
			}

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = originalStdout }()

			err := Run(cfg)
			require.NoError(t, err)

			w.Close()
			var output bytes.Buffer
			output.ReadFrom(r)
			outputStr := output.String()

			if tt.projectRoot == "." {
				// For current directory, just verify it starts with ket-
				assert.Contains(t, outputStr, "name: ket-")
			} else {
				// For other cases, verify the exact job name
				assert.Contains(t, outputStr, "name: "+tt.expectedJob)
			}
		})
	}
}

func TestLauncher_DryRunMode_ConfigurationValidation(t *testing.T) {
	cfg := config.Config{
		Mode:            "launch",
		ProjectRoot:     "test-project",
		Image:           "custom-image:v1.2.3",
		Debug:           true,
		TestCommand:     "go test -v ./...",
		KeepTestRunner:  true,
		BackoffLimit:    5,
		ActiveDeadlineS: 7200,
		WorkspacePath:   "/custom/workspace",
		DryRun:          true,
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := Run(cfg)
	require.NoError(t, err)

	w.Close()
	var output bytes.Buffer
	output.ReadFrom(r)
	outputStr := output.String()

	// Verify all configuration was applied correctly
	assert.Contains(t, outputStr, "image: custom-image:v1.2.3")
	assert.Contains(t, outputStr, "go test -v ./...")
	assert.Contains(t, outputStr, "backoffLimit: 5")
	assert.Contains(t, outputStr, "activeDeadlineSeconds: 7200")
	assert.Contains(t, outputStr, "/custom/workspace")
}

func TestLauncher_DryRunMode_NamespaceGeneration(t *testing.T) {
	cfg := config.Config{
		Mode:        "launch",
		ProjectRoot: "test-project",
		Image:       "alpine:latest",
		TestCommand: "echo test",
		DryRun:      true,
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := Run(cfg)
	require.NoError(t, err)

	w.Close()
	var output bytes.Buffer
	output.ReadFrom(r)
	outputStr := output.String()

	// Verify namespace follows expected pattern
	lines := strings.Split(outputStr, "\n")
	var namespaceName string
	for _, line := range lines {
		if strings.Contains(line, "name: kubernetes-embedded-test-") {
			parts := strings.Split(line, ": ")
			if len(parts) == 2 {
				namespaceName = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	require.NotEmpty(t, namespaceName, "Namespace name not found in output")
	assert.Contains(t, namespaceName, "kubernetes-embedded-test-test-project-")
	assert.Len(t, namespaceName, len("kubernetes-embedded-test-test-project-")+8) // 8 char UUID
}
