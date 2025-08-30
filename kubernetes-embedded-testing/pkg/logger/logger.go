package logger

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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
	component     Component
	level         LogLevel
	showPrefix    bool
	showTimestamp bool
}

func New(component Component) *Logger {
	return &Logger{
		component:     component,
		level:         INFO,
		showPrefix:    true,
		showTimestamp: true,
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) SetPrefix(show bool) {
	l.showPrefix = show
}

func (l *Logger) SetTimestamp(show bool) {
	l.showTimestamp = show
}

func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	var parts []string

	if l.showTimestamp {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	}

	if l.showPrefix {
		parts = append(parts, fmt.Sprintf("[%s]", level), fmt.Sprintf("[%s]", l.component))
	}

	message := fmt.Sprintf(format, args...)
	parts = append(parts, message)

	return strings.Join(parts, " ")
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

// StreamLogs streams logs from an io.Reader with the logger's prefix format
func (l *Logger) StreamLogs(reader io.Reader) {
	if l.level == SILENT {
		return
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		formattedLine := l.formatMessage(INFO, "%s", line)
		fmt.Println(formattedLine)
	}
}

func SetGlobalLevel(level LogLevel) {
	LauncherLogger.SetLevel(level)
	KubeLogger.SetLevel(level)
	TestRunnerLogger.SetLevel(level)
}

func SetGlobalPrefix(show bool) {
	LauncherLogger.SetPrefix(show)
	KubeLogger.SetPrefix(show)
	TestRunnerLogger.SetPrefix(show)
}

func SetGlobalTimestamp(show bool) {
	LauncherLogger.SetTimestamp(show)
	KubeLogger.SetTimestamp(show)
	TestRunnerLogger.SetTimestamp(show)
}

// ConfigureFromConfig configures all loggers based on the provided config
func ConfigureFromConfig(prefix, timestamp bool) {
	SetGlobalPrefix(prefix)
	SetGlobalTimestamp(timestamp)
}

var (
	LauncherLogger   = New(LAUNCHER)
	KubeLogger       = New(KUBE)
	TestRunnerLogger = New(TESTRUNNER)
)
