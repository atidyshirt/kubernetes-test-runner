package launcher

import (
	"testing"
	"testrunner/pkg/config"
)

func TestLauncher_ConfigValidation(t *testing.T) {
	cfg := config.Config{
		Mode:          "launch",
		TargetPod:     "test-pod",
		TestCommand:   "npm test",
		ProcessToTest: "npm start",
		ProjectRoot:   ".",
		Namespace:     "test-namespace",
	}

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
