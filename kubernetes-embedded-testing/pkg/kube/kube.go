package kube

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/logger"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		cfg, err = kubeconfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(cfg)
}

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

func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset, name string) error {
	return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

func CreateJob(ctx context.Context, client *kubernetes.Clientset, cfg config.Config, namespace string) (*batchv1.Job, error) {
	hostProjectRoot := filepath.Join(cfg.KindWorkspacePath, cfg.ProjectRoot)
	if cfg.ProjectRoot == "." {
		hostProjectRoot = cfg.KindWorkspacePath
	}

	workingDir, err := calculateWorkingDirectory(cfg.ProjectRoot, cfg.KindWorkspacePath)
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
							Name:            "testrunner",
							Image:           cfg.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/sh", "-c"},
							Args:            []string{cfg.TestCommand},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source-code",
									MountPath: cfg.KindWorkspacePath,
								},
								{
									Name:      "reports",
									MountPath: "/reports",
								},
							},
							WorkingDir: workingDir,
						},
					},
				},
			},
		},
	}

	return client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
}

func StreamJobLogs(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job, namespace string) error {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
	})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for job %s", job.Name)
	}
	pod := pods.Items[0]

	req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow: true,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	buf := make([]byte, 2000)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	logger.KubeLogger.Info("Finished streaming logs from %s", pod.Name)
	return nil
}

func WaitForJobCompletion(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job, namespace string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			jobStatus, err := client.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get job status: %w", err)
			}

			if jobStatus.Status.Succeeded > 0 {
				logger.KubeLogger.Info("Job %s completed successfully", job.Name)
				return nil
			}

			if jobStatus.Status.Failed > 0 {
				return fmt.Errorf("job %s failed", job.Name)
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func CopyTestResults(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job, namespace string) error {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
	})
	if err != nil {
		return fmt.Errorf("failed to list pods for job: %w", err)
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for job %s", job.Name)
	}
	pod := pods.Items[0]

	logger.KubeLogger.Info("Streaming test results from pod %s", pod.Name)

	req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: "testrunner",
		Follow:    false,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer stream.Close()

	buf := make([]byte, 2000)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			content := string(buf[:n])
			lines := strings.Split(content, "\n")
			logger.StreamPodLogs(lines)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read pod logs: %w", err)
		}
	}

	logger.KubeLogger.Info("Test results streamed successfully")
	return nil
}

func findRepositoryRoot(startPath string) (string, error) {
	repoRoot := startPath
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			return repoRoot, nil
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			return "", fmt.Errorf("could not find repository root")
		}
		repoRoot = parent
	}
}

func calculateWorkingDirectory(projectRoot, kindWorkspacePath string) (string, error) {
	if projectRoot == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}

		repoRoot, err := findRepositoryRoot(cwd)
		if err != nil {
			return "", err
		}

		relPath, err := filepath.Rel(repoRoot, cwd)
		if err != nil {
			return "", fmt.Errorf("failed to calculate relative path: %w", err)
		}

		return filepath.Join(kindWorkspacePath, relPath), nil
	}

	return filepath.Join(kindWorkspacePath, projectRoot), nil
}
