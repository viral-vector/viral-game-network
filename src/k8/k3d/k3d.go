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

	portDefinition := fmt.Sprintf("%d-%d:%d-%d@server:%d:direct",
		portMin, portMax, portMin, portMax, 0)

	cmd_b := exec.Command(
		"k3d", "cluster", "create", cluster_name, "-p", portDefinition, "--config", "/etc/rancher/k3d/config.yaml",
		"--registry-config", "/etc/rancher/k3d/registries.yaml",
	)
	output, err := cmd_b.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create k3d cluster: %v\nOutput: %s", err, string(output))
	}
	fmt.Println("Created Cluster: " + cluster_name)

	cmd_d := exec.Command("k3d", "kubeconfig", "write", cluster_name)
	output, err = cmd_d.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig: %v\nOutput: %s", err, string(output))
	}

	return nil
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
