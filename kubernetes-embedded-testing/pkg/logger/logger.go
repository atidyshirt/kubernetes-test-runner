package logger

import (
	"fmt"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	SILENT
)

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
	case SILENT:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

type Component string

const (
	LAUNCHER   Component = "LAUNCHER"
	KUBE       Component = "KUBE"
	TESTRUNNER Component = "TESTRUNNER"
)

type Logger struct {
	component Component
	level     LogLevel
}

func New(component Component) *Logger {
	return &Logger{
		component: component,
		level:     INFO,
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, l.component, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG && l.level != SILENT {
		fmt.Println(l.formatMessage(DEBUG, format, args...))
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO && l.level != SILENT {
		fmt.Println(l.formatMessage(INFO, format, args...))
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN && l.level != SILENT {
		fmt.Println(l.formatMessage(WARN, format, args...))
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR && l.level != SILENT {
		fmt.Println(l.formatMessage(ERROR, format, args...))
	}
}

func SetGlobalLevel(level LogLevel) {
	LauncherLogger.SetLevel(level)
	KubeLogger.SetLevel(level)
	TestRunnerLogger.SetLevel(level)
}

var (
	LauncherLogger   = New(LAUNCHER)
	KubeLogger       = New(KUBE)
	TestRunnerLogger = New(TESTRUNNER)
)
