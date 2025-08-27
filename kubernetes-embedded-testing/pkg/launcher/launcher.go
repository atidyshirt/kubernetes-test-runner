package launcher

import (
	"context"
	"fmt"

	"testrunner/pkg/config"
	"testrunner/pkg/kube"
	"testrunner/pkg/logger"

	"github.com/google/uuid"
)

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

	job, err := kube.CreateJob(ctx, client, cfg, ns)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	if err := kube.StreamJobLogs(ctx, client, job, ns); err != nil {
		logger.LauncherLogger.Warn("Log stream failed: %v", err)
	}

	if err := kube.WaitForJobCompletion(ctx, client, job, ns); err != nil {
		logger.LauncherLogger.Error("Job failed: %v", err)
	}

	if err := kube.CopyTestResults(ctx, client, job, ns); err != nil {
		logger.LauncherLogger.Error("Failed to copy test results: %v", err)
	}

	if !cfg.KeepNamespace {
		logger.LauncherLogger.Info("Cleaning up test namespace %s", ns)
		if err := kube.DeleteNamespace(ctx, client, ns); err != nil {
			logger.LauncherLogger.Error("Failed to delete test namespace: %v", err)
		}
	}

	return nil
}
