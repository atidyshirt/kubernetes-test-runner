package main

import (
	"fmt"
	"os"

	"testrunner/pkg/config"
	"testrunner/pkg/launcher"
	"testrunner/pkg/runner"
)

func main() {
	cfg := config.ParseFlags()

	switch cfg.Mode {
	case "launch":
		if err := launcher.Run(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "launch failed: %v\n", err)
			os.Exit(1)
		}
	case "run":
		if err := runner.Run(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
			os.Exit(2)
		}
	default:
		fmt.Fprintln(os.Stderr, "--mode must be 'launch' or 'run'")
		os.Exit(3)
	}
}
