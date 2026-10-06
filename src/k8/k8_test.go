package k8

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func fakeCluster(t *testing.T, objects ...runtime.Object) *fake.Clientset {
	t.Helper()
	previous := getClient()
	client := fake.NewSimpleClientset(objects...)
	ConfigureClient(client)
	t.Cleanup(func() { ConfigureClient(previous) })
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

func TestPendingAndOtherNamespacePodsReserveHostPorts(t *testing.T) {
	oldRange := portRange
	portRange = []int32{30000, 30000}
	t.Cleanup(func() { portRange = oldRange })
	for _, tc := range []struct {
		name, namespace, node string
		selector              map[string]string
	}{
		{"pending", namespace, "", map[string]string{"kubernetes.io/hostname": "worker"}},
		{"other namespace", "another-app", "worker", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "reserved", Namespace: tc.namespace}, Spec: v1.PodSpec{NodeName: tc.node, NodeSelector: tc.selector, Containers: []v1.Container{{Ports: []v1.ContainerPort{{HostPort: 30000}}}}}}
			client := fakeCluster(t, &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}}, pod)
			// client-go's default fake ignores field selectors. Model real API filtering.
			client.PrependReactor("list", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
				list := action.(ktesting.ListAction)
				items := []v1.Pod{}
				if (action.GetNamespace() == "" || action.GetNamespace() == pod.Namespace) && (list.GetListRestrictions().Fields.Empty() || pod.Spec.NodeName == "worker") {
					items = append(items, *pod)
				}
				return true, &v1.PodList{Items: items}, nil
			})
			if _, _, err := FindOpenNodePort(); err == nil {
				t.Fatal("reserved host port allocated again")
			}
		})
	}
}

func TestEmptyClusterAndMissingConfiguration(t *testing.T) {
	fakeCluster(t)
	if _, _, err := FindOpenNodePort(); err == nil {
		t.Fatal("empty cluster accepted")
	}
	ConfigureClient(nil)
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

func TestInitializeUsesExplicitKubeconfigWithoutK3d(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"apiVersion":"v1","kind":"PodList","items":[]}`)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "kubeconfig")
	content := fmt.Sprintf("apiVersion: v1\nkind: Config\nclusters:\n- name: test\n  cluster:\n    server: %s\ncontexts:\n- name: test\n  context:\n    cluster: test\n    user: test\ncurrent-context: test\nusers:\n- name: test\n  user: {}\n", server.URL)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", path)
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")
	previous := getClient()
	ConfigureClient(nil)
	t.Cleanup(func() { ConfigureClient(previous) })
	Initialize()
	if _, _, err := GetAllServerPodsList("", -1, ""); err != nil {
		t.Fatal("explicit kubeconfig was ignored", err)
	}
}

func TestClientReplacementDuringRequests(t *testing.T) {
	previous := getClient()
	t.Cleanup(func() { ConfigureClient(previous) })
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "game", Namespace: namespace}}
	first, second := fake.NewSimpleClientset(pod), fake.NewSimpleClientset(pod)
	ConfigureClient(first)
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		for i := 0; i < 200; i++ {
			ConfigureClient(first)
			ConfigureClient(second)
		}
	}()
	go func() {
		defer workers.Done()
		for i := 0; i < 200; i++ {
			if _, err := GetServerPod("game"); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	workers.Wait()
}

func TestClusterStatusReturnsPodListFailures(t *testing.T) {
	client := fakeCluster(t, &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}})
	client.PrependReactor("list", "pods", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("pod listing failed")
	})
	if _, err := GetClusterStatus(); err == nil {
		t.Fatal("pod listing failure was discarded")
	}
}

func TestAddingLabelsWithoutConfigurationReturnsError(t *testing.T) {
	previous := getClient()
	ConfigureClient(nil)
	t.Cleanup(func() { ConfigureClient(previous) })
	if err := AddServerPodLabel("game", map[string]string{"server": "id"}); err == nil {
		t.Fatal("missing client accepted")
	}
}
