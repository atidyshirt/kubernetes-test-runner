package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DEBUG represents debug level logging
	DEBUG LogLevel = iota
	// INFO represents info level logging
	INFO
	// WARN represents warning level logging
	WARN
	// ERROR represents error level logging
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Component represents the source component of a log message
type Component string

const (
	// LAUNCHER represents the launcher component
	LAUNCHER Component = "LAUNCHER"
	// KUBE represents the Kubernetes component
	KUBE Component = "KUBE"
	// RUNNER represents the test runner component
	RUNNER Component = "RUNNER"
	// MOCHA represents the Mocha test framework component
	MOCHA Component = "MOCHA"
	// NPM represents the NPM package manager component
	NPM Component = "NPM"
	// MIRRORD represents the mirrord traffic interception component
	MIRRORD Component = "MIRRORD"
	// TESTRUNNER represents the test runner component
	TESTRUNNER Component = "TESTRUNNER"
	// POD represents the Kubernetes pod component
	POD Component = "POD"
)

// Logger provides structured logging with context
type Logger struct {
	component Component
	level     LogLevel
}

// New creates a new logger for a specific component
func New(component Component) *Logger {
	return &Logger{
		component: component,
		level:     INFO,
	}
}

// SetLevel sets the minimum log level for this logger
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// formatMessage formats a log message with timestamp, level, component, and message
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, l.component, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		log.Print(l.formatMessage(DEBUG, format, args...))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		log.Print(l.formatMessage(INFO, format, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		log.Print(l.formatMessage(WARN, format, args...))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		log.Print(l.formatMessage(ERROR, format, args...))
	}
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	log.Print(l.formatMessage(ERROR, format, args...))
	os.Exit(1)
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(component Component) *Logger {
	return &Logger{
		component: component,
		level:     l.level,
	}
}

// Global loggers for common components
var (
	LauncherLogger = New(LAUNCHER)
	KubeLogger     = New(KUBE)
	RunnerLogger   = New(RUNNER)
)

// SetGlobalLevel sets the log level for all global loggers
func SetGlobalLevel(level LogLevel) {
	LauncherLogger.SetLevel(level)
	KubeLogger.SetLevel(level)
	RunnerLogger.SetLevel(level)
}

// Debug logs a debug message for a component
func Debug(component Component, format string, args ...interface{}) {
	New(component).Debug(format, args...)
}

// Info logs an info message for a component
func Info(component Component, format string, args ...interface{}) {
	New(component).Info(format, args...)
}

// Warn logs a warning message for a component
func Warn(component Component, format string, args ...interface{}) {
	New(component).Warn(format, args...)
}

// Error logs an error message for a component
func Error(component Component, format string, args ...interface{}) {
	New(component).Error(format, args...)
}
