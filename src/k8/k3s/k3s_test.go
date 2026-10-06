package k8_k3s

import "testing"

func TestUnsupportedLifecycleReturnsErrors(t *testing.T) {
	if err := CreateCluster(30000, 30001); err == nil {
		t.Fatal("unsupported creation reported success")
	}
	if err := DeleteCluster(); err == nil {
		t.Fatal("unsupported deletion reported success")
	}
}
