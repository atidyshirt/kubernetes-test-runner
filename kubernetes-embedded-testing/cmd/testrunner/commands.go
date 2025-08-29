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

	addRootFlags(rootCmd)

	launchCmd := createLaunchCommand(ctx)
	rootCmd.AddCommand(launchCmd)

	manifestCmd := createManifestCommand(ctx)
	rootCmd.AddCommand(manifestCmd)

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
		// Disable help display on errors since test failures are expected
		SilenceUsage:  true,
		SilenceErrors: true,
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
		if testErr, ok := err.(*launcher.TestExecutionError); ok {
			testExitCode = testErr.ExitCode
			return fmt.Errorf("test execution failed with exit code %d: %s", testErr.ExitCode, testErr.Message)
		}
		return fmt.Errorf("launch failed: %w", err)
	}
	return nil
}

// createManifestCommand creates the manifest command
func createManifestCommand(ctx context.Context) *cobra.Command {
	manifestCmd := &cobra.Command{
		Use:   "manifest",
		Short: "Generate Kubernetes manifests for tests",
		Long: "Generate Kubernetes manifests that would be applied when running tests. " +
			"This command outputs the YAML manifests to stdout without applying them to the cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeManifest(ctx)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add manifest-specific flags (same as launch command)
	addLaunchFlags(manifestCmd)

	return manifestCmd
}

// executeManifest handles the manifest command execution
func executeManifest(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	default:
	}

	cfg := buildConfig()
	cfg.Ctx = ctx
	cfg.DryRun = true // Force dry run mode for manifest command

	if err := launcher.Run(*cfg); err != nil {
		return fmt.Errorf("manifest generation failed: %w", err)
	}
	return nil
}
