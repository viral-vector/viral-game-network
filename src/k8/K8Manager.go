package k8

type K8Manager interface {
	CheckCluster() (bool, error)
	CreateCluster(portMin int32, portMax int32) error
	DeleteCluster() error
	WrtiteKubeConfig() error
} 