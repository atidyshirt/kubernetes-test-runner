package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_New(t *testing.T) {
	logger := New(TESTRUNNER)
	if logger.component != TESTRUNNER {
		t.Errorf("expected component TESTRUNNER, got %v", logger.component)
	}
	if logger.level != INFO {
		t.Errorf("expected level INFO, got %v", logger.level)
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := New(TESTRUNNER)
	logger.SetLevel(ERROR)
	if logger.level != ERROR {
		t.Errorf("expected level ERROR, got %v", logger.level)
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

			originalOutput := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = originalOutput }()

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

			w.Close()
			var output bytes.Buffer
			output.ReadFrom(r)

			if tt.expected {
				if !strings.Contains(output.String(), "test message") {
					t.Errorf("expected message to be logged, but it wasn't")
				}
			} else {
				if strings.Contains(output.String(), "test message") {
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
		expected  Component
	}{
		{"LAUNCHER", LAUNCHER, LAUNCHER},
		{"KUBE", KUBE, KUBE},
		{"TESTRUNNER", TESTRUNNER, TESTRUNNER},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.component)
			if logger.component != tt.expected {
				t.Errorf("expected component %s, got %v", tt.expected, logger.component)
			}
		})
	}
}

func TestGlobalLoggers(t *testing.T) {
	if LauncherLogger.component != LAUNCHER {
		t.Errorf("expected LauncherLogger component LAUNCHER, got %v", LauncherLogger.component)
	}
	if KubeLogger.component != KUBE {
		t.Errorf("expected KubeLogger component KUBE, got %v", KubeLogger.component)
	}
	if TestRunnerLogger.component != TESTRUNNER {
		t.Errorf("expected TestRunnerLogger component TESTRUNNER, got %v", TestRunnerLogger.component)
	}
}

func TestSetGlobalLevel(t *testing.T) {
	SetGlobalLevel(DEBUG)
	if LauncherLogger.level != DEBUG {
		t.Errorf("expected LauncherLogger level DEBUG, got %v", LauncherLogger.level)
	}
	if KubeLogger.level != DEBUG {
		t.Errorf("expected KubeLogger level DEBUG, got %v", KubeLogger.level)
	}
	if TestRunnerLogger.level != DEBUG {
		t.Errorf("expected TestRunnerLogger level DEBUG, got %v", TestRunnerLogger.level)
	}

	SetGlobalLevel(INFO)
	if LauncherLogger.level != INFO {
		t.Errorf("expected LauncherLogger level INFO, got %v", LauncherLogger.level)
	}
}

func TestLogger_AllMethods(t *testing.T) {
	logger := New(TESTRUNNER)
	logger.SetLevel(DEBUG)

	originalOutput := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalOutput }()

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	w.Close()
	var output bytes.Buffer
	output.ReadFrom(r)

	outputStr := output.String()
	expectedMessages := []string{"debug message", "info message", "warn message", "error message"}

	for _, msg := range expectedMessages {
		if !strings.Contains(outputStr, msg) {
			t.Errorf("expected output to contain %q, but it didn't", msg)
		}
	}
}

func TestLogger_StreamLogs(t *testing.T) {
	logger := New(TESTRUNNER)
	logger.SetLevel(INFO)

	// Test with prefix and timestamp enabled
	logger.SetPrefix(true)
	logger.SetTimestamp(true)

	// Create a simple reader with test content
	content := "test line 1\ntest line 2\ntest line 3"
	reader := strings.NewReader(content)

	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	// Stream the logs
	logger.StreamLogs(reader)
	w.Close()

	// Read the output
	var output bytes.Buffer
	output.ReadFrom(r)
	outputStr := output.String()

	// Check that each line has the expected format
	lines := strings.Split(outputStr, "\n")
	assert.Len(t, lines, 4) // 3 lines + empty line at end

	// Check first line format: [timestamp] [INFO] [TESTRUNNER] test line 1
	assert.Contains(t, lines[0], "[INFO]")
	assert.Contains(t, lines[0], "[TESTRUNNER]")
	assert.Contains(t, lines[0], "test line 1")
	// Check that it contains a timestamp (should be in format [YYYY/MM/DD HH:MM:SS])
	assert.Regexp(t, `\[\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\]`, lines[0])

	// Test with prefix disabled
	logger.SetPrefix(false)
	content2 := "no prefix line"
	reader2 := strings.NewReader(content2)

	r2, w2, _ := os.Pipe()
	os.Stdout = w2

	logger.StreamLogs(reader2)
	w2.Close()

	var output2 bytes.Buffer
	output2.ReadFrom(r2)
	outputStr2 := output2.String()

	// Should not contain prefix components
	assert.NotContains(t, outputStr2, "[INFO]")
	assert.NotContains(t, outputStr2, "[TESTRUNNER]")
	assert.Contains(t, outputStr2, "no prefix line")
}
