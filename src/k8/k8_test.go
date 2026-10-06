package k8

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
)

func fakeCluster(t *testing.T, objects ...runtime.Object) *fake.Clientset {
	t.Helper()
	previous := clientset
	client := fake.NewSimpleClientset(objects...)
	ConfigureClient(client)
	t.Cleanup(func() { clientset = previous })
	return client
}

func TestServerPodLifecycle(t *testing.T) {
	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}, Status: v1.NodeStatus{Addresses: []v1.NodeAddress{{Type: v1.NodeExternalIP, Address: "203.0.113.1"}}}}
	fakeCluster(t, node)
	if err := CreateNameSpace(); err != nil {
		t.Fatal(err)
	}
	if err := CreateNameSpace(); err != nil {
		t.Fatal("namespace creation is not idempotent", err)
	}
	pod, err := CreateServerPod("game", node, 30000, 4000, "game:1", []string{"server"}, map[string]string{"LOBBY_ID": "Lobby:game"})
	if err != nil {
		t.Fatal(err)
	}
	if pod.Namespace != namespace || pod.Spec.RestartPolicy != v1.RestartPolicyNever || pod.Spec.NodeSelector["kubernetes.io/hostname"] != "worker" {
		t.Fatalf("bad placement: %+v", pod.Spec)
	}
	container := pod.Spec.Containers[0]
	if container.Image != "game:1" || len(container.Ports) != 2 || container.Ports[0].Protocol != v1.ProtocolTCP || container.Ports[1].Protocol != v1.ProtocolUDP || GetPodHostPort(pod) != 30000 {
		t.Fatalf("bad server container: %+v", container)
	}
	if container.Env[0].Name != "LOBBY_ID" || container.Env[0].Value != "Lobby:game" {
		t.Fatal(container.Env)
	}
	pod.Spec.NodeName = "worker"
	if _, err := clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	locatedNode, locatedPod, err := LocateServerPod("game")
	if err != nil || locatedNode.Name != "worker" || locatedPod.Name != "game" {
		t.Fatalf("%+v %+v %v", locatedNode, locatedPod, err)
	}
	if err := AddServerPodLabel("game", map[string]string{"server": "record"}); err != nil {
		t.Fatal(err)
	}
	pods, _, err := GetAllServerPodsList("game", 10, "")
	if err != nil || len(pods) != 1 || pods[0].Labels["server"] != "record" {
		t.Fatalf("%v %v", pods, err)
	}
	if err := DeleteServerPod("game"); err != nil {
		t.Fatal(err)
	}
	pods, _, err = GetAllServerPodsList("", -1, "")
	if err != nil || len(pods) != 0 {
		t.Fatal("pod was not deleted", err)
	}
}

func TestNodePortSelectionAndExhaustion(t *testing.T) {
	oldRange := portRange
	portRange = []int32{30000, 30001}
	t.Cleanup(func() { portRange = oldRange })
	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}}
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "used", Namespace: namespace}, Spec: v1.PodSpec{NodeName: "worker", Containers: []v1.Container{{Ports: []v1.ContainerPort{{HostPort: 30000}}}}}}
	fakeCluster(t, node, pod)
	selected, port, err := FindOpenNodePort()
	if err != nil || selected.Name != "worker" || port != 30001 {
		t.Fatalf("%+v %d %v", selected, port, err)
	}
	pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports, v1.ContainerPort{HostPort: 30001})
	clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	if _, _, err := FindOpenNodePort(); err == nil {
		t.Fatal("exhausted node accepted")
	}
}

func TestEmptyClusterAndMissingConfiguration(t *testing.T) {
	fakeCluster(t)
	if _, _, err := FindOpenNodePort(); err == nil {
		t.Fatal("empty cluster accepted")
	}
	clientset = nil
	if _, _, err := GetAllServerPodsList("", 10, ""); err == nil {
		t.Fatal("missing client accepted")
	}
	if _, _, err := LocateServerPod("missing"); err == nil {
		t.Fatal("missing client accepted")
	}
	if err := CreateNameSpace(); err == nil {
		t.Fatal("missing client accepted")
	}
}

func TestNodeAddressAndPortHelpers(t *testing.T) {
	node := &v1.Node{Status: v1.NodeStatus{Addresses: []v1.NodeAddress{{Type: v1.NodeInternalIP, Address: "10.0.0.1"}, {Type: v1.NodeExternalIP, Address: "203.0.113.1"}}}}
	if GetNodeExternalIP(node) != "203.0.113.1" {
		t.Fatal("external IP not selected")
	}
	if !reflect.DeepEqual(GetPortRange(3, 5), []int32{3, 4, 5}) || len(GetPortRange(5, 3)) != 0 {
		t.Fatal("incorrect port range")
	}
}
