package main

import (
	"context"
	"fmt"

	"testrunner/pkg/launcher"

	"github.com/spf13/cobra"
)

// createRootCommand creates the main root command
func createRootCommand(ctx context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ket",
		Short: "Kubernetes Embedded Testing - Deploy and run tests in Kubernetes",
		Long: `ket (Kubernetes Embedded Testing) is a tool for deploying and running tests 
in Kubernetes environments. It creates an isolated namespace, deploys your source code, 
and runs the specified test command.`,
	}

	// Add persistent flags
	addPersistentFlags(rootCmd)

	// Add launch command
	launchCmd := createLaunchCommand(ctx)
	rootCmd.AddCommand(launchCmd)

	return rootCmd
}

// createLaunchCommand creates the launch command
func createLaunchCommand(ctx context.Context) *cobra.Command {
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch tests in Kubernetes",
		Long: `Launch tests in Kubernetes by:
1. Creating an isolated namespace
2. Deploying your source code
3. Running the specified test command
4. Streaming results back to stdout
5. Cleaning up automatically`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeLaunch(ctx)
		},
	}

	// Add launch-specific flags
	launchCmd.Flags().StringVarP(&testCommand, "test-command", "t", "", "Test command to execute")
	launchCmd.Flags().BoolVarP(&keepNamespace, "keep-namespace", "k", false, "Keep test namespace after run")
	launchCmd.Flags().Int32VarP(&backoffLimit, "backoff-limit", "b", 1, "Job backoff limit")
	launchCmd.Flags().Int64VarP(&activeDeadline, "active-deadline-seconds", "d", 1800, "Job deadline in seconds")

	return launchCmd
}

// executeLaunch handles the launch command execution
func executeLaunch(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	default:
	}

	cfg := buildConfig()
	if err := launcher.Run(cfg); err != nil {
		return fmt.Errorf("launch failed: %w", err)
	}
	return nil
}
