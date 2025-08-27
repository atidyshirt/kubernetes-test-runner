package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode              string `yaml:"mode" json:"mode"`
	ProjectRoot       string `yaml:"projectRoot" json:"projectRoot"`
	Image             string `yaml:"image" json:"image"`
	Debug             bool   `yaml:"debug" json:"debug"`
	TestCommand       string `yaml:"testCommand" json:"testCommand"`
	KeepNamespace     bool   `yaml:"keepNamespace" json:"keepNamespace"`
	BackoffLimit      int32  `yaml:"backoffLimit" json:"backoffLimit"`
	ActiveDeadlineS   int64  `yaml:"activeDeadlineS" json:"activeDeadlineS"`
	KindWorkspacePath string `yaml:"kindWorkspacePath" json:"kindWorkspacePath"`
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
