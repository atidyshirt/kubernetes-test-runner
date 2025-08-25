package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type LogStreamer struct {
	logFiles struct {
		stdout string
		stderr string
	}
}

func NewLogStreamer() *LogStreamer {
	return &LogStreamer{}
}

func (ls *LogStreamer) StartMirrordLogStreaming(stdout, stderr *os.File) (stdoutPath, stderrPath string, err error) {
	ls.logFiles.stdout = stdout.Name()
	ls.logFiles.stderr = stderr.Name()

	UnifiedLog(INFO, "LOGGER", "Mirrord logs will be written to: %s (stdout), %s (stderr)", ls.logFiles.stdout, ls.logFiles.stderr)

	go ls.streamUnifiedLogs()
	return ls.logFiles.stdout, ls.logFiles.stderr, nil
}

func (ls *LogStreamer) GetLogFilePaths() (stdout, stderr string) {
	return ls.logFiles.stdout, ls.logFiles.stderr
}

func (ls *LogStreamer) GetLogContents() (stdout, stderr string, err error) {
	if ls.logFiles.stdout != "" {
		stdoutBytes, err := os.ReadFile(ls.logFiles.stdout)
		if err != nil {
			return "", "", fmt.Errorf("failed to read stdout log: %w", err)
		}
		stdout = string(stdoutBytes)
	}

	if ls.logFiles.stderr != "" {
		stderrBytes, err := os.ReadFile(ls.logFiles.stderr)
		if err != nil {
			return stdout, "", fmt.Errorf("failed to read stderr log: %w", err)
		}
		stderr = string(stderrBytes)
	}

	return stdout, stderr, nil
}

func (ls *LogStreamer) streamUnifiedLogs() {
	UnifiedLog(INFO, "LOGGER", "Starting unified log streaming...")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastStdoutSize := int64(0)
	lastStderrSize := int64(0)

	for {
		select {
		case <-ticker.C:
			ls.checkAndStreamNewLogs(&lastStdoutSize, &lastStderrSize)
		}
	}
}

func (ls *LogStreamer) checkAndStreamNewLogs(lastStdoutSize, lastStderrSize *int64) {
	if ls.logFiles.stdout != "" {
		if info, err := os.Stat(ls.logFiles.stdout); err == nil && info.Size() > *lastStdoutSize {
			ls.streamLogContent(ls.logFiles.stdout, lastStdoutSize, "MIRRORD.OUT")
		}
	}

	if ls.logFiles.stderr != "" {
		if info, err := os.Stat(ls.logFiles.stderr); err == nil && info.Size() > *lastStderrSize {
			ls.streamLogContent(ls.logFiles.stderr, lastStderrSize, "MIRRORD.ERR")
		}
	}
}

func (ls *LogStreamer) streamLogContent(logFile string, lastSize *int64, logType string) {
	file, err := os.Open(logFile)
	if err != nil {
		return
	}
	defer file.Close()

	file.Seek(*lastSize, 0)

	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil || n == 0 {
		return
	}

	content := string(buf[:n])
	*lastSize += int64(n)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(strings.ToLower(line), "mirrord") {
			UnifiedLog(INFO, logType, "%s", line)
		}
	}
}

func StreamPodLogs(lines []string) {
	for _, line := range lines {
		if line != "" {
			UnifiedLog(INFO, "POD", "%s", line)
		}
	}
}
