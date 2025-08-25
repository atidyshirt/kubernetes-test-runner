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
	default:
		return "UNKNOWN"
	}
}

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
	if l.level <= DEBUG {
		fmt.Println(l.formatMessage(DEBUG, format, args...))
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		fmt.Println(l.formatMessage(INFO, format, args...))
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		fmt.Println(l.formatMessage(WARN, format, args...))
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		fmt.Println(l.formatMessage(ERROR, format, args...))
	}
}

func SetGlobalLevel(level LogLevel) {
	LauncherLogger.SetLevel(level)
	KubeLogger.SetLevel(level)
	RunnerLogger.SetLevel(level)
}

var (
	LauncherLogger = New(LAUNCHER)
	KubeLogger     = New(KUBE)
	RunnerLogger   = New(RUNNER)
)

func UnifiedLog(level LogLevel, component, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	levelStr := level.String()

	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] [%s] [%s] %s\n", timestamp, levelStr, component, message)
}
