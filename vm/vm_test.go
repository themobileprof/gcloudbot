package vm

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/themobileprof/gcloudbot/process"
)

// Mock functions to simulate process.ReadInput and process.ExecuteGCloudCommand
var mockInput func() string
var mockExecuteGCloud func(args ...string) (string, error)

// Helper function to setup tests
func setupTest(inputs []string) {
	inputIndex := 0
	mockInput = func() string {
		if inputIndex < len(inputs) {
			result := inputs[inputIndex]
			inputIndex++
			return result
		}
		return ""
	}
}

func TestGetVMName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBase string
	}{
		{
			name:     "basic name",
			input:    "testvm",
			wantBase: "testvm-",
		},
		{
			name:     "name with spaces",
			input:    "test vm",
			wantBase: "test-vm-",
		},
		{
			name:     "uppercase name",
			input:    "TestVM",
			wantBase: "testvm-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest([]string{tt.input})
			process.ReadInput = mockInput

			got := getVMName()
			if !strings.HasPrefix(got, tt.wantBase) {
				t.Errorf("getVMName() = %v, want prefix %v", got, tt.wantBase)
			}
			// Check that a number was appended
			suffix := got[len(tt.wantBase):]
			if _, err := strconv.Atoi(suffix); err != nil {
				t.Errorf("getVMName() suffix %v is not a number", suffix)
			}
		})
	}
}

func TestChooseVMType(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		want    int
		wantErr bool
	}{
		{
			name:   "choose free VM",
			inputs: []string{"1"},
			want:   1,
		},
		{
			name:   "choose custom VM",
			inputs: []string{"2"},
			want:   2,
		},
		{
			name:    "invalid input then valid",
			inputs:  []string{"3", "1"},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(tt.inputs)
			process.ReadInput = mockInput

			got := chooseVMType()
			if got != tt.want {
				t.Errorf("chooseVMType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupFreeVM(t *testing.T) {
	mockExecuteGCloud = func(args ...string) (string, error) {
		switch {
		case strings.Contains(strings.Join(args, " "), "zones list"):
			return "us-central1-a us-central1-b", nil
		case strings.Contains(strings.Join(args, " "), "machine-types describe"):
			return "e2-micro", nil
		default:
			return "", nil
		}
	}
	process.ExecuteGCloudCommand = mockExecuteGCloud

	config := setupFreeVM("test-vm", "us-central1-a")

	expected := VMConfig{
		Name:         "test-vm",
		Zone:         "us-central1-a",
		MachineType:  "e2-micro",
		ImageFamily:  "debian-11",
		ImageProject: "debian-cloud",
		DiskSize:     30,
		IsFreeVM:     true,
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("setupFreeVM() = %v, want %v", config, expected)
	}
}

func TestChooseOS(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   osInfo
		system string
	}{
		{
			name:  "choose debian",
			input: "1",
			want: osInfo{
				family:  "debian-11",
				project: "debian-cloud",
			},
			system: "Debian",
		},
		{
			name:  "choose ubuntu",
			input: "2",
			want: osInfo{
				family:  "ubuntu-2204-lts",
				project: "ubuntu-os-cloud",
			},
			system: "Ubuntu",
		},
		{
			name:  "choose rocky",
			input: "3",
			want: osInfo{
				family:  "rocky-linux-9",
				project: "rocky-linux-cloud",
			},
			system: "Rocky Linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest([]string{tt.input})
			process.ReadInput = mockInput

			got := chooseOS()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("chooseOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidRAM(t *testing.T) {
	tests := []struct {
		name string
		ram  int
		want bool
	}{
		{"valid 1GB", 1, true},
		{"valid 2GB", 2, true},
		{"valid 4GB", 4, true},
		{"valid 8GB", 8, true},
		{"valid 16GB", 16, true},
		{"valid 32GB", 32, true},
		{"valid 64GB", 64, true},
		{"valid 128GB", 128, true},
		{"invalid 3GB", 3, false},
		{"invalid 0GB", 0, false},
		{"invalid negative", -1, false},
		{"invalid 256GB", 256, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidRAM(tt.ram); got != tt.want {
				t.Errorf("isValidRAM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateVM(t *testing.T) {
	tests := []struct {
		name    string
		config  VMConfig
		wantErr bool
	}{
		{
			name: "create free VM",
			config: VMConfig{
				Name:         "test-vm",
				Zone:         "us-central1-a",
				MachineType:  "e2-micro",
				ImageFamily:  "debian-11",
				ImageProject: "debian-cloud",
				DiskSize:     30,
				IsFreeVM:     true,
			},
			wantErr: false,
		},
		{
			name: "create custom VM",
			config: VMConfig{
				Name:         "custom-vm",
				Zone:         "us-central1-a",
				MachineType:  "e2-standard-2",
				ImageFamily:  "ubuntu-2204-lts",
				ImageProject: "ubuntu-os-cloud",
				DiskSize:     200,
				IsFreeVM:     false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the ExecuteGCloudCommand to simulate successful VM creation
			mockExecuteGCloud = func(args ...string) (string, error) {
				return "VM created successfully", nil
			}
			process.ExecuteGCloudCommand = mockExecuteGCloud

			// Mock the PromptYesNo to always return true
			process.PromptYesNo = func(string) bool {
				return true
			}

			err := createVM(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("createVM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
