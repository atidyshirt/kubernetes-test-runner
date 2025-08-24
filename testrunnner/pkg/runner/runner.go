package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/logger"
	"testrunner/pkg/report"
)

func Run(cfg config.Config) error {
	fmt.Println("==> Running inside runner pod")

	if cfg.Debug {
		logger.Debug(logger.RUNNER, "Runner configuration:")
		logger.Debug(logger.RUNNER, "  Test command: %s", cfg.TestCommand)
		logger.Debug(logger.RUNNER, "  Process to test: %s", cfg.ProcessToTest)
		logger.Debug(logger.RUNNER, "  Target pod: %s", cfg.TargetPod)
		logger.Debug(logger.RUNNER, "  Target namespace: %s", cfg.TargetNS)
		logger.Debug(logger.RUNNER, "  Working directory: %s", getWorkingDir())
	}

	// Check if we're in the right directory
	if err := os.Chdir("/workspace"); err != nil {
		return fmt.Errorf("failed to change to workspace directory: %w", err)
	}

	// List files in workspace for debugging
	if cfg.Debug {
		files, err := os.ReadDir(".")
		if err == nil {
			logger.Debug(logger.RUNNER, "Files in workspace:")
			for _, file := range files {
				logger.Debug(logger.RUNNER, "  %s", file.Name())
			}
		}
	}

	// Execute the test command
	result := executeTestCommand(cfg)

	// Write JSON report
	if err := writeReport(result); err != nil {
		logger.Warn(logger.RUNNER, "Failed to write report: %v", err)
	}

	// Exit with appropriate code
	if result.Success {
		fmt.Println("Tests completed successfully")
		return nil
	} else {
		fmt.Println("Tests failed")
		return fmt.Errorf("test execution failed: %s", result.Details)
	}
}

func executeTestCommand(cfg config.Config) report.Result {
	startTime := time.Now()

	// Split the test command into parts
	parts := strings.Fields(cfg.TestCommand)
	if len(parts) == 0 {
		return report.Result{
			Success:     false,
			Details:     "No test command provided",
			StartTime:   startTime,
			EndTime:     time.Now(),
			ExitCode:    -1,
			TestCommand: cfg.TestCommand,
			TargetPod:   cfg.TargetPod,
			TargetNS:    cfg.TargetNS,
		}
	}

	// Create command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = "/workspace"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"TARGET_NAMESPACE="+cfg.TargetNS,
		"TARGET_POD="+cfg.TargetPod,
		"PROCESS_TO_TEST="+cfg.ProcessToTest,
	)

	// Execute command
	logger.Info(logger.RUNNER, "Executing test command: %s", cfg.TestCommand)
	logger.Info(logger.RUNNER, "Working directory: %s", cmd.Dir)

	err := cmd.Run()
	endTime := time.Now()

	result := report.Result{
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
		TestCommand: cfg.TestCommand,
		TargetPod:   cfg.TargetPod,
		TargetNS:    cfg.TargetNS,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
			result.Details = fmt.Sprintf("Test command failed with exit code %d: %v", exitErr.ExitCode(), err)
		} else {
			result.ExitCode = -1
			result.Success = false
			result.Details = fmt.Sprintf("Failed to execute test command: %v", err)
		}
	} else {
		result.ExitCode = 0
		result.Success = true
		result.Details = "All tests passed successfully"
	}

	return result
}

func writeReport(result report.Result) error {
	// Ensure report directory exists
	reportDir := "/reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Write JSON report
	reportPath := filepath.Join(reportDir, "test-results.json")
	f, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	logger.Info(logger.RUNNER, "Report written to %s", reportPath)
	return nil
}

func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
