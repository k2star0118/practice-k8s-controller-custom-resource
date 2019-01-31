package util

import (
	"os"

	log "github.com/Sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	k8sAppType "k8s.io/client-go/kubernetes/typed/apps/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	myresourceclientset "k8s-controller-custom-resource/pkg/client/clientset/versioned"
)

func GetKubernetsConfig() (*restclient.Config) {
	// construct the path to resolve to `~/.kube/config`
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	return config
}

func GetKubernetesClient() (kubernetes.Interface) {
	config := GetKubernetsConfig()

	// generate the client based off of the config
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	return client
}

func GetMyKubernetesClient() (myresourceclientset.Interface) {
	config := GetKubernetsConfig()

	myresourceClient, err := myresourceclientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	log.Info("Successfully constructed k8s client")
	return myresourceClient
}

// retrieve the Kubernetes cluster client from outside of the cluster
func GetBothKubernetesClient() (kubernetes.Interface, myresourceclientset.Interface) {
	client := GetKubernetesClient()
	myresourceClient := GetMyKubernetesClient()
	return client, myresourceClient
}

func GetDeploymentClient() (k8sAppType.DeploymentInterface) {
	client := GetKubernetesClient()
	deploymentsClient := client.AppsV1().Deployments(apiv1.NamespaceDefault)
	return deploymentsClient
}

