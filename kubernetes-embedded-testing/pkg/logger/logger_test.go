package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogger_New(t *testing.T) {
	logger := New(TESTRUNNER)

	if logger.level != INFO {
		t.Errorf("expected level INFO, got %v", logger.level)
	}

	if logger.component != TESTRUNNER {
		t.Errorf("expected component TESTRUNNER, got %v", logger.component)
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := New(TESTRUNNER)

	logger.SetLevel(DEBUG)
	if logger.level != DEBUG {
		t.Errorf("expected level DEBUG, got %v", logger.level)
	}

	logger.SetLevel(ERROR)
	if logger.level != ERROR {
		t.Errorf("expected level ERROR, got %v", logger.level)
	}
}

func TestLogger_WithContext(t *testing.T) {
	logger := New(TESTRUNNER)

	ctxLogger := logger.WithContext(LAUNCHER)
	if ctxLogger.component != LAUNCHER {
		t.Errorf("expected component LAUNCHER, got %v", ctxLogger.component)
	}

	// Original logger should not be modified
	if logger.component != TESTRUNNER {
		t.Errorf("original logger component should not be modified, got %v", logger.component)
	}
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		testLevel LogLevel
		expected  bool
	}{
		{"DEBUG enabled at DEBUG", DEBUG, DEBUG, true},
		{"INFO enabled at DEBUG", DEBUG, INFO, true},
		{"WARN enabled at DEBUG", DEBUG, WARN, true},
		{"ERROR enabled at DEBUG", DEBUG, ERROR, true},
		{"DEBUG disabled at INFO", INFO, DEBUG, false},
		{"INFO enabled at INFO", INFO, INFO, true},
		{"WARN enabled at INFO", INFO, WARN, true},
		{"ERROR enabled at INFO", INFO, ERROR, true},
		{"DEBUG disabled at WARN", WARN, DEBUG, false},
		{"INFO disabled at WARN", WARN, INFO, false},
		{"WARN enabled at WARN", WARN, WARN, true},
		{"ERROR enabled at WARN", WARN, ERROR, true},
		{"DEBUG disabled at ERROR", ERROR, DEBUG, false},
		{"INFO disabled at ERROR", ERROR, INFO, false},
		{"WARN disabled at ERROR", ERROR, WARN, false},
		{"ERROR enabled at ERROR", ERROR, ERROR, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(TESTRUNNER)
			logger.SetLevel(tt.level)

			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			// Test with the appropriate log level method
			switch tt.testLevel {
			case DEBUG:
				logger.Debug("test message")
			case INFO:
				logger.Info("test message")
			case WARN:
				logger.Warn("test message")
			case ERROR:
				logger.Error("test message")
			}

			if tt.expected {
				if !strings.Contains(buf.String(), "test message") {
					t.Errorf("expected message to be logged, but it wasn't")
				}
			} else {
				if strings.Contains(buf.String(), "test message") {
					t.Errorf("expected message not to be logged, but it was")
				}
			}
		})
	}
}

func TestLogger_ComponentPrefixes(t *testing.T) {
	tests := []struct {
		name      string
		component Component
		expected  string
	}{
		{"LAUNCHER", LAUNCHER, "LAUNCHER"},
		{"KUBE", KUBE, "KUBE"},
		{"RUNNER", RUNNER, "RUNNER"},
		{"MOCHA", MOCHA, "MOCHA"},
		{"NPM", NPM, "NPM"},
		{"MIRRORD", MIRRORD, "MIRRORD"},
		{"TESTRUNNER", TESTRUNNER, "TESTRUNNER"},
		{"POD", POD, "POD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.component)

			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			logger.Info("test message")

			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("expected component prefix %s in log output, got: %s", tt.expected, buf.String())
			}
		})
	}
}

func TestGlobalLoggers(t *testing.T) {
	// Test that global loggers are properly initialized
	if LauncherLogger == nil {
		t.Error("LauncherLogger should not be nil")
	}

	if KubeLogger == nil {
		t.Error("KubeLogger should not be nil")
	}

	if RunnerLogger == nil {
		t.Error("RunnerLogger should not be nil")
	}

	// Test that they have the correct components
	if LauncherLogger.component != LAUNCHER {
		t.Errorf("LauncherLogger should have component LAUNCHER, got %v", LauncherLogger.component)
	}

	if KubeLogger.component != KUBE {
		t.Errorf("KubeLogger should have component KUBE, got %v", KubeLogger.component)
	}

	if RunnerLogger.component != RUNNER {
		t.Errorf("RunnerLogger should have component RUNNER, got %v", RunnerLogger.component)
	}
}

func TestSetGlobalLevel(t *testing.T) {
	// Test setting global level
	SetGlobalLevel(DEBUG)

	// All global loggers should now accept DEBUG level
	LauncherLogger.Debug("debug message")
	KubeLogger.Debug("debug message")
	RunnerLogger.Debug("debug message")

	// Reset to INFO for other tests
	SetGlobalLevel(INFO)
}

func TestLogger_ContextLogging(t *testing.T) {
	logger := New(TESTRUNNER).WithContext(LAUNCHER)

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected message to be logged, got: %s", output)
	}

	if !strings.Contains(output, "LAUNCHER") {
		t.Errorf("expected context to be included in log, got: %s", output)
	}
}

func TestLogger_AllMethods(t *testing.T) {
	logger := New(TESTRUNNER)

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test all logging methods
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	if !strings.Contains(output, "info message") {
		t.Error("Info message not found in output")
	}

	if !strings.Contains(output, "warn message") {
		t.Error("Warn message not found in output")
	}

	if !strings.Contains(output, "error message") {
		t.Error("Error message not found in output")
	}
}
