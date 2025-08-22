package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type Config struct {
	Mode            string
	ProjectRoot     string
	TestCommand     string
	Namespace       string
	Image           string
	TargetPod       string
	TargetNS        string
	MirrordURL      string
	ManifestsDir    string
	KeepNamespace   bool
	BackoffLimit    int32
	ActiveDeadlineS int64
	ProcessToTest   string
	TestTimeout     int64
	Debug           bool
}

func ParseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.Mode, "mode", "launch", "Mode: launch | run")
	flag.StringVar(&cfg.ProjectRoot, "project-root", ".", "Project root path")
	flag.StringVar(&cfg.TestCommand, "test-command", "", "Test command to execute (e.g., 'mocha **/*.spec.js')")
	flag.StringVar(&cfg.Image, "image", "node:18-alpine", "Runner image")
	flag.StringVar(&cfg.TargetPod, "target-pod", "", "Target pod to test against")
	flag.StringVar(&cfg.TargetNS, "target-namespace", "default", "Target namespace")
	flag.StringVar(&cfg.MirrordURL, "mirrord-url", "https://github.com/metalbear-co/mirrord/releases/latest/download/mirrord", "Mirrord binary download URL")
	flag.StringVar(&cfg.ManifestsDir, "manifests-dir", "", "Directory with manifests to apply")
	flag.BoolVar(&cfg.KeepNamespace, "keep-namespace", false, "Keep test namespace after run")
	flag.Int64Var(&cfg.ActiveDeadlineS, "active-deadline-seconds", 1800, "Job deadline in seconds")
	flag.StringVar(&cfg.ProcessToTest, "proc", "", "Process to test against (e.g., 'npm run start')")
	flag.Int64Var(&cfg.TestTimeout, "test-timeout", 300, "Test timeout in seconds")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")

	var backoff int
	flag.IntVar(&backoff, "backoff-limit", 1, "Job backoff limit")

	flag.Parse()

	// Validate required fields
	if cfg.Mode == "launch" {
		if cfg.TargetPod == "" {
			fmt.Fprintln(os.Stderr, "Error: --target-pod is required for launch mode")
			os.Exit(1)
		}
		if cfg.TestCommand == "" {
			fmt.Fprintln(os.Stderr, "Error: --test-command is required for launch mode")
			os.Exit(1)
		}
		if cfg.ProcessToTest == "" {
			fmt.Fprintln(os.Stderr, "Error: --proc is required for launch mode")
			os.Exit(1)
		}
	}

	// Set defaults
	if cfg.Namespace == "" {
		// Generate a unique UUID namespace for test isolation
		cfg.Namespace = fmt.Sprintf("testrunner-%s", uuid.New().String()[:8])
	}

	cfg.BackoffLimit = int32(backoff)
	return cfg
}
