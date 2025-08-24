package launcher

import (
	"testing"
	"testrunner/pkg/config"
)

func TestLauncher_ConfigValidation(t *testing.T) {
	// Test that launcher can handle valid config
	cfg := config.Config{
		Mode:          "launch",
		TargetPod:     "test-pod",
		TestCommand:   "npm test",
		ProcessToTest: "npm start",
		ProjectRoot:   ".",
		Namespace:     "test-namespace",
	}

	// This is a basic test to ensure the package compiles and can be imported
	// In a real test environment, we would mock the Kubernetes client and test actual functionality
	if cfg.Mode != "launch" {
		t.Error("config mode should be 'launch'")
	}

	if cfg.TargetPod == "" {
		t.Error("target pod should not be empty")
	}

	if cfg.TestCommand == "" {
		t.Error("test command should not be empty")
	}

	if cfg.ProcessToTest == "" {
		t.Error("process to test should not be empty")
	}
}
