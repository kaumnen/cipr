package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	fmt.Println("Building binary for testing...")

	buildCmd := exec.Command("go", "build", "-o", "/tmp/cipr", "../")
	err := buildCmd.Run()
	if err != nil {
		fmt.Println("Failed to build binary:", err)
		os.Exit(1)
	}

	fmt.Println("Binary built successfully at /tmp/cipr")

	err = os.Chmod("/tmp/cipr", 0755)
	if err != nil {
		fmt.Println("Failed to set executable permission on binary:", err)
		os.Exit(1)
	}

	exitVal := m.Run()

	err = os.Remove("/tmp/cipr")
	if err != nil {
		fmt.Println("Failed to remove binary:", err)
	}

	fmt.Println("Binary at /tmp/cipr removed successfully.")

	os.Exit(exitVal)
}

func ExecuteCommandAndCaptureOutput(args ...string) (string, error) {
	out, err := exec.Command("/tmp/cipr", args[0:]...).Output()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}

	return string(out), err
}

func TestAWSFilterCommand(t *testing.T) {
	expectedOutput := `18.34.0.0/19
16.15.192.0/18
54.231.0.0/16
52.216.0.0/15
18.34.232.0/21
16.15.176.0/20
16.182.0.0/16
3.5.0.0/19
44.192.134.240/28
44.192.140.64/28
2600:1ff0:8000::/39
2600:1f68:8000::/39
2600:1ff8:8000::/40
2600:1ff9:8000::/40
2600:1ffa:8000::/40
2600:1fa0:8000::/39`

	output, err := ExecuteCommandAndCaptureOutput("aws", "--filter-region", "us-east-1", "--filter-service", "s3")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(output))
}
