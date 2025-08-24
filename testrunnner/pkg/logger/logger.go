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
	DEBUG LogLevel = iota
	INFO
	WARN
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
	LAUNCHER   Component = "LAUNCHER"
	KUBE       Component = "KUBE"
	RUNNER     Component = "RUNNER"
	MOCHA      Component = "MOCHA"
	NPM        Component = "NPM"
	MIRRORD    Component = "MIRRORD"
	TESTRUNNER Component = "TESTRUNNER"
	POD        Component = "POD"
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
		log.Printf(l.formatMessage(DEBUG, format, args...))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		log.Printf(l.formatMessage(INFO, format, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		log.Printf(l.formatMessage(WARN, format, args...))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		log.Printf(l.formatMessage(ERROR, format, args...))
	}
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	log.Printf(l.formatMessage(ERROR, format, args...))
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

// Convenience functions for quick logging
func Debug(component Component, format string, args ...interface{}) {
	New(component).Debug(format, args...)
}

func Info(component Component, format string, args ...interface{}) {
	New(component).Info(format, args...)
}

func Warn(component Component, format string, args ...interface{}) {
	New(component).Warn(format, args...)
}

func Error(component Component, format string, args ...interface{}) {
	New(component).Error(format, args...)
}
