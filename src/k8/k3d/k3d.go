package k8_k3d

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func DeleteCluster() error {
	cluster_name := "viral-game-network"

	output, err := runK3d(2*time.Minute, "cluster", "delete", cluster_name)
	if err != nil {
		return fmt.Errorf("failed to delete k3d cluster: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func CreateCluster(portMin int32, portMax int32) error {
	cluster_name := "viral-game-network"
	if portMin < 1 || portMax > 65535 || portMin > portMax {
		return fmt.Errorf("invalid cluster port range: %d-%d", portMin, portMax)
	}

	nodes := []string{"server:0"}

	// Build an array of port definitions using the provided port range.
	var portDefinitions []string
	for _, node := range nodes {
		pd := fmt.Sprintf("%d-%d:%d-%d@%s:direct", portMin, portMax, portMin, portMax, node)
		portDefinitions = append(portDefinitions, pd)
	}

	// Build command arguments.
	cmdArgs := []string{
		"cluster", "create", cluster_name,
		"--registry-config", "/etc/rancher/k3d/registries.yaml",
	}
	// Append each port definition using the -p flag.
	for _, pd := range portDefinitions {
		cmdArgs = append(cmdArgs, "-p", pd)
	}
	// Append additional config file flag.
	cmdArgs = append(cmdArgs, "--config", "/etc/rancher/k3d/config.yaml")

	output, err := runK3d(5*time.Minute, cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to create k3d cluster: %w\nOutput: %s", err, string(output))
	}
	fmt.Println("Created Cluster: " + cluster_name)

	return WrtiteKubeConfig()
}

func CheckCluster() (bool, error) {
	cluster_name := "viral-game-network"

	output, err := runK3d(30*time.Second, "cluster", "list")
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == cluster_name {
			return true, nil
		}
	}

	return false, nil
}

func WrtiteKubeConfig() error {
	cluster_name := "viral-game-network"

	output, err := runK3d(30*time.Second, "kubeconfig", "write", cluster_name)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w\nOutput: %s", err, string(output))
	}
	fmt.Printf("kubeconfig: Output: %s", string(output))
	return nil
}

func runK3d(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	output, err := exec.CommandContext(ctx, "k3d", args...).CombinedOutput()
	if ctx.Err() != nil {
		return output, ctx.Err()
	}
	return output, err
}
