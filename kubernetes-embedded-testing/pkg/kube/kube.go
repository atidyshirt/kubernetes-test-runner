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

// NewClient creates a new Kubernetes client
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

// CreateNamespace creates a new namespace in Kubernetes
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

// DeleteNamespace deletes a namespace from Kubernetes
func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset, name string) error {
	return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

// CreateJob creates a new job in Kubernetes
func CreateJob(ctx context.Context, client *kubernetes.Clientset, cfg config.Config, namespace string) (*batchv1.Job, error) {
	hostProjectRoot := filepath.Join(cfg.KindWorkspacePath, cfg.ProjectRoot)
	if cfg.ProjectRoot == "." {
		hostProjectRoot = cfg.KindWorkspacePath
	}

	workingDir, err := calculateWorkingDirectory(cfg.ProjectRoot, cfg.KindWorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate working directory: %w", err)
	}

	logger.Info(logger.KUBE, "Job configuration: hostPath=%s, workingDir=%s, kindWorkspacePath=%s", hostProjectRoot, workingDir, cfg.KindWorkspacePath)

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
							Name:    "testrunner",
							Image:   cfg.Image,
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								fmt.Sprintf(`
set -e
echo "Starting test execution%s"
echo "Target pod: %s in namespace: %s"

# Run the test command
echo "Running test command: %s"
eval %s
TEST_EXIT_CODE=$?
echo "Test command completed with exit code: $TEST_EXIT_CODE"

exit $TEST_EXIT_CODE
`,
									func() string {
										if cfg.ProcessToTest != "" {
											return " with mirrord"
										}
										return ""
									}(),
									cfg.TargetPod, cfg.TargetNS,
									cfg.TestCommand, cfg.TestCommand),
							},
							Env: []corev1.EnvVar{
								{Name: "TARGET_NAMESPACE", Value: cfg.TargetNS},
								{Name: "TARGET_POD", Value: cfg.TargetPod},
								{Name: "PROCESS_TO_TEST", Value: cfg.ProcessToTest},
								{Name: "TEST_COMMAND", Value: cfg.TestCommand},
							},
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

// StreamJobLogs streams logs from a job's pods
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
	logger.Info(logger.KUBE, "Finished streaming logs from %s", pod.Name)
	return nil
}

// WaitForJobCompletion waits for a job to complete and returns the final status
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
				logger.Info(logger.KUBE, "Job %s completed successfully", job.Name)
				return nil
			}

			if jobStatus.Status.Failed > 0 {
				return fmt.Errorf("job %s failed", job.Name)
			}

			time.Sleep(5 * time.Second)
		}
	}
}

// CopyTestResults streams test results and logs to stdout
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

	logger.Info(logger.KUBE, "Streaming test results from pod %s", pod.Name)

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
			for _, line := range lines {
				if line != "" {
					if strings.Contains(line, "[MIRRORD]") {
						fmt.Printf("[MIRRORD] %s\n", line)
					} else if strings.Contains(line, "[NPM]") {
						fmt.Printf("[NPM] %s\n", line)
					} else if strings.Contains(line, "[MOCHA]") {
						fmt.Printf("[MOCHA] %s\n", line)
					} else if strings.Contains(line, "[POD]") {
						fmt.Printf("[POD] %s\n", line)
					} else if strings.Contains(line, "[TESTRUNNER]") {
						fmt.Printf("[TESTRUNNER] %s\n", line)
					} else {
						fmt.Printf("[POD] %s\n", line)
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read pod logs: %w", err)
		}
	}

	logger.Info(logger.KUBE, "Test results streamed successfully")
	return nil
}

// findRepositoryRoot finds the repository root by looking for a .git directory
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

// calculateWorkingDirectory calculates the working directory for the container
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
