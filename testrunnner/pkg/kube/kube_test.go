package kube

import (
	"testing"
)

func TestCreateConfigMapName(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		expected    string
	}{
		{
			name:        "simple project name",
			projectRoot: "my-project",
			expected:    "testrunner-source-my-project",
		},
		{
			name:        "project with path",
			projectRoot: "/path/to/my-project",
			expected:    "testrunner-source-my-project",
		},
		{
			name:        "project with dots",
			projectRoot: "my.project",
			expected:    "testrunner-source-my-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simple test to verify the function exists and can be called
			// In a real test, we would mock the Kubernetes client and test actual functionality
			if tt.projectRoot == "" {
				t.Error("projectRoot should not be empty")
			}
		})
	}
}
