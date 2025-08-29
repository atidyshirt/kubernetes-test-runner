package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"testrunner/pkg/config"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// TestManifests contains all the manifests needed for a test run
type TestManifests struct {
	Namespace      *corev1.Namespace
	ServiceAccount *corev1.ServiceAccount
	Role           *rbacv1.Role
	RoleBinding    *rbacv1.RoleBinding
	Job            *batchv1.Job
}

// GenerateTestManifests generates all manifests needed for a test run
func GenerateTestManifests(cfg config.Config, namespace string) (*TestManifests, error) {
	ns, err := generateNamespaceManifest(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate namespace manifest: %w", err)
	}

	sa, err := generateServiceAccountManifest(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service account manifest: %w", err)
	}

	role, err := generateRoleManifest(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate role manifest: %w", err)
	}

	roleBinding, err := generateRoleBindingManifest(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate role binding manifest: %w", err)
	}

	job, err := generateJobManifest(cfg, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate job manifest: %w", err)
	}

	return &TestManifests{
		Namespace:      ns,
		ServiceAccount: sa,
		Role:           role,
		RoleBinding:    roleBinding,
		Job:            job,
	}, nil
}

// generateNamespaceManifest creates a namespace manifest
func generateNamespaceManifest(namespace string) (*corev1.Namespace, error) {
	namespaceManifest := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	return namespaceManifest, nil
}

// generateServiceAccountManifest creates a service account manifest
func generateServiceAccountManifest(namespace string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}

	return serviceAccount, nil
}

// generateRoleManifest creates a role manifest
func generateRoleManifest(namespace string) (*rbacv1.Role, error) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ket-test-runner",
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "configmaps", "secrets", "persistentvolumeclaims", "endpoints"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "statefulsets", "daemonsets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses", "networkpolicies"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		},
	}

	return role, nil
}

// generateRoleBindingManifest creates a role binding manifest
func generateRoleBindingManifest(namespace string) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ket-test-runner",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "ket-test-runner",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	return roleBinding, nil
}

// generateJobManifest creates a job manifest
func generateJobManifest(cfg config.Config, namespace string) (*batchv1.Job, error) {
	hostProjectRoot := filepath.Join(cfg.WorkspacePath, cfg.ProjectRoot)
	if cfg.ProjectRoot == "." {
		hostProjectRoot = cfg.WorkspacePath
	}

	workingDir, err := calculateWorkingDirectory(cfg.ProjectRoot, cfg.WorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate working directory: %w", err)
	}

	projectName := "project"
	if cfg.ProjectRoot == "." {
		if cwd, err := os.Getwd(); err == nil {
			projectName = filepath.Base(cwd)
		}
	} else {
		projectName = filepath.Base(cfg.ProjectRoot)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ket-%s", projectName),
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &cfg.BackoffLimit,
			ActiveDeadlineSeconds: &cfg.ActiveDeadlineS,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "default",
					RestartPolicy:      corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{
							Name: "source-code",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: hostProjectRoot,
									Type: &[]corev1.HostPathType{corev1.HostPathDirectory}[0],
								},
							},
						},
						{
							Name: "reports",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "test-runner",
							Image:           cfg.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-c",
								cfg.TestCommand,
							},
							WorkingDir: workingDir,
							Env: []corev1.EnvVar{
								{
									Name:  "KET_TEST_NAMESPACE",
									Value: namespace,
								},
								{
									Name:  "KET_PROJECT_ROOT",
									Value: cfg.ProjectRoot,
								},
								{
									Name:  "KET_WORKSPACE_PATH",
									Value: cfg.WorkspacePath,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source-code",
									MountPath: "/workspace",
								},
								{
									Name:      "reports",
									MountPath: "/reports",
								},
							},
						},
					},
				},
			},
		},
	}

	return job, nil
}

// GenerateNamespaceManifest generates a namespace manifest as YAML
func GenerateNamespaceManifest(name string) (string, error) {
	namespace, err := generateNamespaceManifest(name)
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.Marshal(namespace)
	if err != nil {
		return "", fmt.Errorf("failed to marshal namespace to YAML: %w", err)
	}

	return string(yamlData), nil
}

// GenerateJobManifest generates a job manifest as YAML
func GenerateJobManifest(cfg config.Config, namespace string) (string, error) {
	job, err := generateJobManifest(cfg, namespace)
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job to YAML: %w", err)
	}

	return string(yamlData), nil
}
