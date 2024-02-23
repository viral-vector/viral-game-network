package k8

import (
	"log"
	"fmt"
	"context"
	"k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var ctx = context.Background()
var namespace = "viral-game-network"
var clientset *kubernetes.Clientset


func init() {
	go func() {
		erro := CreateCluster()
		if erro != nil {
			log.Fatalf(erro.Error())

			return
		}

		// Create the clientset
		config, err := clientcmd.BuildConfigFromFlags("", "/root/.config/k3d/kubeconfig-viral-game-network.yaml")
		if err != nil {
			log.Fatal(err)
			return
		}
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatal(err)
			return
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
			log.Fatal(err)
		}
		fmt.Printf("Created namespace %q.\n", result.Name)

		log.Println(clientset.Discovery().ServerVersion())
	}()
}

func GetAllServerPods() ([]v1.Pod, error) {
	if clientset == nil {
		return nil, fmt.Errorf("Kubernetes not initialized!")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("GetAllServerPods: %s", err)
		return nil, err
	}
	return pods.Items, nil
}

// Locate Server Pod
func LocateServerPod(label string) (*v1.Node, *v1.Pod, error) {
	if clientset == nil {
		return nil, nil, fmt.Errorf("Kubernetes not initialized!")
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, label, metav1.GetOptions{})
    if err != nil {
        return nil, nil, err
    }

	node, err := clientset.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
    if err != nil {
        return nil, nil, err
    }

	service, _ := locate_service(label)
	if service == nil {
		return nil, nil, fmt.Errorf("Server Pod Not Accessible!")
	}

	return node, pod, nil
}

// Create Server Pod
func CreateServerPod(label string, sPort int32) (*v1.Pod, *v1.Service, error) {
	if clientset == nil {
		return nil, nil, fmt.Errorf("Kubernetes not initialized!")
	}

	label = "server-"+label 
	aPort := int32(8080)

	// Create Pod
	pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: label,
			Namespace: namespace,
			Labels: map[string]string{
				"app": label,
			},
        },
        Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
            Containers: []v1.Container{
                {
                    Name:    label,
                    Image:   "gcr.io/google-samples/node-hello:1.0", //TODO: Change to your image
                    // Command: []string{"nginx", "-g", "daemon off;"}, // TODO: Change to your command
					Ports: []v1.ContainerPort{
						{
							ContainerPort:  aPort,
						},
					},
                },
            },
        },
    }

    pod, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
    if err != nil {
		fmt.Println("Failed to create pod: %v", err)
       	return nil, nil, err
    }

	ser, err := create_service(label, aPort, sPort)
	if err != nil {
		fmt.Println("Failed to create pod: %v", err)
		// Delete the server pod if we failed to create the service entry
		DeleteServerPod(label)

		return nil, nil, err
	}

	return pod, ser, nil
}

// Delete Server Pod
func DeleteServerPod(label string) error {
	if clientset == nil {
		return fmt.Errorf("Kubernetes not initialized!")
	}

	err := clientset.CoreV1().Pods(namespace).Delete(
        ctx,
        label,
        metav1.DeleteOptions{},
    )
	if err != nil {
		return fmt.Errorf("Failed to delete pod: %v", err)
	}

	err = delete_service(label)

	return nil
}


// Locate a Service
func locate_service(label string) (*v1.Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("Kubernetes not initialized!")
	}

	service, err := clientset.CoreV1().Services(namespace).Get(ctx, label+"-service", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Creates a Service
func create_service(label string, aPort int32, sPort int32) (*v1.Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("Kubernetes not initialized!")
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: label+ "-service" ,
			Namespace: namespace,
			Labels: map[string]string{
				"app": label+ "-service",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Port: aPort,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: aPort,
					},
					NodePort: sPort,
					Protocol: v1.ProtocolTCP,	
				},
			},
			Selector: map[string]string{
				"app": label,
			},
		},
	}

	service, err := clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Failed to create service: %v", err)
		return nil, err
	}
	return service, nil
}

// Deletes a Service
func delete_service(label string) error {
	if clientset == nil {
		return fmt.Errorf("Kubernetes not initialized!")
	}
	
	err := clientset.CoreV1().Services(namespace).Delete(
        ctx,
        label+"-service",
        metav1.DeleteOptions{},
    )
	if err != nil {
		return fmt.Errorf("Failed to delete service: %v", err)
	}

	return nil
}