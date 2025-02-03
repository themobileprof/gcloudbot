package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/themobileprof/gcloudbot/process"
)

// CloudConfig holds the GCloud configuration
type CloudConfig struct {
	Project string
	Zone    string
	Region  string
}

var Config CloudConfig

func init() {
	// Ensure gcloud is installed
	if err := checkGCloudInstalled(); err != nil {
		log.Fatalf("gcloud CLI not found: %v", err)
	}

	fmt.Println(">>> Welcome to the Google Cloud Terminal Robot, press Enter to proceed")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	if !checkRequirements() {
		os.Exit(1)
	}

	config := setupGCloudConfig()
	if config.Zone == "" {
		if err := initializeGCloud(); err != nil {
			log.Fatalf("Failed to initialize GCloud: %v", err)
		}
		config = setupGCloudConfig()
		if config.Zone == "" {
			log.Fatalf("Zone is still not set after initialization")
		}
	}
}

func checkGCloudInstalled() error {
	_, err := exec.LookPath("gcloud")
	return err
}

func setupGCloudConfig() CloudConfig {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "gcloud")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}
	}

	zone, _ := process.ExecuteGCloudCommand("config", "get-value", "compute/zone")
	region, _ := process.ExecuteGCloudCommand("config", "get-value", "compute/region")
	project, _ := process.ExecuteGCloudCommand("config", "get-value", "project")

	return CloudConfig{
		Zone:    strings.TrimSpace(zone),
		Region:  strings.TrimSpace(region),
		Project: strings.TrimSpace(project),
	}
}

func initializeGCloud() error {
	fmt.Println("We cannot detect a Zone in your configuration, so let us help you with configuration.")
	fmt.Println("Please remember to configure your zone as us-west1-[a,b,c], us-central1-[a,b,c], or us-east1-[a,b,c] especially if you need the Google free VM")
	fmt.Println("Press Enter to proceed...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	cmd := exec.Command("gcloud", "init")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkRequirements() bool {
	// Check if user is authenticated
	if output, err := process.ExecuteGCloudCommand("auth", "list", "--filter=status:ACTIVE", "--format=value(account)"); err != nil || output == "" {
		fmt.Println("Error: You are not authenticated with gcloud. Please run 'gcloud auth login' first.")
		return false
	}

	// Check if billing is enabled for the current project
	project, err := process.ExecuteGCloudCommand("config", "get-value", "project")
	if err == nil && project != "" {
		if output, err := process.ExecuteGCloudCommand("beta", "billing", "projects", "describe", strings.TrimSpace(project)); err != nil || !strings.Contains(output, "billingEnabled: true") {
			fmt.Println("Warning: Billing might not be enabled for this project")
		}
	}

	fmt.Println("Requirement(s):")
	fmt.Println("1) You must have attached billing to your project of choice.")
	fmt.Println("For more information, visit: https://cloud.google.com/billing/docs/how-to/create-billing-account#create-new-billing-account")

	return process.PromptYesNo("Have you met the requirements above?")
}
