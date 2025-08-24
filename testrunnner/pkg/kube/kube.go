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
	// Try in-cluster config first
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// fallback to local kubeconfig
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

// CreateConfigMapFromDirectory creates a ConfigMap from a local directory
func CreateConfigMapFromDirectory(ctx context.Context, client *kubernetes.Clientset, namespace, name, dirPath string) error {
	data := make(map[string]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Skip node_modules and other common directories
		relPath, _ := filepath.Rel(dirPath, path)
		if strings.Contains(relPath, "node_modules") ||
			strings.Contains(relPath, ".git") ||
			strings.Contains(relPath, "dist") ||
			strings.Contains(relPath, "build") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			logger.Warn(logger.KUBE, "Could not read file %s: %v", path, err)
			return nil
		}

		// Use relative path as key, but encode it to handle special characters
		key := strings.ReplaceAll(relPath, "/", "_")
		data[key] = string(content)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	logger.Info(logger.KUBE, "Created ConfigMap %s with %d files", name, len(data))
	return nil
}

func CreateJob(ctx context.Context, client *kubernetes.Clientset, cfg config.Config) (*batchv1.Job, error) {
	// Create ConfigMap from local directory
	projectRoot := cfg.ProjectRoot

	// Get the actual directory name for naming purposes
	var projectName string
	if projectRoot == "." {
		// If project root is ".", get the current working directory name
		if cwd, err := os.Getwd(); err == nil {
			projectName = filepath.Base(cwd)
		} else {
			projectName = "project"
		}
	} else {
		// Use the last part of the path as the project name
		projectName = filepath.Base(projectRoot)
	}

	configMapName := fmt.Sprintf("ket-source-%s", projectName)
	if err := CreateConfigMapFromDirectory(ctx, client, cfg.Namespace, configMapName, cfg.ProjectRoot); err != nil {
		return nil, fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ket-%s", projectName),
			Namespace: cfg.Namespace,
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
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "reports",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:    "extract-source",
							Image:   "node:18-alpine",
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								"cd /workspace && for key in /source/*; do filename=$(basename $key); target_path=$(echo $filename | sed 's/_/\\//g'); mkdir -p $(dirname $target_path) && cp $key $target_path; done && npm ci",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source-code",
									MountPath: "/source",
								},
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
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
cd /workspace
echo "Starting test execution with mirrord --steal"
echo "Target pod: %s in namespace: %s"
echo "Using mirrord --steal to intercept traffic from running pod"

# Check if mirrord is available and working
if [ -x "/tools/mirrord" ] && /tools/mirrord --version >/dev/null 2>&1; then
    echo "Mirrord is available, starting in steal mode"
    # Start mirrord in steal mode to intercept traffic from the target pod
    /tools/mirrord exec --steal --target-pod %s --target-namespace %s -- %s &
    MIRRORD_PID=$!
    echo "Mirrord started with --steal, PID: $MIRRORD_PID"
    
    # Wait a moment for mirrord to establish connection
    sleep 3
else
    echo "Warning: Mirrord is not available or not working"
    echo "Tests will run without traffic interception"
    echo "This means tests will run against the local Express server, not the target pod"
fi

# Run the test command
echo "Running test command: %s"
eval %s
TEST_EXIT_CODE=$?
echo "Test command completed with exit code: $TEST_EXIT_CODE"

# Clean up mirrord if it was started
if [ ! -z "$MIRRORD_PID" ]; then
    echo "Cleaning up mirrord process (PID: $MIRRORD_PID)"
    kill $MIRRORD_PID 2>/dev/null || true
    wait $MIRRORD_PID 2>/dev/null || true
fi

exit $TEST_EXIT_CODE
`, cfg.TargetPod, cfg.TargetNS, cfg.TargetPod, cfg.TargetNS, cfg.ProcessToTest, cfg.TestCommand, cfg.TestCommand),
							},
							Env: []corev1.EnvVar{
								{Name: "TARGET_NAMESPACE", Value: cfg.TargetNS},
								{Name: "TARGET_POD", Value: cfg.TargetPod},
								{Name: "PROCESS_TO_TEST", Value: cfg.ProcessToTest},
								{Name: "TEST_COMMAND", Value: cfg.TestCommand},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
								{
									Name:      "reports",
									MountPath: "/reports",
								},
							},
							WorkingDir: "/workspace",
						},
					},
				},
			},
		},
	}

	return client.BatchV1().Jobs(cfg.Namespace).Create(ctx, job, metav1.CreateOptions{})
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
	// Get the pod for this job
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

	// Stream logs from the testrunner container
	req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: "testrunner",
		Follow:    false, // Get all logs since we're copying after completion
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer stream.Close()

	// Copy logs to stdout with prefixes for different log sources
	buf := make([]byte, 2000)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			output := string(buf[:n])
			// Add prefixes to different types of logs for better identification
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if line = strings.TrimSpace(line); line != "" {
					// Identify and prefix different log sources
					if strings.Contains(line, "mocha") || strings.Contains(line, "Express Server") || strings.Contains(line, "should") {
						fmt.Printf("[MOCHA] %s\n", line)
					} else if strings.Contains(line, "npm") || strings.Contains(line, "Running test command") {
						fmt.Printf("[NPM] %s\n", line)
					} else if strings.Contains(line, "mirrord") || strings.Contains(line, "Target pod") {
						fmt.Printf("[MIRRORD] %s\n", line)
					} else if strings.Contains(line, "Starting test execution") || strings.Contains(line, "Test command completed") {
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
			return fmt.Errorf("error reading logs: %w", err)
		}
	}

	logger.Info(logger.KUBE, "Test results streamed successfully")
	return nil
}
