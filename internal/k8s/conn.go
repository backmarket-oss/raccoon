package k8s

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// AuthenticateToCluster returns a Clientset depending if you are in cluster or out cluster.
func AuthenticateToCluster(location, kubeConfig string) (*kubernetes.Clientset, error) {
	switch location {
	case "in":
		return authenticateInCluster()
	case "out":
		return authenticateOutOfCluster(kubeConfig)
	default:
		return nil, fmt.Errorf("k8s: unknown cluster location, please use either 'in' our 'out', %v", location)
	}
}

func authenticateInCluster() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get client config")
	}
	// creates the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate client set")
	}
	return clientSet, nil
}

func authenticateOutOfCluster(kubeConfig string) (*kubernetes.Clientset, error) {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get client config")
	}

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate client set")
	}
	return clientSet, nil
}
