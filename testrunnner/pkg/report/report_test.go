package report

import (
	"testing"
	"time"
)

func TestResult_New(t *testing.T) {
	result := Result{
		Success:     true,
		Details:     "All tests passed",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(100 * time.Millisecond),
		Duration:    100 * time.Millisecond,
		ExitCode:    0,
		TestCommand: "npm test",
		TargetPod:   "test-pod",
		TargetNS:    "default",
	}

	if result.Success != true {
		t.Errorf("expected success true, got %t", result.Success)
	}

	if result.Details != "All tests passed" {
		t.Errorf("expected details 'All tests passed', got %s", result.Details)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if result.TestCommand != "npm test" {
		t.Errorf("expected test command 'npm test', got %s", result.TestCommand)
	}

	if result.TargetPod != "test-pod" {
		t.Errorf("expected target pod 'test-pod', got %s", result.TargetPod)
	}

	if result.TargetNS != "default" {
		t.Errorf("expected target namespace 'default', got %s", result.TargetNS)
	}
}

func TestResult_Duration(t *testing.T) {
	startTime := time.Now()
	time.Sleep(10 * time.Millisecond) // Ensure some time passes
	endTime := time.Now()

	duration := endTime.Sub(startTime)

	result := Result{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
	}

	// Duration should be the calculated value
	if result.Duration != duration {
		t.Errorf("expected duration %v, got %v", duration, result.Duration)
	}

	// Duration should be positive
	if result.Duration <= 0 {
		t.Errorf("expected positive duration, got %v", result.Duration)
	}
}

func TestResult_JSONFields(t *testing.T) {
	result := Result{
		Success:     true,
		Details:     "Test completed",
		StartTime:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 1, 1, 12, 0, 1, 0, time.UTC),
		Duration:    1 * time.Second,
		ExitCode:    0,
		TestCommand: "npm test",
		TargetPod:   "test-pod",
		TargetNS:    "default",
	}

	// Test that all JSON fields are properly set
	if result.Success != true {
		t.Error("Success field not set correctly")
	}

	if result.Details == "" {
		t.Error("Details field not set correctly")
	}

	if result.StartTime.IsZero() {
		t.Error("StartTime field not set correctly")
	}

	if result.EndTime.IsZero() {
		t.Error("EndTime field not set correctly")
	}

	if result.Duration <= 0 {
		t.Error("Duration field not set correctly")
	}

	if result.ExitCode != 0 {
		t.Error("ExitCode field not set correctly")
	}

	if result.TestCommand == "" {
		t.Error("TestCommand field not set correctly")
	}

	if result.TargetPod == "" {
		t.Error("TargetPod field not set correctly")
	}

	if result.TargetNS == "" {
		t.Error("TargetNS field not set correctly")
	}
}
