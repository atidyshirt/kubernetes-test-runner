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

	// Add root flags
	addRootFlags(rootCmd)

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
		Long: "Launch tests in Kubernetes by:" +
			"1. Mounting your source code into the test runner pod" +
			"2. Running the specified test command" +
			"3. Streaming results back to stdout" +
			"4. Cleaning up test-runner resources automatically",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeLaunch(ctx)
		},
	}

	// Add launch-specific flags
	addLaunchFlags(launchCmd)

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
	cfg.Ctx = ctx
	if err := launcher.Run(*cfg); err != nil {
		// Check if this is a test execution error with an exit code
		if testErr, ok := err.(*launcher.TestExecutionError); ok {
			// Store the exit code for the main function to use
			testExitCode = testErr.ExitCode
			return fmt.Errorf("test execution failed with exit code %d: %s", testErr.ExitCode, testErr.Message)
		}
		return fmt.Errorf("launch failed: %w", err)
	}
	return nil
}
