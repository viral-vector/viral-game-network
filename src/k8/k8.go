package k8

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
	k8_k3d "viral-game-network/src/k8/k3d"
)

var ctx = context.Background()
var namespace = "vgn-app"
var clientset kubernetes.Interface
var clientMutex sync.RWMutex
var clusterLifecycle sync.Mutex
var portRange = []int32{30000, 30030}
var kubeconfig = "/root/.config/k3d/kubeconfig-viral-game-network.yaml"

func Initialize() error {
	if config, err := rest.InClusterConfig(); err == nil {
		return configureFromConfig(config)
	} else if err != rest.ErrNotInCluster {
		return err
	}
	if os.Getenv("KUBECONFIG") != "" {
		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return err
		}
		return configureFromConfig(config)
	}
	running, err := k8_k3d.CheckCluster()
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}
	return LoadConfiguration()
}

// ConfigureClient may replace the client while handlers and jobs are running.
func ConfigureClient(client kubernetes.Interface) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	clientset = client
}

func getClient() kubernetes.Interface {
	clientMutex.RLock()
	defer clientMutex.RUnlock()
	return clientset
}

func configureFromConfig(config *rest.Config) error {
	config = rest.CopyConfig(config)
	config.Timeout = 15 * time.Second
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	ConfigureClient(client)
	return nil
}

func LoadConfiguration() error {
	if err := k8_k3d.WrtiteKubeConfig(); err != nil {
		return err
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	return configureFromConfig(config)
}

func CheckCreateCluster() error {
	clusterLifecycle.Lock()
	defer clusterLifecycle.Unlock()
	running, err := k8_k3d.CheckCluster()
	if err != nil {
		return err
	}
	if !running {
		if err := k8_k3d.CreateCluster(portRange[0], portRange[1]); err != nil {
			return fmt.Errorf("create cluster: %w", err)
		}
	}
	if err := LoadConfiguration(); err != nil {
		return err
	}
	if err := CreateNameSpace(); err != nil {
		return err
	}
	if version, err := getClient().Discovery().ServerVersion(); err == nil {
		log.Println("k8 cluster initialized:", version)
	}
	return nil
}

func CheckDeleteCluster() error {
	clusterLifecycle.Lock()
	defer clusterLifecycle.Unlock()
	if err := k8_k3d.DeleteCluster(); err != nil {
		return fmt.Errorf("delete cluster: %w", err)
	}
	ConfigureClient(nil)
	return nil
}

func CreateNameSpace() error {
	clientset := getClient()
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}

	existing, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil && existing.Name == namespace {
		return nil
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("read namespace: %w", err)
	}

	// Define the namespace
	space := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	// Create the namespace
	result, err := clientset.CoreV1().Namespaces().Create(ctx, space, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	log.Println("Created K8 Namespace: " + result.Name)

	return nil
}

func GetClusterStatus() (map[string]interface{}, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	// Get Version
	verInfo := ""
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("Cluster Critical Error")
	}
	if version == nil {
		return nil, fmt.Errorf("Cluster Version returned nil")
	} else {
		verInfo = version.String()
	}

	// Get Nodes
	nodeInfo := map[string]interface{}{}
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list cluster nodes: %w", err)
	}
	nodeInfo["nodes"] = map[string]interface{}{}
	nodeInfo["total"] = 0
	if nodes != nil {
		nodeInfo["total"] = len(nodes.Items)
		for _, node := range nodes.Items {
			pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				FieldSelector: "spec.nodeName=" + node.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("list pods on node %s: %w", node.Name, err)
			}
			nodeInfo["nodes"].(map[string]interface{})[node.Name] = map[string]interface{}{
				"name":  node.Name,
				"ipv4":  GetNodeExternalIP(&node),
				"pods":  len(pods.Items),
				"phase": node.Status.Phase,
			}
		}
	}

	return map[string]interface{}{
		"version": verInfo,
		"nodes":   nodeInfo,
	}, nil
}

func GetPortRange(min, max int32) []int32 {
	if min < 1 || max > 65535 || min > max {
		return nil
	}
	var ports []int32
	for i := min; i <= max; i++ {
		ports = append(ports, i)
	}
	return ports
}

func GetNode(name string) (*v1.Node, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}
	return clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

func FindOpenNodePort() (*v1.Node, int32, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, -1, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	// Get the list of nodes
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, -1, fmt.Errorf("error listing nodes: %s", err.Error())
	}

	// Read all namespaces so ports used by other applications are reserved too.
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, -1, fmt.Errorf("list host port reservations: %w", err)
	}
	var selected *v1.Node
	var available []int32
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if node.Spec.Unschedulable {
			continue
		}
		ready := true
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
				ready = false
			}
		}
		if !ready {
			continue
		}
		used := make(map[int32]bool)
		for _, pod := range pods.Items {
			if pod.Spec.NodeName != "" {
				if pod.Spec.NodeName != node.Name {
					continue
				}
			} else if !podCanUseNode(&pod, node) {
				continue
			}
			for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
				for _, port := range container.Ports {
					if port.HostPort != 0 {
						used[port.HostPort] = true
					}
				}
			}
		}
		ports := []int32{}
		for _, port := range GetPortRange(portRange[0], portRange[1]) {
			if !used[port] {
				ports = append(ports, port)
			}
		}
		if len(ports) > len(available) {
			selected = node.DeepCopy()
			available = ports
		}
	}
	if len(available) == 0 {
		return nil, -1, fmt.Errorf("no open node ports")
	}
	return selected, available[rand.Intn(len(available))], nil
}

func podCanUseNode(pod *v1.Pod, node *v1.Node) bool {
	for key, value := range pod.Spec.NodeSelector {
		actual := node.Labels[key]
		if key == "kubernetes.io/hostname" && actual == "" {
			actual = node.Name
		}
		if actual != value {
			return false
		}
	}
	return true
}

// Kill all server pods
func KillAllServerPods() error {
	clientset := getClient()
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error pulling pods: %s", err)
	}
	var failures error
	for _, pod := range pods.Items {
		err := clientset.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			failures = errors.Join(failures, fmt.Errorf("delete pod %s: %w", pod.Name, err))
		}
	}
	return failures
}

func GetAllServerPodsPager(
	search string,
	limit int64,
	currentPage int,
	desiredPage int,
	continueToken string,
) ([]*v1.Pod, int, string, error) {
	if limit < 1 || currentPage < 1 || desiredPage < 1 {
		return nil, 0, "", fmt.Errorf("K8 Error: invalid pagination")
	}
	clientset := getClient()
	if clientset == nil {
		return nil, 0, "", fmt.Errorf("K8 Error: Cluster Not Running")
	}

	group, _, err := getServerPodsList(clientset, search, -1, "")
	if err != nil {
		return nil, 0, "", fmt.Errorf("K8 Error: %w", err)
	}
	count := len(group)
	if count == 0 || int64(desiredPage) > (int64(count)-1)/limit+1 {
		return nil, count, "", nil
	}

	// A continuation token points after the previous page. Only the next page
	// can resume it; backward navigation and page jumps restart the list.
	if currentPage == desiredPage-1 && continueToken != "" {
		currentPage = desiredPage
	} else {
		currentPage = 1
		continueToken = ""
	}

	restarted := false
	for currentPage <= desiredPage {
		pods, newContinueToken, err := getServerPodsList(clientset, search, limit, continueToken)
		if apierrors.IsResourceExpired(err) && !restarted {
			// Kubernetes expires list snapshots. Retry once from the beginning.
			restarted = true
			currentPage = 1
			continueToken = ""
			continue
		}
		if err != nil {
			return nil, count, "", fmt.Errorf("K8 Error: %w", err)
		}
		if currentPage == desiredPage {
			return pods, count, newContinueToken, nil
		}
		if newContinueToken == "" {
			return nil, count, "", nil
		}
		continueToken = newContinueToken
		currentPage++
	}

	return nil, count, "", nil
}

// Get All server Pods
func GetAllServerPodsList(search string, limit int64, continueToken string) ([]*v1.Pod, string, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, "", fmt.Errorf("K8 Error: Cluster Not Running")
	}
	return getServerPodsList(clientset, search, limit, continueToken)
}

func getServerPodsList(clientset kubernetes.Interface, search string, limit int64, continueToken string) ([]*v1.Pod, string, error) {
	listOptions := metav1.ListOptions{}
	if continueToken != "" {
		listOptions.Continue = continueToken
	}
	if search != "" {
		listOptions.LabelSelector = "app=" + search
	}
	if limit > 0 {
		listOptions.Limit = limit
	}

	list, err := clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, "", fmt.Errorf("error pulling pods: %w", err)
	}

	pods := make([]*v1.Pod, len(list.Items))
	for i := range list.Items {
		pods[i] = &list.Items[i]
	}
	return pods, list.Continue, nil
}

// GetServerPod reads the pod without requiring it to be scheduled onto a node.
func GetServerPod(label string) (*v1.Pod, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, label, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

// Locate Server Pod
func LocateServerPod(label string) (*v1.Node, *v1.Pod, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, label, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	node, err := clientset.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		return nil, pod, err
	}

	return node, pod, err
}

// Create Server Pod
func CreateServerPod(label string, node *v1.Node, sPort int32, aPort int32, image string, command []string, env map[string]string) (*v1.Pod, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}
	if node == nil || node.Name == "" {
		return nil, fmt.Errorf("game server requires a target node")
	}
	if sPort < 1 || sPort > 65535 || aPort < 1 || aPort > 65535 {
		return nil, fmt.Errorf("game server ports must be between 1 and 65535")
	}
	hostname := node.Labels["kubernetes.io/hostname"]
	if hostname == "" {
		hostname = node.Name
	}

	// Convert map to slice of corev1.EnvVar
	var envVars []v1.EnvVar
	for key, value := range env {
		envVars = append(envVars, v1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// Create Pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      label,
			Namespace: namespace,
			Labels: map[string]string{
				"app":      label,
				"hostport": strconv.Itoa(int(sPort)),
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			NodeSelector: map[string]string{
				"kubernetes.io/hostname": hostname,
			},
			Containers: []v1.Container{
				{
					Name:            label,
					Image:           image,
					Command:         command,
					ImagePullPolicy: v1.PullAlways,
					Ports: []v1.ContainerPort{
						{
							Name:          "tcp-" + strconv.Itoa(int(sPort)),
							ContainerPort: aPort,
							HostPort:      sPort,
							Protocol:      v1.ProtocolTCP,
						},
						{
							Name:          "udp-" + strconv.Itoa(int(sPort)),
							ContainerPort: aPort,
							HostPort:      sPort,
							Protocol:      v1.ProtocolUDP,
						},
					},
					Env: envVars,
				},
			},
		},
	}

	pod, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		log.Println("Failed to create pod: ", err)
		return nil, err
	}

	return pod, nil
}

func AddServerPodLabel(label string, labels map[string]string) error {
	clientset := getClient()
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}
	// Create a JSON patch to add/update the label.
	patchDMap := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labels,
		},
	}

	patchData, err := json.Marshal(patchDMap)
	if err != nil {
		return fmt.Errorf("AddServerPodLabel Error: %w", err)
	}

	_, err = clientset.CoreV1().Pods(namespace).Patch(
		ctx,
		label,
		types.MergePatchType,
		patchData,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("AddServerPodLabel Error: %v", err)
	}
	return nil
}

// Delete Server Pod
func DeleteServerPod(label string) error {
	clientset := getClient()
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}
	err := clientset.CoreV1().Pods(namespace).Delete(
		ctx,
		label,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("DeleteServerPod error: %w", err)
	}
	return nil
}

// Helper function to get the external IP of a node
func GetNodeExternalIP(node *v1.Node) string {
	if node == nil {
		return ""
	}
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeExternalIP {
			return address.Address
		}
	}
	// Fallback to internal IP if external IP is not available
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

// Helper function to get the pods hostport
func GetPodHostPort(pod *v1.Pod) int32 {
	if pod == nil {
		return 0
	}
	for _, cnt := range pod.Spec.Containers {
		for _, port := range cnt.Ports {
			if port.HostPort != 0 {
				return port.HostPort
			}
		}
	}
	return int32(0)
}

func GetPodLogParts(pod *v1.Pod, tailLines *int64) ([]string, error) {
	clientset := getClient()
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	result := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		TailLines: tailLines,
	})

	stream, err := result.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetPodLogParts Error: %q: %v", pod.Name, err)
	}
	defer stream.Close()

	var builder []string
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		builder = append(builder, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("GetPodLogParts Error: %q: %w", pod.Name, err)
	}
	return builder, nil
}

func GetLogsCluster(tailLines *int64) ([]v1.Event, error) {
	return GetLogsClusterWithContext(ctx, tailLines)
}

func GetLogsClusterWithContext(request context.Context, tailLines *int64) ([]v1.Event, error) {
	if tailLines != nil && *tailLines < 0 {
		return nil, fmt.Errorf("event limit must be nonnegative")
	}
	clientset := getClient()
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	result, err := clientset.CoreV1().Events(namespace).List(request, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("GetLogsCluster Error: %w", err)
	}
	// Handle both modern and legacy event timestamps, with stable ties.
	events := result.Items
	sort.Slice(events, func(i, j int) bool {
		a, b := clusterEventTime(events[i]), clusterEventTime(events[j])
		if a.Equal(b) {
			return events[i].Name < events[j].Name
		}
		return a.After(b)
	})

	// Return only the first 'tail' events.
	if tailLines != nil && int64(len(events)) > *tailLines {
		events = events[:*tailLines]
	}

	return events, nil
}

func clusterEventTime(event v1.Event) time.Time {
	latest := event.EventTime.Time
	for _, candidate := range []time.Time{event.FirstTimestamp.Time, event.LastTimestamp.Time} {
		if candidate.After(latest) {
			latest = candidate
		}
	}
	if event.Series != nil && event.Series.LastObservedTime.Time.After(latest) {
		latest = event.Series.LastObservedTime.Time
	}
	if latest.IsZero() {
		latest = event.CreationTimestamp.Time
	}
	return latest
}
