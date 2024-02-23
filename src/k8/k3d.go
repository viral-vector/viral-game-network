package k8

import (
	"fmt"	
	"os/exec"
)


func CreateCluster() error {
	cluster_name := "viral-game-network"

	fmt.Println("Create Cluster: " + cluster_name)

	// Define the k3d cluster create 
	cmd_a := exec.Command(
		"k3d", "cluster", "delete", cluster_name, "--all",
	)
	cmd_a.CombinedOutput()

	cmd_b := exec.Command(
		"k3d", "cluster", "create", cluster_name, "-p", "30000-30200:30000-30200@server:0", "--config", "/etc/rancher/k3d/config.yaml",
	)
	output, err := cmd_b.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to create k3d cluster: %v\nOutput: %s", err, string(output))
	}
	fmt.Println("Create Cluster: " + cluster_name + " successful")

	cmd_d := exec.Command("k3d", "kubeconfig", "write", cluster_name)
	output, err = cmd_d.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to write kubeconfig: %v\nOutput: %s", err, string(output))
	}

	return nil 
}