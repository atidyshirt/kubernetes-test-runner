package kube

import (
	"context"
	"fmt"
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
		return nil // No mirrord process needed
	}

	jm.ctx, jm.cancel = context.WithCancel(context.Background())

	args := []string{"exec"}
	if jm.cfg.Steal {
		args = append(args, "--steal")
	}
	args = append(args, "--target-pod", jm.cfg.TargetPod)
	args = append(args, "--target-namespace", jm.cfg.TargetNS)
	args = append(args, "--", jm.cfg.ProcessToTest)

	jm.mirrordProcess = exec.CommandContext(jm.ctx, "mirrord", args...)
	jm.mirrordProcess.Stdout = nil
	jm.mirrordProcess.Stderr = nil

	logger.Info(logger.KUBE, "Starting mirrord with process: %s", jm.cfg.ProcessToTest)
	logger.Info(logger.KUBE, "Mirrord command: mirrord %v", args)

	if err := jm.mirrordProcess.Start(); err != nil {
		return fmt.Errorf("failed to start mirrord: %w", err)
	}

	// Give mirrord a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify the process is running
	if jm.mirrordProcess.Process == nil || jm.mirrordProcess.Process.Pid == 0 {
		return fmt.Errorf("mirrord process failed to start")
	}

	logger.Info(logger.KUBE, "Mirrord started successfully with PID: %d", jm.mirrordProcess.Process.Pid)
	return nil
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

	// Give the process a moment to shut down gracefully
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
