package kube

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/logger"
)

// JobManager manages the lifecycle of a test job including mirrord processes
type JobManager struct {
	mirrordProcess *exec.Cmd
	ctx            context.Context
	cancel         context.CancelFunc
	cfg            config.Config
	mirrordStdout  io.ReadCloser
	mirrordStderr  io.ReadCloser
}

// NewJobManager creates a new job manager
func NewJobManager(cfg config.Config) *JobManager {
	return &JobManager{
		cfg: cfg,
	}
}

// StartMirrord starts the mirrord process if configured
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
	args = append(args, "--", jm.cfg.ProcessToTest)

	jm.mirrordProcess = exec.CommandContext(jm.ctx, "mirrord", args...)

	var err error
	jm.mirrordStdout, err = jm.mirrordProcess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	jm.mirrordStderr, err = jm.mirrordProcess.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	logger.Info(logger.KUBE, "Starting mirrord with process: %s", jm.cfg.ProcessToTest)
	logger.Info(logger.KUBE, "Mirrord command: mirrord %v", args)

	if err := jm.mirrordProcess.Start(); err != nil {
		logger.Error(logger.KUBE, "Failed to start mirrord process: %v", err)
		return fmt.Errorf("failed to start mirrord: %w", err)
	}

	logger.Info(logger.KUBE, "Mirrord process started, PID: %d", jm.mirrordProcess.Process.Pid)
	logger.Info(logger.KUBE, "Waiting for mirrord to initialize...")

	time.Sleep(5 * time.Second)

	if jm.mirrordProcess.Process == nil || jm.mirrordProcess.Process.Pid == 0 {
		return fmt.Errorf("mirrord process failed to start")
	}

	logger.Info(logger.KUBE, "Mirrord started successfully with PID: %d", jm.mirrordProcess.Process.Pid)
	return nil
}

// StreamMirrordLogs streams mirrord stdout and stderr logs
func (jm *JobManager) StreamMirrordLogs() {
	if jm.mirrordProcess == nil || jm.mirrordStdout == nil || jm.mirrordStderr == nil {
		logger.Warn(logger.KUBE, "Cannot stream mirrord logs: process or pipes not available")
		return
	}

	logger.Info(logger.KUBE, "Starting to stream mirrord logs...")

	// Stream stdout
	go func() {
		logger.Info(logger.KUBE, "Starting mirrord stdout stream")
		buf := make([]byte, 1024)
		for {
			n, err := jm.mirrordStdout.Read(buf)
			if n > 0 {
				logger.Info(logger.KUBE, "Mirrord stdout: %s", string(buf[:n]))
				fmt.Printf("[MIRRORD] %s", string(buf[:n]))
			}
			if err != nil {
				if err != io.EOF {
					logger.Error(logger.KUBE, "Error reading mirrord stdout: %v", err)
				} else {
					logger.Info(logger.KUBE, "Mirrord stdout stream ended")
				}
				break
			}
		}
	}()

	// Stream stderr
	go func() {
		logger.Info(logger.KUBE, "Starting mirrord stderr stream")
		buf := make([]byte, 1024)
		for {
			n, err := jm.mirrordStderr.Read(buf)
			if n > 0 {
				logger.Info(logger.KUBE, "Mirrord stderr: %s", string(buf[:n]))
				fmt.Printf("[MIRRORD-ERROR] %s", string(buf[:n]))
			}
			if err != nil {
				if err != io.EOF {
					logger.Error(logger.KUBE, "Error reading mirrord stderr: %v", err)
				} else {
					logger.Info(logger.KUBE, "Mirrord stderr stream ended")
				}
				break
			}
		}
	}()
}

// StopMirrord stops the mirrord process
func (jm *JobManager) StopMirrord() error {
	if jm.mirrordProcess == nil || jm.mirrordProcess.Process == nil {
		return nil
	}

	logger.Info(logger.KUBE, "Stopping mirrord process (PID: %d)", jm.mirrordProcess.Process.Pid)

	if jm.cancel != nil {
		jm.cancel()
	}

	done := make(chan error, 1)
	go func() {
		done <- jm.mirrordProcess.Wait()
	}()

	select {
	case err := <-done:
		logger.Info(logger.KUBE, "Mirrord process stopped: %v", err)
	case <-time.After(5 * time.Second):
		logger.Warn(logger.KUBE, "Mirrord process did not stop gracefully, forcing kill")
		if err := jm.mirrordProcess.Process.Kill(); err != nil {
			logger.Error(logger.KUBE, "Failed to kill mirrord process: %v", err)
		}
	}

	return nil
}

// Cleanup performs all cleanup operations
func (jm *JobManager) Cleanup() {
	jm.StopMirrord()
}
