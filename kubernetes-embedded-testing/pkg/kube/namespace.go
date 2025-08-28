package kube

import (
	"context"

	"testrunner/pkg/logger"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates a new Kubernetes namespace
func CreateNamespace(ctx context.Context, client *kubernetes.Clientset, name string) (string, error) {
	if name == "" {
		name = "testrunner"
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return name, nil
}

// DeleteNamespace deletes a Kubernetes namespace with graceful cleanup
func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset, name string) error {
	return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
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
