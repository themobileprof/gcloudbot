package config

import (
	"os"
	"os/exec"
	"testing"
)

var execCommand = exec.Command

func TestCheckGCloudInstalled(t *testing.T) {
	// Mock exec.LookPath to simulate gcloud not being installed
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	err := checkGCloudInstalled()
	if err == nil {
		t.Errorf("Expected error when gcloud is not installed, got nil")
	}
}

func TestSetupGCloudConfig(t *testing.T) {
	// Mock exec.Command to simulate gcloud config values
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	config := setupGCloudConfig()
	if config.Zone == "" || config.Region == "" || config.Project == "" {
		t.Errorf("Expected non-empty config values, got %+v", config)
	}
}

// TestHelperProcess is a helper function to mock exec.Command
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}
