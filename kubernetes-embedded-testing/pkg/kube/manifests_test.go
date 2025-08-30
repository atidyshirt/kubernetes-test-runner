package kube

import (
	"regexp"
	"strings"
	"testing"

	"testrunner/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestGenerateTestManifests_CompleteResources(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "test-project",
		Image:           "test-image:latest",
		TestCommand:     "go test ./...",
		BackoffLimit:    2,
		ActiveDeadlineS: 1800,
		WorkspacePath:   "/workspace",
	}

	manifests, err := GenerateTestManifests(cfg, "test-namespace")
	require.NoError(t, err)

	// Test that all resources are generated
	assert.NotNil(t, manifests.Namespace)
	assert.NotNil(t, manifests.ServiceAccount)
	assert.NotNil(t, manifests.Role)
	assert.NotNil(t, manifests.RoleBinding)
	assert.NotNil(t, manifests.Job)

	// Test namespace
	assert.Equal(t, "test-namespace", manifests.Namespace.Name)
	// Namespace has no Spec or Status fields set in our implementation
	assert.Equal(t, metav1.ObjectMeta{Name: "test-namespace"}, manifests.Namespace.ObjectMeta)

	// Test service account
	assert.Equal(t, "default", manifests.ServiceAccount.Name)
	assert.Equal(t, "test-namespace", manifests.ServiceAccount.Namespace)

	// Test role permissions
	role := manifests.Role
	assert.Equal(t, "ket-test-runner", role.Name)
	assert.Equal(t, "test-namespace", role.Namespace)
	assert.Len(t, role.Rules, 4) // core, apps, batch, networking

	// Test role binding
	rb := manifests.RoleBinding
	assert.Equal(t, "ket-test-runner", rb.Name)
	assert.Equal(t, "test-namespace", rb.Namespace)
	assert.Equal(t, "ket-test-runner", rb.RoleRef.Name)
	assert.Equal(t, "default", rb.Subjects[0].Name)

	// Test job configuration
	job := manifests.Job
	assert.Equal(t, "ket-test-project", job.Name)
	assert.Equal(t, "test-namespace", job.Namespace)
	assert.Equal(t, int32(2), *job.Spec.BackoffLimit)
	assert.Equal(t, int64(1800), *job.Spec.ActiveDeadlineSeconds)
}

func TestManifestYAMLOutput(t *testing.T) {
	cfg := config.Config{
		ProjectRoot: ".",
		Image:       "alpine:latest",
		TestCommand: "echo 'hello world'",
	}

	manifests, err := GenerateTestManifests(cfg, "test-ns")
	require.NoError(t, err)

	// Test namespace YAML
	nsYAML, err := yaml.Marshal(manifests.Namespace)
	require.NoError(t, err)
	assert.Contains(t, string(nsYAML), "metadata:")
	assert.Contains(t, string(nsYAML), "name: test-ns")

	// Test service account YAML
	saYAML, err := yaml.Marshal(manifests.ServiceAccount)
	require.NoError(t, err)
	assert.Contains(t, string(saYAML), "metadata:")
	assert.Contains(t, string(saYAML), "name: default")

	// Test role YAML
	roleYAML, err := yaml.Marshal(manifests.Role)
	require.NoError(t, err)
	assert.Contains(t, string(roleYAML), "metadata:")
	assert.Contains(t, string(roleYAML), "name: ket-test-runner")

	// Test role binding YAML
	rbYAML, err := yaml.Marshal(manifests.RoleBinding)
	require.NoError(t, err)
	assert.Contains(t, string(rbYAML), "metadata:")
	assert.Contains(t, string(rbYAML), "name: ket-test-runner")

	// Test job YAML
	jobYAML, err := yaml.Marshal(manifests.Job)
	require.NoError(t, err)
	assert.Contains(t, string(jobYAML), "metadata:")
	assert.Contains(t, string(jobYAML), "echo 'hello world'")
	assert.Contains(t, string(jobYAML), "image: alpine:latest")
}

func TestManifestGeneration_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
	}{
		{
			name: "empty test command",
			cfg: config.Config{
				ProjectRoot: ".",
				Image:       "alpine:latest",
				TestCommand: "",
			},
			expectError: false,
		},
		{
			name: "special characters in test command",
			cfg: config.Config{
				ProjectRoot: ".",
				Image:       "alpine:latest",
				TestCommand: "echo 'test with spaces' && npm test",
			},
			expectError: false,
		},
		{
			name: "very long test command",
			cfg: config.Config{
				ProjectRoot: ".",
				Image:       "alpine:latest",
				TestCommand: strings.Repeat("echo test && ", 100) + "echo done",
			},
			expectError: false,
		},
		{
			name: "complex test command with pipes",
			cfg: config.Config{
				ProjectRoot: ".",
				Image:       "alpine:latest",
				TestCommand: "npm test | tee test.log && cat test.log | grep 'PASS'",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifests, err := GenerateTestManifests(tt.cfg, "test-ns")
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, manifests.Job)
			assert.Equal(t, tt.cfg.TestCommand, manifests.Job.Spec.Template.Spec.Containers[0].Command[2])
		})
	}
}

func TestRBACRuleGeneration(t *testing.T) {
	manifests, err := GenerateTestManifests(config.Config{}, "test-ns")
	require.NoError(t, err)

	role := manifests.Role
	require.NotNil(t, role)

	// Test core API group rules
	coreRule := findRuleByAPIGroup(role.Rules, "")
	require.NotNil(t, coreRule)
	assert.ElementsMatch(t, []string{"pods", "services", "configmaps", "secrets", "persistentvolumeclaims", "endpoints"}, coreRule.Resources)
	assert.ElementsMatch(t, []string{"get", "list", "watch", "create", "update", "patch", "delete"}, coreRule.Verbs)

	// Test apps API group rules
	appsRule := findRuleByAPIGroup(role.Rules, "apps")
	require.NotNil(t, appsRule)
	assert.ElementsMatch(t, []string{"deployments", "statefulsets", "daemonsets"}, appsRule.Resources)
	assert.ElementsMatch(t, []string{"get", "list", "watch", "create", "update", "patch", "delete"}, appsRule.Verbs)

	// Test batch API group rules
	batchRule := findRuleByAPIGroup(role.Rules, "batch")
	require.NotNil(t, batchRule)
	assert.ElementsMatch(t, []string{"jobs", "cronjobs"}, batchRule.Resources)
	assert.ElementsMatch(t, []string{"get", "list", "watch", "create", "update", "patch", "delete"}, batchRule.Verbs)

	// Test networking API group rules
	networkingRule := findRuleByAPIGroup(role.Rules, "networking.k8s.io")
	require.NotNil(t, networkingRule)
	assert.ElementsMatch(t, []string{"ingresses", "networkpolicies"}, networkingRule.Resources)
	assert.ElementsMatch(t, []string{"get", "list", "watch", "create", "update", "patch", "delete"}, networkingRule.Verbs)
}

func TestNamespaceGeneration_Uniqueness(t *testing.T) {
	projectRoot := "test-project"

	// Generate multiple namespaces
	ns1 := GenerateTestNamespace(projectRoot)
	ns2 := GenerateTestNamespace(projectRoot)
	ns3 := GenerateTestNamespace(projectRoot)

	// All should be different
	assert.NotEqual(t, ns1, ns2)
	assert.NotEqual(t, ns2, ns3)
	assert.NotEqual(t, ns1, ns3)

	// All should follow the expected pattern
	pattern := regexp.MustCompile(`^kubernetes-embedded-test-test-project-[a-f0-9]{8}$`)
	assert.True(t, pattern.MatchString(ns1))
	assert.True(t, pattern.MatchString(ns2))
	assert.True(t, pattern.MatchString(ns3))
}

func TestJobVolumeConfiguration(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:   "src",
		WorkspacePath: "/workspace",
	}

	manifests, err := GenerateTestManifests(cfg, "test-ns")
	require.NoError(t, err)

	job := manifests.Job
	require.NotNil(t, job)

	// Test source code volume
	sourceVolume := findVolumeByName(job.Spec.Template.Spec.Volumes, "source-code")
	require.NotNil(t, sourceVolume)
	assert.NotNil(t, sourceVolume.HostPath)
	assert.Equal(t, "/workspace/src", sourceVolume.HostPath.Path)

	// Test reports volume
	reportsVolume := findVolumeByName(job.Spec.Template.Spec.Volumes, "reports")
	require.NotNil(t, reportsVolume)
	assert.NotNil(t, reportsVolume.EmptyDir)

	// Test volume mounts
	container := job.Spec.Template.Spec.Containers[0]
	sourceMount := findVolumeMountByName(container.VolumeMounts, "source-code")
	require.NotNil(t, sourceMount)
	assert.Equal(t, "/workspace", sourceMount.MountPath)

	reportsMount := findVolumeMountByName(container.VolumeMounts, "reports")
	require.NotNil(t, reportsMount)
	assert.Equal(t, "/reports", reportsMount.MountPath)
}

func TestJobContainerConfiguration(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:   "test-project",
		Image:         "node:18-alpine",
		TestCommand:   "npm test",
		WorkspacePath: "/workspace",
	}

	manifests, err := GenerateTestManifests(cfg, "test-ns")
	require.NoError(t, err)

	job := manifests.Job
	require.NotNil(t, job)

	container := job.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "test-runner", container.Name)
	assert.Equal(t, "node:18-alpine", container.Image)
	assert.Equal(t, corev1.PullIfNotPresent, container.ImagePullPolicy)
	assert.Equal(t, "/workspace/test-project", container.WorkingDir)

	// Test command
	assert.Equal(t, []string{"/bin/sh", "-c", "npm test"}, container.Command)

	// Test environment variables
	envVars := make(map[string]string)
	for _, env := range container.Env {
		envVars[env.Name] = env.Value
	}

	assert.Equal(t, "test-ns", envVars["KET_TEST_NAMESPACE"])
	assert.Equal(t, "test-project", envVars["KET_PROJECT_ROOT"])
	assert.Equal(t, "/workspace", envVars["KET_WORKSPACE_PATH"])
}

func TestJobTemplateSpec(t *testing.T) {
	cfg := config.Config{
		ProjectRoot: "test-project",
		Image:       "alpine:latest",
		TestCommand: "echo test",
	}

	manifests, err := GenerateTestManifests(cfg, "test-ns")
	require.NoError(t, err)

	job := manifests.Job
	require.NotNil(t, job)

	// Test pod template spec
	podSpec := job.Spec.Template.Spec
	assert.Equal(t, "default", podSpec.ServiceAccountName)
	assert.Equal(t, corev1.RestartPolicyNever, podSpec.RestartPolicy)

	// Test that we have exactly 2 volumes
	assert.Len(t, podSpec.Volumes, 2)

	// Test that we have exactly 1 container
	assert.Len(t, podSpec.Containers, 1)
}

func TestManifestConsistency(t *testing.T) {
	// Test that multiple calls with same config produce identical manifests
	cfg := config.Config{
		ProjectRoot: "test-project",
		Image:       "alpine:latest",
		TestCommand: "echo test",
	}

	manifests1, err := GenerateTestManifests(cfg, "test-ns")
	require.NoError(t, err)

	manifests2, err := GenerateTestManifests(cfg, "test-ns")
	require.NoError(t, err)

	// The YAML should be identical (ignoring namespace names which have UUIDs)
	// We'll compare the structure by parsing and comparing key fields
	job1 := manifests1.Job
	job2 := manifests2.Job

	assert.Equal(t, job1.Spec.Template.Spec.Containers[0].Image, job2.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, job1.Spec.Template.Spec.Containers[0].Command, job2.Spec.Template.Spec.Containers[0].Command)
	assert.Equal(t, job1.Spec.BackoffLimit, job2.Spec.BackoffLimit)
	assert.Equal(t, job1.Spec.ActiveDeadlineSeconds, job2.Spec.ActiveDeadlineSeconds)
}

// Helper functions
func findRuleByAPIGroup(rules []rbacv1.PolicyRule, apiGroup string) *rbacv1.PolicyRule {
	for _, rule := range rules {
		if len(rule.APIGroups) == 1 && rule.APIGroups[0] == apiGroup {
			return &rule
		}
	}
	return nil
}

func findVolumeByName(volumes []corev1.Volume, name string) *corev1.Volume {
	for _, vol := range volumes {
		if vol.Name == name {
			return &vol
		}
	}
	return nil
}

func findVolumeMountByName(mounts []corev1.VolumeMount, name string) *corev1.VolumeMount {
	for _, mount := range mounts {
		if mount.Name == name {
			return &mount
		}
	}
	return nil
}
