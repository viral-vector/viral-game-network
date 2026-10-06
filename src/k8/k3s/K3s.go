package k8_k3s

import (
	"fmt"
	"os/exec"
	"strings"
)

func CheckCluster() (bool, error) {
	cluster_name := "viral-game-network"

	cmd := exec.Command("kubectl", "get", "nodes", "-o", "wide")
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

func DeleteCluster() error {
	return fmt.Errorf("K3s cluster lifecycle is not implemented")
}

func CreateCluster(portMin int32, portMax int32) error {
	return fmt.Errorf("K3s cluster lifecycle is not implemented")
}
