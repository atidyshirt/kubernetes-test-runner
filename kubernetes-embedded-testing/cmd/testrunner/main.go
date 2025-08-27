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
	configFile        string
	projectRoot       string
	image             string
	debug             bool
	kindWorkspacePath string
	testCommand       string
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
		Short: "Kubernetes Embedded Testing - Deploy and run tests in Kubernetes",
		Long: `ket (Kubernetes Embedded Testing) is a tool for deploying and running tests 
in Kubernetes environments. It creates an isolated namespace, deploys your source code, 
and runs the specified test command.`,
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to config file (YAML/JSON)")
	rootCmd.PersistentFlags().StringVarP(&projectRoot, "project-root", "r", ".", "Project root path")
	rootCmd.PersistentFlags().StringVarP(&image, "image", "i", "ket-test-runner:latest", "Runner image")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&kindWorkspacePath, "kind-workspace-path", "w", "/workspace", "Kind workspace path")

	var launchCmd = &cobra.Command{
		Use:   "launch",
		Short: "Launch tests in Kubernetes",
		Long: `Launch tests in Kubernetes by:
1. Creating an isolated namespace
2. Deploying your source code
3. Running the specified test command
4. Streaming results back to stdout
5. Cleaning up automatically`,
		RunE: func(cmd *cobra.Command, args []string) error {
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled")
			default:
			}

			var cfg config.Config

			if configFile != "" {
				fileConfig, err := config.LoadFromFile(configFile)
				if err != nil {
					return fmt.Errorf("failed to load config file: %w", err)
				}
				cfg = *fileConfig
			}

			cfg.Mode = "launch"

			if projectRoot != "." {
				cfg.ProjectRoot = projectRoot
			}
			if image != "ket-test-runner:latest" {
				cfg.Image = image
			}
			if debug {
				cfg.Debug = debug
			}
			if testCommand != "" {
				cfg.TestCommand = testCommand
			}
			if keepNamespace {
				cfg.KeepNamespace = keepNamespace
			}
			if backoffLimit != 1 {
				cfg.BackoffLimit = backoffLimit
			}
			if activeDeadline != 1800 {
				cfg.ActiveDeadlineS = activeDeadline
			}
			if kindWorkspacePath != "/workspace" {
				cfg.KindWorkspacePath = kindWorkspacePath
			}

			if cfg.TestCommand == "" {
				return fmt.Errorf("test command is required. Either provide --test-command flag or specify testCommand in config file")
			}

			if err := launcher.Run(cfg); err != nil {
				return fmt.Errorf("launch failed: %w", err)
			}
			return nil
		},
	}

	launchCmd.Flags().StringVarP(&testCommand, "test-command", "t", "", "Test command to execute")
	launchCmd.Flags().BoolVarP(&keepNamespace, "keep-namespace", "k", false, "Keep test namespace after run")
	launchCmd.Flags().Int32VarP(&backoffLimit, "backoff-limit", "b", 1, "Job backoff limit")
	launchCmd.Flags().Int64VarP(&activeDeadline, "active-deadline-seconds", "d", 1800, "Job deadline in seconds")

	rootCmd.AddCommand(launchCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
