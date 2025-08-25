package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"testrunner/pkg/config"
	"testrunner/pkg/launcher"

	"github.com/spf13/cobra"
)

var (
	projectRoot       string
	image             string
	debug             bool
	kindWorkspacePath string
	targetPod         string
	targetNamespace   string
	testCommand       string
	mirrordProcess    string
	steal             bool
	keepNamespace     bool
	backoffLimit      int32
	activeDeadline    int64
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	var rootCmd = &cobra.Command{
		Use:   "ket",
		Short: "Kubernetes Embedded Testing - Run tests in Kubernetes with traffic interception",
		Long: `ket (Kubernetes Embedded Testing) is a tool for running integration tests 
in Kubernetes environments with traffic interception capabilities.

It supports the launch mode to deploy tests to Kubernetes and run them with traffic interception.`,
	}

	rootCmd.PersistentFlags().StringVarP(&projectRoot, "project-root", "r", ".", "Project root path")
	rootCmd.PersistentFlags().StringVarP(&image, "image", "i", "node:18-alpine", "Runner image")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&kindWorkspacePath, "kind-workspace-path", "w", "/workspace", "Kind workspace path")

	var launchCmd = &cobra.Command{
		Use:   "launch",
		Short: "Launch tests in Kubernetes with traffic interception",
		Long: `Launch tests in Kubernetes by:
1. Creating an isolated namespace
2. Deploying your source code as a ConfigMap
3. Running tests with mirrord traffic interception
4. Streaming results back to stdout
5. Cleaning up automatically`,
		RunE: func(cmd *cobra.Command, args []string) error {
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled")
			default:
			}

			cfg := config.Config{
				Mode:              "launch",
				ProjectRoot:       projectRoot,
				Image:             image,
				Debug:             debug,
				TargetPod:         targetPod,
				TargetNS:          targetNamespace,
				TestCommand:       testCommand,
				ProcessToTest:     mirrordProcess,
				Steal:             steal,
				KeepNamespace:     keepNamespace,
				BackoffLimit:      backoffLimit,
				ActiveDeadlineS:   activeDeadline,
				KindWorkspacePath: kindWorkspacePath,
			}
			cfg.SetDefaults()
			if err := launcher.Run(cfg); err != nil {
				return fmt.Errorf("launch failed: %w", err)
			}
			return nil
		},
	}

	launchCmd.Flags().StringVarP(&targetPod, "target-pod", "p", "", "Target pod to test against (required)")
	launchCmd.Flags().StringVarP(&targetNamespace, "target-namespace", "n", "default", "Target namespace")
	launchCmd.Flags().StringVarP(&testCommand, "test-command", "t", "", "Test command to execute (required)")
	launchCmd.Flags().StringVarP(&mirrordProcess, "mirrord-process", "m", "", "Process to run with mirrord (optional, enables traffic interception)")
	launchCmd.Flags().BoolVarP(&steal, "steal", "s", false, "Enable mirrord steal mode (requires --mirrord-process)")
	launchCmd.Flags().BoolVarP(&keepNamespace, "keep-namespace", "k", false, "Keep test namespace after run")
	launchCmd.Flags().Int32VarP(&backoffLimit, "backoff-limit", "b", 1, "Job backoff limit")
	launchCmd.Flags().Int64VarP(&activeDeadline, "active-deadline-seconds", "d", 1800, "Job deadline in seconds")

	launchCmd.MarkFlagRequired("target-pod")
	launchCmd.MarkFlagRequired("test-command")

	rootCmd.AddCommand(launchCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
