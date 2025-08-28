package main

import (
	"testrunner/pkg/config"

	"github.com/spf13/cobra"
)

// addPersistentFlags adds global flags to the root command
func addPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to config file (YAML/JSON)")
	cmd.PersistentFlags().StringVarP(&projectRoot, "project-root", "r", ".", "Project root path")
	cmd.PersistentFlags().StringVarP(&image, "image", "i", "ket-test-runner:latest", "Runner image")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	cmd.PersistentFlags().StringVarP(&kindWorkspacePath, "kind-workspace-path", "w", "/workspace", "Kind workspace path")
}

// buildConfig builds the configuration from flags and config file
func buildConfig() config.Config {
	var cfg config.Config

	if configFile != "" {
		if fileConfig, err := config.LoadFromFile(configFile); err == nil {
			cfg = *fileConfig
		}
	}

	cfg.Mode = "launch"

	// Always set defaults for required fields
	cfg.Image = image
	cfg.ProjectRoot = projectRoot
	cfg.KindWorkspacePath = kindWorkspacePath
	cfg.BackoffLimit = backoffLimit
	cfg.ActiveDeadlineS = activeDeadline

	if debug {
		cfg.Debug = debug
	}
	if testCommand != "" {
		cfg.TestCommand = testCommand
	}
	if keepNamespace {
		cfg.KeepNamespace = keepNamespace
	}

	return cfg
}
