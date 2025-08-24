package launcher

import (
	"context"
	"fmt"

	"testrunner/pkg/config"
	"testrunner/pkg/kube"
	"testrunner/pkg/logger"

	"github.com/google/uuid"
)

// Run executes the launcher with the given configuration
func Run(cfg config.Config) error {
	ctx := context.Background()

	if cfg.Debug {
		logger.SetGlobalLevel(logger.DEBUG)
	}

	client, err := kube.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	testNamespace := fmt.Sprintf("ket-%s", uuid.New().String()[:8])
	ns, err := kube.CreateNamespace(ctx, client, testNamespace)
	if err != nil {
		return fmt.Errorf("failed to create test namespace: %w", err)
	}

	logger.Info(logger.LAUNCHER, "Test namespace %s created", ns)

	jobManager := kube.NewJobManager(cfg)
	defer jobManager.Cleanup()

	if err := jobManager.StartMirrord(); err != nil {
		logger.Warn(logger.LAUNCHER, "Failed to start mirrord: %v", err)
		logger.Info(logger.LAUNCHER, "Continuing without traffic interception")
	} else {
		jobManager.StreamMirrordLogs()
	}

	job, err := kube.CreateJob(ctx, client, cfg, ns)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	logger.Info(logger.LAUNCHER, "Job %s created in namespace %s", job.Name, ns)

	if err := kube.StreamJobLogs(ctx, client, job, ns); err != nil {
		logger.Warn(logger.LAUNCHER, "Log stream failed: %v", err)
	}

	if err := kube.WaitForJobCompletion(ctx, client, job, ns); err != nil {
		logger.Error(logger.LAUNCHER, "Job failed: %v", err)
	}

	if err := kube.CopyTestResults(ctx, client, job, ns); err != nil {
		logger.Error(logger.LAUNCHER, "Failed to copy test results: %v", err)
	}

	if !cfg.KeepNamespace {
		logger.Info(logger.LAUNCHER, "Cleaning up test namespace %s", ns)
		if err := kube.DeleteNamespace(ctx, client, ns); err != nil {
			logger.Error(logger.LAUNCHER, "Failed to delete test namespace: %v", err)
		} else {
			logger.Info(logger.LAUNCHER, "Test namespace %s deleted", ns)
		}
	} else {
		logger.Info(logger.LAUNCHER, "Keeping test namespace %s as requested", ns)
	}

	return nil
}
