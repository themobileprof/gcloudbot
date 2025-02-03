package process

import "os/exec"

var ExecuteGCloudCommand = func(args ...string) (string, error) {
	cmd := exec.Command("gcloud", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
