package launcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/kube"
)

func Run(cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ActiveDeadlineS)*time.Second)
	defer cancel()

	if cfg.Debug {
		log.Printf("Starting testrunner in launch mode")
		log.Printf("Target pod: %s in namespace: %s", cfg.TargetPod, cfg.TargetNS)
		log.Printf("Process to test: %s", cfg.ProcessToTest)
		log.Printf("Test command: %s", cfg.TestCommand)
		log.Printf("Project root: %s", cfg.ProjectRoot)
	}

	client, err := kube.NewClient()
	if err != nil {
		return fmt.Errorf("failed to build kube client: %w", err)
	}

	// Create namespace for test
	ns, err := kube.CreateNamespace(ctx, client, cfg.Namespace)
	if err != nil {
		return fmt.Errorf("create namespace: %w", err)
	}
	log.Printf("Namespace %s created", ns)

	// Create job
	job, err := kube.CreateJob(ctx, client, cfg)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	log.Printf("Job %s created in namespace %s", job.Name, cfg.Namespace)

	// Stream logs until completion
	log.Printf("Starting log stream for job %s", job.Name)
	if err := kube.StreamJobLogs(ctx, client, job, cfg.Namespace); err != nil {
		log.Printf("Warning: log stream failed: %v", err)
	}

	// Wait for job completion
	log.Printf("Waiting for job %s to complete", job.Name)
	if err := kube.WaitForJobCompletion(ctx, client, job, cfg.Namespace); err != nil {
		log.Printf("Job failed: %v", err)
		// Don't return error here, let cleanup happen
	} else {
		log.Printf("Job %s completed successfully", job.Name)
	}

	// Copy test results before cleanup
	log.Printf("Retrieving test results...")
	if err := kube.CopyTestResults(ctx, client, job, cfg.Namespace); err != nil {
		log.Printf("Warning: failed to stream test results: %v", err)
	} else {
		log.Printf("Test results displayed above")
	}

	// Clean up namespace if not kept
	if !cfg.KeepNamespace {
		log.Printf("Cleaning up namespace %s", cfg.Namespace)
		if err := kube.DeleteNamespace(ctx, client, cfg.Namespace); err != nil {
			log.Printf("Warning: cleanup failed: %v", err)
		} else {
			log.Printf("Namespace %s deleted", cfg.Namespace)
		}
	} else {
		log.Printf("Keeping namespace %s as requested", cfg.Namespace)
	}

	return nil
}
