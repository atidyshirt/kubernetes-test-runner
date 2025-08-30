package launcher

import (
	"context"
	"fmt"
	"os"

	"testrunner/pkg/config"
	"testrunner/pkg/kube"
	"testrunner/pkg/logger"

	"sigs.k8s.io/yaml"
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

	if cfg.DryRun {
		return runDryRun(cfg)
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

func runDryRun(cfg config.Config) error {
	// Set log level to SILENT in manifest mode for clean output
	if cfg.ManifestMode {
		logger.SetGlobalLevel(logger.SILENT)
	}

	logger.LauncherLogger.Info("DRY RUN MODE - Generating manifests without applying")

	namespace := kube.GenerateTestNamespace(cfg.ProjectRoot)
	logger.LauncherLogger.Info("Would create namespace: %s", namespace)

	// Generate all manifests
	manifests, err := kube.GenerateTestManifests(cfg, namespace)
	if err != nil {
		return fmt.Errorf("failed to generate test manifests: %w", err)
	}

	// Convert all manifests to YAML
	nsYAML, err := yaml.Marshal(manifests.Namespace)
	if err != nil {
		return fmt.Errorf("failed to marshal namespace to YAML: %w", err)
	}

	saYAML, err := yaml.Marshal(manifests.ServiceAccount)
	if err != nil {
		return fmt.Errorf("failed to marshal service account to YAML: %w", err)
	}

	roleYAML, err := yaml.Marshal(manifests.Role)
	if err != nil {
		return fmt.Errorf("failed to marshal role to YAML: %w", err)
	}

	roleBindingYAML, err := yaml.Marshal(manifests.RoleBinding)
	if err != nil {
		return fmt.Errorf("failed to marshal role binding to YAML: %w", err)
	}

	jobYAML, err := yaml.Marshal(manifests.Job)
	if err != nil {
		return fmt.Errorf("failed to marshal job to YAML: %w", err)
	}

	// Output all manifests to stdout
	fmt.Fprintf(os.Stdout, "---\n%s\n---\n%s\n---\n%s\n---\n%s\n---\n%s\n",
		string(nsYAML), string(saYAML), string(roleYAML), string(roleBindingYAML), string(jobYAML))

	logger.LauncherLogger.Info("DRY RUN COMPLETE - No resources were created")
	return nil
}
