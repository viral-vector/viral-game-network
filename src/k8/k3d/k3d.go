package k8_k3d

import (
	"fmt"
	"os/exec"
	"strings"
)

func DeleteCluster() error {
	cluster_name := "viral-game-network"

	cmd := exec.Command(
		"k3d", "cluster", "delete", cluster_name, "--all",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete k3d cluster: %v\nOutput: %s", err, string(output))
	}
	return nil
}

func CreateCluster(portMin int32, portMax int32) error {
	cluster_name := "viral-game-network"

	// Delete the cluster if it exists
	DeleteCluster()

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

	cmd := exec.Command("k3d", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create k3d cluster: %v\nOutput: %s", err, string(output))
	}
	fmt.Println("Created Cluster: " + cluster_name)

	return WrtiteKubeConfig()
}

func CheckCluster() (bool, error) {
	cluster_name := "viral-game-network"

	cmd := exec.Command("k3d", "cluster", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, cluster_name) {
			return true, nil
		}
	}

	return false, nil
}

func WrtiteKubeConfig() error {
	cluster_name := "viral-game-network"

	cmd := exec.Command("k3d", "kubeconfig", "write", cluster_name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig: %v\nOutput: %s", err, string(output))
	}
	fmt.Printf("kubeconfig: Output: %s", string(output))
	return nil
}
