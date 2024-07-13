package k8

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	k8_k3d "viral-game-network/src/k8/k3d"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var ctx = context.Background()
var config *rest.Config
var namespace = "viral-game-network"
var clientset *kubernetes.Clientset
var portRange = []int32{30000, 30005}

type PodService struct {
	Pod     v1.Pod
	Service *v1.Service
}

func init() {
	var err error

	kinit, err := k8_k3d.CheckCluster()
	if err != nil {
		log.Println("k8 init error: ", err)
		return
	}
	if !kinit {
		log.Println("k8 init error: Cluster not running")
		return
	}

	LoadConfiguration()
}

func LoadConfiguration() error {
	var err error

	// Create the clientset from the config
	config, err = clientcmd.BuildConfigFromFlags("", "/root/.config/k3d/kubeconfig-viral-game-network.yaml")
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
	var err error

	kinit, err := k8_k3d.CheckCluster()
	if err != nil {
		log.Println("k8 init error: ", err)
		return err
	}
	if !kinit {
		erro := k8_k3d.CreateCluster(portRange[0], portRange[1])
		if erro != nil {
			return fmt.Errorf("k8 error creating cluster: " + erro.Error())
		}
		// Load the configuration
		LoadConfiguration()
		// Namespace
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
	erro := k8_k3d.DeleteCluster()
	if erro != nil {
		return fmt.Errorf("k8 error creating cluster: " + erro.Error())
	}
	return nil
}

func CreateNameSpace() error {
	if clientset == nil {
		return fmt.Errorf("kubernetes not initialized")
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

	fmt.Println("Created K8 Namespace: " + result.Name)

	return nil
}

func GetClusterStatus() (map[string]interface{}, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes not initialized")
	}

	// Get Version
	verInfo := ""
	version, _ := clientset.Discovery().ServerVersion()
	if version != nil {
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
				"ipv4":  GetNodeExternalIP(node),
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

func FindOpenNodePort() (string, int32, error) {
	if clientset == nil {
		return "", -1, fmt.Errorf("kubernetes not initialized")
	}

	// Get the list of nodes
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", -1, fmt.Errorf("error listing nodes: %s", err.Error())
	}

	// Final return values
	nodeFinal := ""
	portFinal := int32(-1)

	// Create map of node ports
	nodePorts := make(map[string][]int32)
	for _, node := range nodes.Items {
		nodePorts[node.Name] = []int32{}
		// Get the list of pods on each node & Collect used ports
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		if err != nil {
			log.Printf("error listing pods on node %s: %s", node.Name, err.Error())
			continue
		}

		usedPorts := map[string]bool{}
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					usedPorts[strconv.Itoa(int(port.ContainerPort))] = true
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
			nodeFinal = node
		}
	}
	portFinal = nodePorts[nodeFinal][rand.Intn(len(nodePorts[nodeFinal]))]

	return nodeFinal, portFinal, nil
}

// Kill all server pods
func KillAllServerPods() error {
	if clientset == nil {
		return fmt.Errorf("kubernetes not initialized")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error pulling pods: %s", err)
	}
	for _, pod := range pods.Items {
		label := strings.Replace(pod.Name, "server-", "", -1)
		DeleteServerPod(label)
	}
	return nil
}

// Get All server Pods
func GetAllServerPodsAndServices() ([]PodService, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes not initialized")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error pulling pods: %s", err)
	}

	var podServices []PodService
	for _, pod := range pods.Items {
		podName := strings.Replace(pod.Name, "server-", "", -1)

		service, err := LocateService(podName)
		if err != nil {
			log.Printf("Error locating service for pod %s: %s", podName, err.Error())
		}
		podServices = append(podServices, PodService{
			Pod:     pod,
			Service: service,
		})
	}

	return podServices, nil
}

// Locate Server Pod
func LocateServerPod(label string) (*v1.Node, *v1.Pod, *v1.Service, error) {
	if clientset == nil {
		return nil, nil, nil, fmt.Errorf("kubernetes not initialized")
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, "server-"+label, metav1.GetOptions{})
	if err != nil {
		return nil, nil, nil, err
	}

	node, err := clientset.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, nil, err
	}

	service, err := LocateService(label)
	if err != nil {
		return nil, nil, nil, err
	}

	return node, pod, service, nil
}

// Create Server Pod
func CreateServerPod(label string, node string, sPort int32, aPort int32, image string, command []string) (*v1.Pod, *v1.Service, error) {
	if clientset == nil {
		return nil, nil, fmt.Errorf("kubernetes not initialized")
	}

	// Create Pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-" + label,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "server-" + label,
			},
		},
		Spec: v1.PodSpec{
			// NodeName:      node,
			HostNetwork:   true,
			RestartPolicy: v1.RestartPolicyNever,
			NodeSelector: map[string]string{
				"kubernetes.io/hostname": node,
			},
			Containers: []v1.Container{
				{
					Name:    "server-" + label,
					Image:   image,
					Command: command,
					Ports: []v1.ContainerPort{
						{
							ContainerPort: aPort,
							Protocol:      v1.ProtocolTCP,
						},
						{
							ContainerPort: aPort,
							Protocol:      v1.ProtocolUDP,
						},
					},
				},
			},
		},
	}

	pod, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Failed to create pod: ", err)
		return nil, nil, err
	}

	ser, err := CreateService(label, aPort, sPort)
	if err != nil {
		fmt.Println("Failed to create pod: ", err)
		DeleteServerPod(label)
		return nil, nil, err
	}

	return pod, ser, nil
}

// Delete Server Pod
func DeleteServerPod(label string) error {
	if clientset == nil {
		return fmt.Errorf("kubernetes not initialized")
	}

	err := clientset.CoreV1().Pods(namespace).Delete(
		ctx,
		"server-"+label,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete pod: %v", err)
	}

	err = DeleteService(label)
	if err != nil {
		return err
	}
	return nil
}

// Locate a Service
func LocateService(label string) (*v1.Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes not initialized")
	}

	service, err := clientset.CoreV1().Services(namespace).Get(ctx, "service-"+label, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Creates a Service
func CreateService(label string, aPort int32, sPort int32) (*v1.Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes not initialized")
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-" + label,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "service-" + label,
			},
		},
		Spec: v1.ServiceSpec{
			Type:                  v1.ServiceTypeNodePort,
			ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyLocal,
			Ports: []v1.ServicePort{
				{
					Name:     "service-" + label + "-port-tcp",
					Port:     aPort,
					NodePort: sPort,
					Protocol: v1.ProtocolTCP,
				},
				{
					Name:     "service-" + label + "-port-udp",
					Port:     aPort,
					NodePort: sPort,
					Protocol: v1.ProtocolUDP,
				},
			},
			Selector: map[string]string{
				"app": "server-" + label,
			},
		},
	}

	service, err := clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %v", err)
	}
	return service, nil
}

// Deletes a Service
func DeleteService(label string) error {
	if clientset == nil {
		return fmt.Errorf("kubernetes not initialized")
	}

	err := clientset.CoreV1().Services(namespace).Delete(
		ctx,
		"service-"+label,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete service: %v", err)
	}

	return nil
}

// Helper function to get the external IP of a node
func GetNodeExternalIP(node v1.Node) string {
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
