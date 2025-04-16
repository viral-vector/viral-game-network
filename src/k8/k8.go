package k8

import (
	"context"
	"fmt"
	"log"
	"bufio"
	"math/rand"
	"sort"
	"strconv"
	"encoding/json"
	k8_k3d "viral-game-network/src/k8/k3d"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var ctx = context.Background()
var config *rest.Config
var namespace = "viral-game-network"
var clientset *kubernetes.Clientset
var portRange = []int32{30000, 30030}
var kubeconfig = "/root/.config/k3d/kubeconfig-viral-game-network.yaml"

func init() {
	kinit, err := k8_k3d.CheckCluster()
	if err != nil {
		log.Println("k8 init error: ", err)
		return
	}
	if !kinit {
		log.Println("k8 init error: K8 Error: Cluster Not Running")
		return
	}

	LoadConfiguration()
}

func LoadConfiguration() error {
	// Create the clientset from the config
	k8_k3d.WrtiteKubeConfig()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Println("k8 config error: ", err)
		return err
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("k8 clientset error: ", err)
		return err
	}
	return nil
}

func CheckCreateCluster() error {
	kinit, err := k8_k3d.CheckCluster()
	if err != nil {
		log.Println("k8 init error: ", err)
		return err
	}
	if !kinit {
		err = k8_k3d.CreateCluster(portRange[0], portRange[1])
		if err != nil {
			return fmt.Errorf("k8 error creating cluster: " + err.Error())
		}
		LoadConfiguration()
		CreateNameSpace()
		// Print versrion
		version, err := clientset.Discovery().ServerVersion()
		if err == nil {
			log.Println("k8 cluster initialized: ", version.String())
		}
	}

	return nil
}

func CheckDeleteCluster() error {
	err := k8_k3d.DeleteCluster()
	if err != nil {
		return fmt.Errorf("k8 error creating cluster: " + err.Error())
	}
	return nil
}

func CreateNameSpace() error {
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}

	existing, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil && existing.Name == namespace {
		return nil
	}

	// Define the namespace
	space := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	// Create the namespace
	result, err := clientset.CoreV1().Namespaces().Create(ctx, space, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %v", err)
	}

	log.Println("Created K8 Namespace: " + result.Name)

	return nil
}

func GetClusterStatus() (map[string]interface{}, error) {
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
	nodes, _ := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	nodeInfo["nodes"] = map[string]interface{}{}
	nodeInfo["total"] = 0
	if nodes != nil {
		nodeInfo["total"] = len(nodes.Items)
		for _, node := range nodes.Items {
			pods, _ := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				FieldSelector: "spec.nodeName=" + node.Name,
			})
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
	var ports []int32
	for i := min; i <= max; i++ {
		ports = append(ports, i)
	}
	return ports
}

func GetNode(name string) (*v1.Node, error) {
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}
	return clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

func FindOpenNodePort() (*v1.Node, int32, error) {
	if clientset == nil {
		return nil, -1, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	// Get the list of nodes
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, -1, fmt.Errorf("error listing nodes: %s", err.Error())
	}

	// Final return values
	nodeIndex := ""
	portFinal := int32(-1)

	// Create map of node ports
	nodePorts := make(map[string][]int32)
	for _, node := range nodes.Items {
		nodePorts[node.Name] = []int32{}
		// Get the list of pods on each node & Collect used ports
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		if err != nil {
			log.Printf("FindOpenNodePort: @ Error %s: %s", node.Name, err.Error())
			continue
		}

		usedPorts := map[string]bool{}
		for _, pod := range pods.Items {
			// Iterate through each container and its ports.
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					// If using hostPort, it will be set in this field.
					if port.HostPort != 0 {
						usedPorts[strconv.Itoa(int(port.HostPort))] = true
					}
				}
			}
		}

		// Set unused ports
		for _, port := range GetPortRange(portRange[0], portRange[1]) {
			if _, exists := usedPorts[strconv.Itoa(int(port))]; !exists {
				nodePorts[node.Name] = append(nodePorts[node.Name], port)
			}
		}
	}

	// Find the node with the most ports & select a random port
	maxPorts := int32(0)
	for node, ports := range nodePorts {
		if int32(len(ports)) > maxPorts {
			maxPorts = int32(len(ports))
			nodeIndex = node
		}
	}
	portFinal = nodePorts[nodeIndex][rand.Intn(len(nodePorts[nodeIndex]))]

	// Get the node
	nodeFinal, err := GetNode(nodeIndex)
	if err != nil {
		return nil, -1, fmt.Errorf("error getting node: %s", err.Error())
	}

	return nodeFinal, portFinal, nil
}

// Kill all server pods
func KillAllServerPods() error {
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error pulling pods: %s", err)
	}
	for _, pod := range pods.Items {
		DeleteServerPod(pod.Name)
	}
	return nil
}

// Get All server Pods
func GetAllServerPods() ([]*v1.Pod, error) {
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error pulling pods: %s", err)
	}

	pods := make([]*v1.Pod, len(list.Items))
	for i := range list.Items {
		pods[i] = &list.Items[i]
	}
	return pods, nil
}

// Locate Server Pod
func LocateServerPod(label string) (*v1.Node, *v1.Pod, error) {
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
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
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
				"app": label,
				"hostport": strconv.Itoa(int(sPort)),
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			NodeSelector: map[string]string{
				"kubernetes.io/hostname": node.Name,
			},
			Containers: []v1.Container{
				{
					Name:    label,
					Image:   image,
					Command: command,
					ImagePullPolicy: v1.PullAlways,
					Ports: []v1.ContainerPort{
						{
							Name:          "tcp-"+strconv.Itoa(int(sPort)),
							ContainerPort: aPort,
							HostPort: 	   sPort,
							Protocol:      v1.ProtocolTCP,
						},
						{
							Name:          "udp-"+strconv.Itoa(int(sPort)),
							ContainerPort: aPort,
							HostPort: 	   sPort,
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
	if clientset == nil {
		return fmt.Errorf("K8 Error: Cluster Not Running")
	}
	err := clientset.CoreV1().Pods(namespace).Delete(
		ctx,
		label,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("DeleteServerPod Erorr: %v", err)
	}
	return nil
}

// Helper function to get the external IP of a node
func GetNodeExternalIP(node *v1.Node) string {
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
	for _, cnt := range pod.Spec.Containers {
		for _, port := range cnt.Ports {
			return port.HostPort
		}
	}
	return int32(0)
}

func GetPodLogParts(pod *v1.Pod, tailLines *int64) ([]string, error) {
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
	if clientset == nil {
		return nil, fmt.Errorf("K8 Error: Cluster Not Running")
	}

	result, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("GetLogsCluster Error: %v", err)
	}
	// Sort events by EventTime (most recent first).
	events := result.Items
	sort.Slice(events, func(i, j int) bool {
		return events[i].EventTime.Time.After(events[j].EventTime.Time)
	})

	// Return only the first 'tail' events.
	if int64(len(events)) > *tailLines {
		events = events[:*tailLines]
	}

	return events, nil
}