package launcher

import (
	"context"
	"fmt"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/kube"
	"testrunner/pkg/logger"
)

func Run(cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ActiveDeadlineS)*time.Second)
	defer cancel()

	// Set log level based on debug flag
	if cfg.Debug {
		logger.SetGlobalLevel(logger.DEBUG)
	}

	logger.Info(logger.LAUNCHER, "Starting testrunner in launch mode")
	logger.Debug(logger.LAUNCHER, "Target pod: %s in namespace: %s", cfg.TargetPod, cfg.TargetNS)
	logger.Debug(logger.LAUNCHER, "Process to test: %s", cfg.ProcessToTest)
	logger.Debug(logger.LAUNCHER, "Test command: %s", cfg.TestCommand)
	logger.Debug(logger.LAUNCHER, "Project root: %s", cfg.ProjectRoot)

	client, err := kube.NewClient()
	if err != nil {
		return fmt.Errorf("failed to build kube client: %w", err)
	}

	// Create namespace for test
	ns, err := kube.CreateNamespace(ctx, client, cfg.Namespace)
	if err != nil {
		return fmt.Errorf("create namespace: %w", err)
	}
	logger.Info(logger.LAUNCHER, "Namespace %s created", ns)

	// Create job
	job, err := kube.CreateJob(ctx, client, cfg)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	logger.Info(logger.LAUNCHER, "Job %s created in namespace %s", job.Name, cfg.Namespace)

	// Stream logs until completion
	logger.Info(logger.LAUNCHER, "Starting log stream for job %s", job.Name)
	if err := kube.StreamJobLogs(ctx, client, job, cfg.Namespace); err != nil {
		logger.Warn(logger.LAUNCHER, "Log stream failed: %v", err)
	}

	// Wait for job completion
	logger.Info(logger.LAUNCHER, "Waiting for job %s to complete", job.Name)
	if err := kube.WaitForJobCompletion(ctx, client, job, cfg.Namespace); err != nil {
		logger.Error(logger.LAUNCHER, "Job failed: %v", err)
		// Don't return error here, let cleanup happen
	} else {
		logger.Info(logger.LAUNCHER, "Job %s completed successfully", job.Name)
	}

	// Copy test results before cleanup
	logger.Info(logger.LAUNCHER, "Retrieving test results...")
	if err := kube.CopyTestResults(ctx, client, job, cfg.Namespace); err != nil {
		logger.Warn(logger.LAUNCHER, "Failed to stream test results: %v", err)
	} else {
		logger.Info(logger.LAUNCHER, "Test results displayed above")
	}

	// Clean up namespace if not kept
	if !cfg.KeepNamespace {
		logger.Info(logger.LAUNCHER, "Cleaning up namespace %s", cfg.Namespace)
		if err := kube.DeleteNamespace(ctx, client, cfg.Namespace); err != nil {
			logger.Warn(logger.LAUNCHER, "Cleanup failed: %v", err)
		} else {
			logger.Info(logger.LAUNCHER, "Namespace %s deleted", cfg.Namespace)
		}
	} else {
		logger.Info(logger.LAUNCHER, "Keeping namespace %s as requested", cfg.Namespace)
	}

	return nil
}
