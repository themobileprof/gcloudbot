package vm

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/themobileprof/gcloudbot/config"
	"github.com/themobileprof/gcloudbot/process"
)

// Configuration constants
const (
	waitLong  = 3 * time.Second
	waitSmall = 1 * time.Second
)

// VMConfig holds the configuration for a VM instance
type VMConfig struct {
	Name         string
	Zone         string
	MachineType  string
	ImageFamily  string
	ImageProject string
	DiskSize     int
	IsFreeVM     bool
}

func VM() {
	fmt.Println(">>> This walk-through will help you easily setup a VM on Google Cloud")
	time.Sleep(waitLong)

	vmName := getVMName()
	vmType := chooseVMType()

	var vm VMConfig
	if vmType == 1 {
		vm = setupFreeVM(vmName, config.Config.Zone)
	} else {
		vm = setupCustomVM(vmName, config.Config.Zone)
	}

	if err := createVM(vm); err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}
}

func getVMName() string {
	fmt.Println("\nFirstly, type the unique name you would like to call this machine (no spaces):")
	name := process.ReadInput()
	// Ensure name follows GCP naming conventions
	sanitized := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	return sanitized + "-" + strconv.Itoa(rand.Intn(10000))
}

func chooseVMType() int {
	fmt.Println("\nSecondly, what type of machine do you want?")
	fmt.Println("1. The Free Instance VM (e2-micro)")
	fmt.Println("2. To setup a custom machine")
	for {
		input := process.ReadInput()
		if val, err := strconv.Atoi(input); err == nil && (val == 1 || val == 2) {
			return val
		}
		fmt.Println("Please enter either 1 or 2")
	}
}

func setupFreeVM(name, defaultZone string) VMConfig {
	zones, err := getUSZones()
	if err != nil {
		log.Fatalf("Failed to get US zones: %v", err)
	}

	zone := chooseZone(zones, defaultZone)

	// Verify e2-micro is available in the selected zone
	if available, err := checkMachineTypeAvailable(zone, "e2-micro"); err != nil || !available {
		log.Fatalf("e2-micro is not available in zone %s", zone)
	}

	return VMConfig{
		Name:         name,
		Zone:         zone,
		MachineType:  "e2-micro",
		ImageFamily:  "debian-11", // Updated to Debian 11
		ImageProject: "debian-cloud",
		DiskSize:     30,
		IsFreeVM:     true,
	}
}

func setupCustomVM(name, zone string) VMConfig {
	osInfo := chooseOS()
	ram := chooseRAM()
	machineType := chooseMachineType(zone, ram)

	return VMConfig{
		Name:         name,
		Zone:         zone,
		MachineType:  machineType,
		ImageFamily:  osInfo.family,
		ImageProject: osInfo.project,
		DiskSize:     200,
		IsFreeVM:     false,
	}
}

func createVM(config VMConfig) error {
	args := []string{
		"compute",
		"instances",
		"create", config.Name,
		"--image-family=" + config.ImageFamily,
		"--image-project=" + config.ImageProject,
		"--machine-type=" + config.MachineType,
		"--boot-disk-size=" + strconv.Itoa(config.DiskSize) + "GB",
		"--zone=" + config.Zone,
	}

	// Add defaults for better security and compatibility
	args = append(args, []string{
		"--shielded-secure-boot",
		"--shielded-vtpm",
		"--shielded-integrity-monitoring",
	}...)

	if config.IsFreeVM {
		args = append(args, "--provisioning-model=STANDARD")
	}

	fmt.Printf("\nCreating VM with command: gcloud %s\n", strings.Join(args, " "))
	if !process.PromptYesNo("Would you like to proceed?") {
		return fmt.Errorf("VM creation cancelled by user")
	}

	output, err := process.ExecuteGCloudCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to create VM: %v\nOutput: %s", err, output)
	}

	fmt.Printf("VM Created successfully!\n%s\n", output)
	return nil
}

func checkMachineTypeAvailable(zone, machineType string) (bool, error) {
	output, err := process.ExecuteGCloudCommand("compute", "machine-types", "describe",
		machineType,
		"--zone", zone,
		"--format=value(name)")
	return output != "", err
}

func chooseZone(zones []string, defaultZone string) string {
	if defaultZone != "" {
		for _, zone := range zones {
			if zone == defaultZone {
				return defaultZone
			}
		}
	}

	// Pick a random zone from the allowed list
	rand.Seed(time.Now().UnixNano())
	return zones[rand.Intn(len(zones))]
}

type osInfo struct {
	family  string
	project string
}

func chooseOS() osInfo {
	fmt.Println("\nWhich Operating System do you want to use?")
	fmt.Println("1. Debian 11 (default)")
	fmt.Println("2. Ubuntu 22.04 LTS")
	fmt.Println("3. Rocky Linux 9")

	choice := process.ReadInput()
	switch choice {
	case "2":
		return osInfo{
			family:  "ubuntu-2204-lts",
			project: "ubuntu-os-cloud",
		}
	case "3":
		return osInfo{
			family:  "rocky-linux-9",
			project: "rocky-linux-cloud",
		}
	default:
		return osInfo{
			family:  "debian-11",
			project: "debian-cloud",
		}
	}
}

func chooseRAM() int {
	fmt.Println("\nHow many GB RAM do you need for your Machine?")
	fmt.Println("Available options: 1GB, 2GB, 4GB, 8GB, 16GB, 32GB, 64GB, 128GB")

	for {
		input := process.ReadInput()
		if ram, err := strconv.Atoi(input); err == nil && isValidRAM(ram) {
			return ram
		}
		fmt.Println("Please enter a valid RAM size in GB")
	}
}

// Helper functions

func getUSZones() ([]string, error) {
	output, err := process.ExecuteGCloudCommand("compute", "zones", "list",
		"--filter=(region:us-west1 OR region:us-east1 OR region:us-central1) AND status=UP",
		"--format=value(name)")
	if err != nil {
		return nil, err
	}
	return strings.Fields(output), nil
}

func getMachineTypes(zone string, ramMB int) ([]string, error) {
	output, err := process.ExecuteGCloudCommand("compute", "machine-types", "list",
		"--filter=zone:"+zone+" AND memoryMb="+strconv.Itoa(ramMB)+" AND name ~ ^e2",
		"--format=value(name)",
		"--sort-by=name")
	if err != nil {
		return nil, err
	}
	return strings.Fields(output), nil
}

func chooseMachineType(zone string, ram int) string {
	machines, err := getMachineTypes(zone, ram*1024)
	if err != nil {
		log.Fatalf("Failed to get machine types: %v", err)
	}

	if len(machines) == 0 {
		log.Fatalf("No machine types found with %dGB RAM in zone %s", ram, zone)
	}

	fmt.Printf("\nAvailable machines with %dGB RAM:\n", ram)
	for i, machine := range machines {
		fmt.Printf("%d: %s\n", i+1, machine)
	}

	for {
		input := process.ReadInput()
		if idx, err := strconv.Atoi(input); err == nil && idx > 0 && idx <= len(machines) {
			return machines[idx-1]
		}
		fmt.Printf("Please enter a number between 1 and %d\n", len(machines))
	}
}

func isValidRAM(ram int) bool {
	validSizes := []int{1, 2, 4, 8, 16, 32, 64, 128}
	for _, size := range validSizes {
		if ram == size {
			return true
		}
	}
	return false
}
