package main

import (
	"fmt"
	"os"

	"testrunner/pkg/config"
	"testrunner/pkg/launcher"
	"testrunner/pkg/runner"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	projectRoot string
	image       string
	debug       bool

	// Launch-specific flags
	targetPod       string
	targetNamespace string
	testCommand     string
	processToTest   string
	keepNamespace   bool
	backoffLimit    int32
	activeDeadline  int64
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ket",
		Short: "Kubernetes Embedded Testing - Run tests in Kubernetes with traffic interception",
		Long: `ket (Kubernetes Embedded Testing) is a tool for running integration tests 
in Kubernetes environments with traffic interception capabilities.

It supports two main modes:
- launch: Deploy tests to Kubernetes and run them with traffic interception
- run: Execute tests within a Kubernetes pod (internal use)`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&projectRoot, "project-root", "r", ".", "Project root path")
	rootCmd.PersistentFlags().StringVarP(&image, "image", "i", "node:18-alpine", "Runner image")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")

	// Launch command
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
			cfg := config.Config{
				Mode:            "launch",
				ProjectRoot:     projectRoot,
				Image:           image,
				Debug:           debug,
				TargetPod:       targetPod,
				TargetNS:        targetNamespace,
				TestCommand:     testCommand,
				ProcessToTest:   processToTest,
				KeepNamespace:   keepNamespace,
				BackoffLimit:    backoffLimit,
				ActiveDeadlineS: activeDeadline,
			}

			// Set defaults (like namespace generation)
			cfg.SetDefaults()

			if err := launcher.Run(cfg); err != nil {
				return fmt.Errorf("launch failed: %w", err)
			}
			return nil
		},
	}

	// Launch command flags
	launchCmd.Flags().StringVarP(&targetPod, "target-pod", "p", "", "Target pod to test against (required)")
	launchCmd.Flags().StringVarP(&targetNamespace, "target-namespace", "n", "default", "Target namespace")
	launchCmd.Flags().StringVarP(&testCommand, "test-command", "t", "", "Test command to execute (required)")
	launchCmd.Flags().StringVarP(&processToTest, "proc", "c", "", "Process to test against (required)")
	launchCmd.Flags().BoolVarP(&keepNamespace, "keep-namespace", "k", false, "Keep test namespace after run")
	launchCmd.Flags().Int32VarP(&backoffLimit, "backoff-limit", "b", 1, "Job backoff limit")
	launchCmd.Flags().Int64VarP(&activeDeadline, "active-deadline-seconds", "d", 1800, "Job deadline in seconds")

	// Mark required flags
	launchCmd.MarkFlagRequired("target-pod")
	launchCmd.MarkFlagRequired("test-command")
	launchCmd.MarkFlagRequired("proc")

	// Run command (internal use)
	var runCmd = &cobra.Command{
		Use:    "run",
		Short:  "Execute tests within Kubernetes pod (internal use)",
		Hidden: true, // Hide this from help as it's for internal use
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Config{
				Mode:        "run",
				ProjectRoot: projectRoot,
				Image:       image,
				Debug:       debug,
			}

			if err := runner.Run(cfg); err != nil {
				return fmt.Errorf("run failed: %w", err)
			}
			return nil
		},
	}

	// Add subcommands
	rootCmd.AddCommand(launchCmd, runCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
