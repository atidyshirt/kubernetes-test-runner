package launcher

import (
	"context"
	"fmt"

	"testrunner/pkg/config"
	"testrunner/pkg/kube"
	"testrunner/pkg/logger"
)

// TestExecutionError represents a test execution failure with an exit code
type TestExecutionError struct {
	ExitCode int
	Message  string
}

func (e *TestExecutionError) Error() string {
	return fmt.Sprintf("%s (exit code: %d)", e.Message, e.ExitCode)
}

func Run(cfg config.Config) error {
	ctx := context.Background()
	if cfg.Ctx != nil {
		ctx = cfg.Ctx
	}

	if cfg.Debug {
		logger.SetGlobalLevel(logger.DEBUG)
	}

	client, err := kube.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	namespace := kube.GenerateTestNamespace(cfg.ProjectRoot)
	logger.LauncherLogger.Info("Using test namespace: %s", namespace)

	createdNamespace, err := kube.CreateNamespace(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	job, err := kube.CreateJob(ctx, client, cfg, createdNamespace)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Ensure cleanup on context cancellation
	defer func() {
		if !cfg.KeepTestRunner {
			logger.LauncherLogger.Info("Cleaning up test namespace %s", createdNamespace)
			if err := kube.ForceDeleteNamespace(context.Background(), client, createdNamespace); err != nil {
				logger.LauncherLogger.Error("Failed to delete test namespace: %v", err)
			}
		}
	}()

	go func() {
		if err := kube.StreamTestOutputToHost(ctx, client, job); err != nil {
			logger.LauncherLogger.Warn("Test output stream failed: %v", err)
		}
	}()

	testResult, err := kube.WaitForTestCompletion(ctx, client, job)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// If the test failed, return a special error that includes the exit code
	if !testResult.Success {
		return &TestExecutionError{
			ExitCode: testResult.ExitCode,
			Message:  "test execution failed",
		}
	}

	return nil
}
