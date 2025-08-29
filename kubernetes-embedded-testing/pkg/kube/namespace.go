package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"testrunner/pkg/logger"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GenerateTestNamespace generates a unique test namespace name using UUID
func GenerateTestNamespace(projectRoot string) string {
	var basename string

	if projectRoot == "." {
		if cwd, err := os.Getwd(); err == nil {
			basename = filepath.Base(cwd)
		} else {
			basename = "default"
		}
	} else {
		basename = filepath.Base(projectRoot)
	}

	// Clean the basename to make it safe for namespace names
	// Replace any non-alphanumeric characters with hyphens
	cleanName := ""
	for _, r := range basename {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			cleanName += string(r)
		} else {
			cleanName += "-"
		}
	}

	// Remove multiple consecutive hyphens
	for i := 0; i < len(cleanName)-1; i++ {
		if cleanName[i] == '-' && cleanName[i+1] == '-' {
			cleanName = cleanName[:i] + cleanName[i+1:]
			i--
		}
	}

	// Remove leading/trailing hyphens
	if len(cleanName) > 0 && cleanName[0] == '-' {
		cleanName = cleanName[1:]
	}
	if len(cleanName) == 0 {
		cleanName = "default"
	}

	namespaceUUID := uuid.New().String()[:8]
	return fmt.Sprintf("kubernetes-embedded-test-%s-%s", cleanName, namespaceUUID)
}

// CreateNamespace creates a new Kubernetes namespace
func CreateNamespace(ctx context.Context, client *kubernetes.Clientset, name string) (string, error) {
	if name == "" {
		name = GenerateTestNamespace("default")
	}

	logger.KubeLogger.Info("Creating test namespace: %s", name)

	_, err := client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create namespace %s: %w", name, err)
	}

	logger.KubeLogger.Info("Successfully created test namespace: %s", name)
	return name, nil
}

// DeleteTestRunnerPod deletes the Kubernetes test runner pod
func DeleteTestRunnerPod(ctx context.Context, client *kubernetes.Clientset, name string) error {
	return client.CoreV1().Pods(name).Delete(ctx, name, metav1.DeleteOptions{})
}

// ForceDeleteNamespace aggressively deletes a namespace with no grace period
func ForceDeleteNamespace(ctx context.Context, client *kubernetes.Clientset, name string) error {
	logger.KubeLogger.Info("Force deleting namespace %s with no grace period", name)

	// First try to remove any finalizers that might be blocking deletion
	ns, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil && len(ns.Spec.Finalizers) > 0 {
		logger.KubeLogger.Info("Removing finalizers from namespace %s", name)
		ns.Spec.Finalizers = []corev1.FinalizerName{}
		_, err = client.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
		if err != nil {
			logger.KubeLogger.Warn("Failed to remove finalizers: %v", err)
		}
	}

	// Force delete with no grace period and background propagation
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &[]int64{0}[0],
		PropagationPolicy:  &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
	}

	return client.CoreV1().Namespaces().Delete(ctx, name, deleteOptions)
}
