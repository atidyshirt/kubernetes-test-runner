package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"testrunner/pkg/config"
	"testrunner/pkg/logger"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateJob creates a new Kubernetes job for running tests
func CreateJob(ctx context.Context, client *kubernetes.Clientset, cfg config.Config, namespace string) (*batchv1.Job, error) {
	hostProjectRoot := filepath.Join(cfg.WorkspacePath, cfg.ProjectRoot)
	if cfg.ProjectRoot == "." {
		hostProjectRoot = cfg.WorkspacePath
	}

	workingDir, err := calculateWorkingDirectory(cfg.ProjectRoot, cfg.WorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate working directory: %w", err)
	}

	projectName := "project"
	if cfg.ProjectRoot == "." {
		if cwd, err := os.Getwd(); err == nil {
			projectName = filepath.Base(cwd)
		}
	} else {
		projectName = filepath.Base(cfg.ProjectRoot)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ket-%s", projectName),
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &cfg.BackoffLimit,
			ActiveDeadlineSeconds: &cfg.ActiveDeadlineS,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{
							Name: "source-code",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: hostProjectRoot,
									Type: &[]corev1.HostPathType{corev1.HostPathDirectory}[0],
								},
							},
						},
						{
							Name: "reports",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "test-runner",
							Image:           cfg.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-c",
								cfg.TestCommand,
							},
							WorkingDir: workingDir,
							Env: []corev1.EnvVar{
								{
									Name:  "KET_TEST_NAMESPACE",
									Value: namespace,
								},
								{
									Name:  "KET_PROJECT_ROOT",
									Value: cfg.ProjectRoot,
								},
								{
									Name:  "KET_WORKSPACE_PATH",
									Value: cfg.WorkspacePath,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source-code",
									MountPath: "/workspace",
								},
								{
									Name:      "reports",
									MountPath: "/reports",
								},
							},
						},
					},
				},
			},
		},
	}

	logger.KubeLogger.Info("Creating job %s in namespace %s", job.Name, namespace)
	createdJob, err := client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	logger.KubeLogger.Info("Job created successfully: %s", createdJob.Name)
	return createdJob, nil
}
