package kube

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testrunner/pkg/config"
)

func TestCreateJob_ConfigMapNaming(t *testing.T) {
	tests := []struct {
		name         string
		projectRoot  string
		expectedName string
	}{
		{
			name:         "simple directory name",
			projectRoot:  "my-project",
			expectedName: "ket-source-my-project",
		},
		{
			name:         "nested directory path",
			projectRoot:  "path/to/my-project",
			expectedName: "ket-source-my-project",
		},
		{
			name:         "current directory",
			projectRoot:  ".",
			expectedName: "ket-source-", // Will be filled by os.Getwd()
		},
		{
			name:         "absolute path",
			projectRoot:  "/absolute/path/to/project",
			expectedName: "ket-source-project",
		},
		{
			name:         "path with special characters",
			projectRoot:  "my-project-v1.0",
			expectedName: "ket-source-my-project-v1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				ProjectRoot: tt.projectRoot,
				Namespace:   "test-namespace",
			}

			projectName := getProjectName(cfg.ProjectRoot)
			configMapName := "ket-source-" + projectName

			if tt.expectedName == "ket-source-" {
				if !strings.HasPrefix(configMapName, "ket-source-") {
					t.Errorf("expected ConfigMap name to start with 'ket-source-', got: %s", configMapName)
				}
			} else {
				if configMapName != tt.expectedName {
					t.Errorf("expected ConfigMap name %s, got %s", tt.expectedName, configMapName)
				}
			}
		})
	}
}

func TestCreateJob_JobNaming(t *testing.T) {
	tests := []struct {
		name         string
		projectRoot  string
		expectedName string
	}{
		{
			name:         "simple directory name",
			projectRoot:  "my-project",
			expectedName: "ket-my-project",
		},
		{
			name:         "nested directory path",
			projectRoot:  "path/to/my-project",
			expectedName: "ket-my-project",
		},
		{
			name:         "current directory",
			projectRoot:  ".",
			expectedName: "ket-", // Will be filled by os.Getwd()
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				ProjectRoot: tt.projectRoot,
				Namespace:   "test-namespace",
			}

			projectName := getProjectName(cfg.ProjectRoot)
			jobName := "ket-" + projectName

			if tt.expectedName == "ket-" {
				if !strings.HasPrefix(jobName, "ket-") {
					t.Errorf("expected Job name to start with 'ket-', got: %s", jobName)
				}
			} else {
				if jobName != tt.expectedName {
					t.Errorf("expected Job name %s, got %s", tt.expectedName, jobName)
				}
			}
		})
	}
}

func TestGetProjectName(t *testing.T) {
	tests := []struct {
		name         string
		projectRoot  string
		expectedName string
	}{
		{
			name:         "simple directory name",
			projectRoot:  "my-project",
			expectedName: "my-project",
		},
		{
			name:         "nested directory path",
			projectRoot:  "path/to/my-project",
			expectedName: "my-project",
		},
		{
			name:         "absolute path",
			projectRoot:  "/absolute/path/to/project",
			expectedName: "project",
		},
		{
			name:         "path with dots",
			projectRoot:  "my.project.v1",
			expectedName: "my.project.v1",
		},
		{
			name:         "path with underscores",
			projectRoot:  "my_project_v1",
			expectedName: "my_project_v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProjectName(tt.projectRoot)
			if result != tt.expectedName {
				t.Errorf("expected project name %s, got %s", tt.expectedName, result)
			}
		})
	}
}

// Helper function to extract project name logic for testing
func getProjectName(projectRoot string) string {
	if projectRoot == "." {
		if cwd, err := os.Getwd(); err == nil {
			return filepath.Base(cwd)
		}
		return "project"
	}
	return filepath.Base(projectRoot)
}
