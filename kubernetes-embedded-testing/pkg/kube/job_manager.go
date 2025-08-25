package kube

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/logger"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobManager struct {
	mirrordProcess *exec.Cmd
	ctx            context.Context
	cancel         context.CancelFunc
	cfg            config.Config
	mirrordStdout  io.ReadCloser
	mirrordStderr  io.ReadCloser
	logStreamer    *logger.LogStreamer
}

func NewJobManager(cfg config.Config) *JobManager {
	return &JobManager{
		cfg:         cfg,
		logStreamer: logger.NewLogStreamer(),
	}
}

func (jm *JobManager) StartMirrord() error {
	if jm.cfg.ProcessToTest == "" {
		return nil
	}

	jm.ctx, jm.cancel = context.WithCancel(context.Background())

	args := []string{"exec"}
	if jm.cfg.Steal {
		args = append(args, "--steal")
	}
	args = append(args, "--target", jm.cfg.TargetPod)
	args = append(args, "--target-namespace", jm.cfg.TargetNS)

	cmdParts := strings.Fields(jm.cfg.ProcessToTest)
	logger.KubeLogger.Info("Debug: ProcessToTest='%s', cmdParts=%v", jm.cfg.ProcessToTest, cmdParts)
	args = append(args, "--", cmdParts[0])
	if len(cmdParts) > 1 {
		args = append(args, cmdParts[1:]...)
	}
	logger.KubeLogger.Info("Debug: Final args=%v", args)

	jm.mirrordProcess = exec.CommandContext(jm.ctx, "mirrord", args...)

	if jm.cfg.ProjectRoot != "" {
		jm.mirrordProcess.Dir = jm.cfg.ProjectRoot
	}

	var err error
	jm.mirrordStdout, err = jm.mirrordProcess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	jm.mirrordStderr, err = jm.mirrordProcess.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	logger.KubeLogger.Info("Starting mirrord with process: %s", jm.cfg.ProcessToTest)
	logger.KubeLogger.Info("Mirrord command: mirrord %v", args)

	if err := jm.mirrordProcess.Start(); err != nil {
		logger.KubeLogger.Error("Failed to start mirrord process: %v", err)
		return fmt.Errorf("failed to start mirrord: %w", err)
	}

	logger.KubeLogger.Info("Mirrord process started, PID: %d", jm.mirrordProcess.Process.Pid)
	logger.KubeLogger.Info("Waiting for mirrord to initialize...")

	if err := jm.waitForMirrordReady(); err != nil {
		return fmt.Errorf("mirrord failed to become ready: %w", err)
	}

	logger.KubeLogger.Info("Mirrord started successfully with PID: %d", jm.mirrordProcess.Process.Pid)
	return nil
}

func (jm *JobManager) waitForMirrordReady() error {
	const (
		maxWaitTime   = 60 * time.Second
		checkInterval = 2 * time.Second
	)

	startTime := time.Now()

	for {
		if time.Since(startTime) > maxWaitTime {
			return fmt.Errorf("mirrord failed to become ready within %v", maxWaitTime)
		}

		if jm.mirrordProcess.Process == nil || jm.mirrordProcess.Process.Pid == 0 {
			return fmt.Errorf("mirrord process died during initialization")
		}

		if err := jm.mirrordProcess.Process.Signal(syscall.Signal(0)); err != nil {
			return fmt.Errorf("mirrord process is not responding: %w", err)
		}

		if ready, err := jm.checkMirrordAgentHealth(); err != nil {
			logger.KubeLogger.Warn("Failed to check mirrord agent health: %v", err)
		} else if ready {
			logger.KubeLogger.Info("Mirrord agent pods are healthy and ready")
			return nil
		}

		logger.KubeLogger.Info("Mirrord not ready yet, waiting %v before next check...", checkInterval)
		time.Sleep(checkInterval)
	}
}

func (jm *JobManager) checkMirrordAgentHealth() (bool, error) {
	client, err := NewClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	pods, err := client.CoreV1().Pods(jm.cfg.TargetNS).List(context.Background(), metav1.ListOptions{
		LabelSelector: "app=mirrord",
	})
	if err != nil {
		allPods, err := client.CoreV1().Pods(jm.cfg.TargetNS).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to list pods: %w", err)
		}

		for _, pod := range allPods.Items {
			if strings.Contains(strings.ToLower(pod.Name), "mirrord") {
				pods = &corev1.PodList{Items: []corev1.Pod{pod}}
				break
			}
		}
	}

	if err != nil {
		return false, fmt.Errorf("failed to find mirrord agent pods: %w", err)
	}

	if len(pods.Items) == 0 {
		logger.KubeLogger.Info("No mirrord agent pods found yet")
		return false, nil
	}

	for _, pod := range pods.Items {
		logger.KubeLogger.Info("Found mirrord agent pod: %s (status: %s)", pod.Name, pod.Status.Phase)

		if pod.Status.Phase == corev1.PodRunning {
			allContainersReady := true
			for _, container := range pod.Status.ContainerStatuses {
				if !container.Ready {
					allContainersReady = false
					logger.KubeLogger.Info("Container %s in pod %s is not ready", container.Name, pod.Name)
					break
				}
			}

			if allContainersReady {
				logger.KubeLogger.Info("Mirrord agent pod %s is running and all containers are ready", pod.Name)

				ready, err := jm.checkMirrordAgentLogs(pod.Name, jm.cfg.TargetNS)
				if err != nil {
					logger.KubeLogger.Warn("Failed to check mirrord agent logs: %v", err)
					continue
				}

				if ready {
					logger.KubeLogger.Info("Mirrord agent %s logs show it's ready", pod.Name)
					return true, nil
				}
			}
		} else if pod.Status.Phase == corev1.PodPending {
			logger.KubeLogger.Info("Mirrord agent pod %s is still pending", pod.Name)
		} else if pod.Status.Phase == corev1.PodFailed {
			logger.KubeLogger.Error("Mirrord agent pod %s has failed", pod.Name)
			return false, fmt.Errorf("mirrord agent pod %s failed", pod.Name)
		}
	}

	return false, nil
}

func (jm *JobManager) checkMirrordAgentLogs(podName, namespace string) (bool, error) {
	client, err := NewClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "mirrord-agent",
		TailLines: &[]int64{50}[0],
	})

	stream, err := req.Stream(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer stream.Close()

	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read pod logs: %w", err)
	}

	if n > 0 {
		logs := string(buf[:n])
		if strings.Contains(logs, "agent ready") {
			return true, nil
		}
	}

	return false, nil
}

func (jm *JobManager) StreamMirrordLogs() {
	if jm.mirrordProcess == nil || jm.mirrordStdout == nil || jm.mirrordStderr == nil {
		logger.KubeLogger.Warn("Cannot stream mirrord logs: process or pipes not available")
		return
	}

	logger.KubeLogger.Info("Starting to stream mirrord logs...")

	mirrordStdoutFile, err := os.CreateTemp("", "mirrord-stdout-*.log")
	if err != nil {
		logger.KubeLogger.Error("Failed to create mirrord stdout log file: %v", err)
		return
	}

	mirrordStderrFile, err := os.CreateTemp("", "mirrord-stderr-*.log")
	if err != nil {
		logger.KubeLogger.Error("Failed to create mirrord stderr log file: %v", err)
		mirrordStdoutFile.Close()
		return
	}

	jm.logStreamer.StartMirrordLogStreaming(mirrordStdoutFile, mirrordStderrFile)

	logger.KubeLogger.Info("Mirrord logs will be written to: %s (stdout), %s (stderr)", mirrordStdoutFile.Name(), mirrordStderrFile.Name())

	go func() {
		defer mirrordStdoutFile.Close()
		logger.KubeLogger.Info("Starting mirrord stdout stream to file")
		buf := make([]byte, 1024)
		for {
			n, err := jm.mirrordStdout.Read(buf)
			if n > 0 {
				if _, writeErr := mirrordStdoutFile.Write(buf[:n]); writeErr != nil {
					logger.KubeLogger.Error("Failed to write to mirrord stdout log file: %v", writeErr)
				}
			}
			if err != nil {
				if err != io.EOF {
					logger.KubeLogger.Error("Error reading mirrord stdout: %v", err)
				} else {
					logger.KubeLogger.Info("Mirrord stdout stream ended")
				}
				break
			}
		}
	}()

	go func() {
		defer mirrordStderrFile.Close()
		logger.KubeLogger.Info("Starting mirrord stderr stream to file")
		buf := make([]byte, 1024)
		for {
			n, err := jm.mirrordStderr.Read(buf)
			if n > 0 {
				if _, writeErr := mirrordStderrFile.Write(buf[:n]); writeErr != nil {
					logger.KubeLogger.Error("Failed to write to mirrord stderr log file: %v", writeErr)
				}
			}
			if err != nil {
				if err != io.EOF {
					logger.KubeLogger.Error("Error reading mirrord stderr: %v", err)
				} else {
					logger.KubeLogger.Info("Mirrord stderr stream ended")
				}
				break
			}
		}
	}()
}

func (jm *JobManager) GetMirrordLogFiles() (stdout, stderr string) {
	return jm.logStreamer.GetLogFilePaths()
}

func (jm *JobManager) GetMirrordLogContents() (stdout, stderr string, err error) {
	return jm.logStreamer.GetLogContents()
}

func (jm *JobManager) StopMirrord() error {
	if jm.mirrordProcess == nil || jm.mirrordProcess.Process == nil {
		return nil
	}

	logger.KubeLogger.Info("Stopping mirrord process (PID: %d)", jm.mirrordProcess.Process.Pid)

	if jm.cancel != nil {
		jm.cancel()
	}

	done := make(chan error, 1)
	go func() {
		done <- jm.mirrordProcess.Wait()
	}()

	select {
	case err := <-done:
		logger.KubeLogger.Info("Mirrord process stopped: %v", err)
	case <-time.After(5 * time.Second):
		logger.KubeLogger.Warn("Mirrord process did not stop gracefully, forcing kill")
		if err := jm.mirrordProcess.Process.Kill(); err != nil {
			logger.KubeLogger.Error("Failed to kill mirrord process: %v", err)
		}
	}

	return nil
}

func (jm *JobManager) Cleanup() {
	jm.StopMirrord()
}
