package kube

import (
	"context"
	"fmt"

	"testrunner/pkg/config"
	"testrunner/pkg/logger"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateJob creates a new Kubernetes job for running tests
func CreateJob(ctx context.Context, client *kubernetes.Clientset, cfg config.Config, namespace string) (*batchv1.Job, error) {
	manifests, err := GenerateTestManifests(cfg, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate test manifests: %w", err)
	}

	logger.KubeLogger.Info("Creating job %s in namespace %s", manifests.Job.Name, namespace)
	createdJob, err := client.BatchV1().Jobs(namespace).Create(ctx, manifests.Job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	logger.KubeLogger.Info("Job created successfully: %s", createdJob.Name)
	return createdJob, nil
}
