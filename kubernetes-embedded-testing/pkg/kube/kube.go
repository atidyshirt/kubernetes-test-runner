package kube

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"testrunner/pkg/logger"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitForTestCompletion waits for the injected test runner job to complete and returns the test results
func WaitForTestCompletion(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job, namespace string) error {
	logger.KubeLogger.Info("Waiting for test job %s to complete", job.Name)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// Check job status
			currentJob, err := client.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				logger.KubeLogger.Warn("Failed to get job status: %v", err)
				continue
			}

			if currentJob.Status.Succeeded > 0 {
				logger.KubeLogger.Info("Test job %s completed successfully", job.Name)
				return nil
			}

			if currentJob.Status.Failed > 0 {
				logger.KubeLogger.Error("Test job %s failed", job.Name)
				return fmt.Errorf("test job %s failed", job.Name)
			}

			logger.KubeLogger.Info("Test job %s is still running...", job.Name)
		}
	}
}

// StreamTestOutputToHost streams the test output from the injected test runner pod back to the host machine
func StreamTestOutputToHost(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job, namespace string) error {
	logger.KubeLogger.Info("Streaming test output from job %s back to host", job.Name)

	// Wait for job to start with timeout
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for test job %s to become active", job.Name)
		case <-ticker.C:
			currentJob, err := client.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				logger.KubeLogger.Warn("Failed to get job status: %v", err)
				continue
			}
			if currentJob.Status.Active > 0 {
				logger.KubeLogger.Warn("Test job %s is active, starting output stream", job.Name)
				goto StartStreaming
			}
		}
	}

StartStreaming:
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + job.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods for test job: %w", err)
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for test job %s", job.Name)
	}

	pod := pods.Items[0]
	logger.KubeLogger.Info("Streaming test output from pod %s", pod.Name)

	req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow: true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get test output stream: %w", err)
	}
	defer stream.Close()

	_, err = io.Copy(os.Stdout, stream)
	return err
}
