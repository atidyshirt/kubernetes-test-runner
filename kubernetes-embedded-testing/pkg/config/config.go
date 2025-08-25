package config

import (
	"fmt"

	"github.com/google/uuid"
)

type Config struct {
	Mode              string
	ProjectRoot       string
	Image             string
	Debug             bool
	TargetPod         string
	TargetNS          string
	TestCommand       string
	ProcessToTest     string
	Steal             bool
	KeepNamespace     bool
	BackoffLimit      int32
	ActiveDeadlineS   int64
	KindWorkspacePath string
}

func (cfg *Config) SetDefaults() {
	if cfg.TargetNS == "" {
		cfg.TargetNS = fmt.Sprintf("ket-%s", uuid.New().String()[:8])
	}
}
