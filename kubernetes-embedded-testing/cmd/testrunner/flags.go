package main

import (
	"testrunner/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to config file (YAML/JSON)")
	cmd.PersistentFlags().StringVarP(&projectRoot, "project-root", "r", ".", "Project root path")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	cmd.PersistentFlags().StringVarP(&workspacePath, "cluster-workspace-path", "w", "/workspace",
		"Absolute path where the local project directory is mounted inside the test runner pod. "+
			"Defaults to '/workspace', matching Kind/K3D volume mounts (e.g., '$(pwd):/workspace'). "+
			"Used for syncing source code between local and cluster environments.")
}

func addLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&image, "image", "i", "ket-test-runner:latest",
		"Container image used for the test runner pod."+
			"Must include dependencies for the test command (e.g., mocha or other test runners).")
	cmd.Flags().StringVarP(&testCommand, "test-command", "t", "",
		"Command to execute inside the test runner pod (e.g., 'mocha **/*.spec.ts').")
	cmd.Flags().BoolVarP(&keepNamespace, "keep-namespace", "k", false,
		"If set, the test namespace will not be deleted after the run for debugging purposes.")
	cmd.Flags().Int32VarP(&backoffLimit, "backoff-limit", "b", 1,
		"Maximum number of retry attempts for a failed Kubernetes job.")
	cmd.Flags().Int64VarP(&activeDeadline, "active-deadline-seconds", "d", 1800,
		"Maximum duration in seconds the job is allowed to run before termination.")
}

func setupViper() *viper.Viper {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("ket-config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.ket")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only log error if it's not a "file not found" error
			// This allows the tool to work without a config file
		}
	}

	return v
}

func buildConfig() *config.Config {
	v := setupViper()

	// Set defaults
	v.SetDefault("mode", "launch")
	v.SetDefault("image", "ket-test-runner:latest")
	v.SetDefault("projectRoot", ".")
	v.SetDefault("clusterWorkspacePath", "/workspace")
	v.SetDefault("backoffLimit", int32(1))
	v.SetDefault("activeDeadlineS", int64(1800))
	v.SetDefault("debug", false)

	// Override with command line flags
	v.Set("projectRoot", projectRoot)
	v.Set("clusterWorkspacePath", workspacePath)
	v.Set("debug", debug)
	v.Set("image", image)
	v.Set("testCommand", testCommand)
	v.Set("keepNamespace", keepNamespace)
	v.Set("backoffLimit", backoffLimit)
	v.Set("activeDeadlineS", activeDeadline)

	cfg, err := config.LoadFromViper(v)
	if err != nil {
		// Fallback to basic config if Viper fails
		cfg = &config.Config{
			Mode:            "launch",
			Image:           image,
			ProjectRoot:     projectRoot,
			WorkspacePath:   workspacePath,
			BackoffLimit:    backoffLimit,
			ActiveDeadlineS: activeDeadline,
			Debug:           debug,
			TestCommand:     testCommand,
			KeepTestRunner:  keepNamespace,
		}
	}

	return cfg
}
