package report

import "time"

// Result represents the result of a test execution
type Result struct {
	Success     bool          `json:"success"`
	Details     string        `json:"details"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	ExitCode    int           `json:"exit_code"`
	TestCommand string        `json:"test_command"`
	TargetPod   string        `json:"target_pod"`
	TargetNS    string        `json:"target_namespace"`
}
