package launcher

import (
	"testing"
	"testrunner/pkg/config"
)

func TestLauncher_ConfigValidation(t *testing.T) {
	cfg := config.Config{
		Mode:        "launch",
		ProjectRoot: "test-project",
		TestCommand: "echo 'test'",
	}

	if cfg.Mode != "launch" {
		t.Error("config mode should be 'launch'")
	}

	if cfg.TestCommand == "" {
		t.Error("test command should not be empty")
	}
}
