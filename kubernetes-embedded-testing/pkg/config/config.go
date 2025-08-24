package config

import (
	"fmt"
	"github.com/google/uuid"
)

// Config holds the configuration for the ket application
type Config struct {
	Mode              string
	ProjectRoot       string
	TestCommand       string
	Namespace         string
	Image             string
	TargetPod         string
	TargetNS          string
	KeepNamespace     bool
	BackoffLimit      int32
	ActiveDeadlineS   int64
	ProcessToTest     string
	Debug             bool
	KindWorkspacePath string
}

// SetDefaults sets default values for the configuration
func (cfg *Config) SetDefaults() {
	if cfg.Namespace == "" {
		cfg.Namespace = fmt.Sprintf("ket-%s", uuid.New().String()[:8])
	}
}
