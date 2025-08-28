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
	logger.KubeLogger.Info("Waiting for Test Runner Job %s to complete", job.Name)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.KubeLogger.Debug("Context cancelled, stopping wait for test completion")
			return ctx.Err()
		case <-ticker.C:
			currentJob, err := client.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				logger.KubeLogger.Warn("Failed to get job status: %v", err)
				continue
			}

			if currentJob.Status.Succeeded > 0 {
				logger.KubeLogger.Info("Test Runner Job %s completed successfully", job.Name)
				return nil
			}

			if currentJob.Status.Failed > 0 {
				logger.KubeLogger.Error("Test Runner Job %s failed", job.Name)
				return fmt.Errorf("Test Runner Job %s failed", job.Name)
			}

			logger.KubeLogger.Debug("Test Runner Job %s is still running...", job.Name)
		}
	}
}

// StreamTestOutputToHost streams the test output from the injected test runner pod back to the host machine
func StreamTestOutputToHost(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job, namespace string) error {
	logger.KubeLogger.Debug("Attempting to stream test output from job %s...", job.Name)

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for Test Runner Job %s to become active", job.Name)
		case <-ticker.C:
			currentJob, err := client.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				logger.KubeLogger.Warn("Failed to get Test Runner Job status: %v", err)
				continue
			}
			if currentJob.Status.Active > 0 {
				logger.KubeLogger.Debug("Test Runner Job %s is active, starting output stream", job.Name)
				goto StartStreaming
			}
		}
	}

StartStreaming:
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + job.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods for Test Runner Job: %w", err)
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for Test Runner Job %s", job.Name)
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

	done := make(chan error, 1)

	go func() {
		_, err := io.Copy(os.Stdout, stream)
		done <- err
	}()

	select {
	case <-ctx.Done():
		logger.KubeLogger.Debug("Context cancelled, stopping output stream")
		return ctx.Err()
	case err := <-done:
		return err
	}
}
