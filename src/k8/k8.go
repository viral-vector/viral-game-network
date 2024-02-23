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
	erro := CreateCluster()
	if erro != nil {
		log.Fatalf(erro.Error())

		return
	}

	// Create the clientset
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.config/k3d/kubeconfig-viral-game-network.yaml")
    if err != nil {
        log.Fatal(err)
    }
    clientset, err = kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err)
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
}

// Locate Server Pod
func LocateServerPod(label string) (*v1.Pod, error) {
	log.Println("Locate Server Pod")

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, label, metav1.GetOptions{})
    if err != nil {
        return nil, err
    }

	service, _ := locate_service(label)
	if service == nil {
		return nil, fmt.Errorf("Server Pod Not Accessible!")
	}

	return pod, nil
}	

// Create Server Pod
func CreateServerPod(label string) (*v1.Pod, error) {
	log.Println("Create Server Pod")

	label = "server-"+label 
	nPort := int32(8080)

	// Create Pod
	pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: "server",
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
							ContainerPort:  nPort,
						},
					},
                },
            },
        },
    }

    pod, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
    if err != nil {
       return nil, err
    }

	create_service(label, nPort)

	return pod, nil
}

// Delete Server Pod
func DeleteServerPod(label string) error {
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
	log.Println("Locate Servie")

	service, err := clientset.CoreV1().Services(namespace).Get(ctx, label+"-service", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Creates a Service
func create_service(label string, gPort int32) (*v1.Service, error) {
	log.Println("Create Servie")

	nPort := int32(30000)

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "server",
			Namespace: namespace,
			Labels: map[string]string{
				"app": label+ "-service",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Port: gPort,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: gPort,
					},
					NodePort: nPort,
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
		return nil, err
	}
	return service, nil
}

// Deletes a Service
func delete_service(label string) error {
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