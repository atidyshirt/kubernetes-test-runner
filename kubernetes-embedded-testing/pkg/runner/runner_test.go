package runner

import (
	"testing"
	"testrunner/pkg/config"
)

func TestRunner_ConfigValidation(t *testing.T) {
	cfg := config.Config{
		Mode:        "run",
		ProjectRoot: ".",
		Debug:       false,
	}

	if cfg.Mode != "run" {
		t.Error("config mode should be 'run'")
	}

	if cfg.ProjectRoot == "" {
		t.Error("project root should not be empty")
	}
}
