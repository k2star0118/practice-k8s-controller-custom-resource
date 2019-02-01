package util

import (
	"errors"
	"fmt"
	"log"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	k8sAppType "k8s.io/client-go/kubernetes/typed/apps/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	myresourceclientset "k8s-controller-custom-resource/pkg/client/clientset/versioned"
)

func GetKubernetesConfig() (*restclient.Config, error) {
	// construct the path to resolve to `~/.kube/config`
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		err = errors.New(fmt.Sprintf("GetKubernetesConfig:\n%v", err))
	}

	return config, err
}

func GetKubernetesClient() (kubernetes.Interface, error) {
	config, err := GetKubernetesConfig()
	if err != nil {
		return nil, err
	}

	// generate the client based off of the config
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		err = errors.New(fmt.Sprintf("GetKubernetesClient:\n%v", err))
	}

	return client, err
}

func GetMyKubernetesClient() (myresourceclientset.Interface, error) {
	config, err := GetKubernetesConfig()
	if err != nil {
		return nil, err
	}

	myResourceClient, err := myresourceclientset.NewForConfig(config)
	if err != nil {
		err = errors.New(fmt.Sprintf("GetMyKubernetesClient:\n%v", err))
	}

	return myResourceClient, err
}

// retrieve the Kubernetes cluster client from outside of the cluster
func GetBothKubernetesClient() (kubernetes.Interface, myresourceclientset.Interface) {
	client, err := GetKubernetesClient()
	if err != nil {
		log.Fatal(err)
	}
	myResourceClient, err := GetMyKubernetesClient()
	if err != nil {
		log.Fatal(err)
	}
	return client, myResourceClient
}

func GetDeploymentClient() (k8sAppType.DeploymentInterface) {
	client, err := GetKubernetesClient()
	if err != nil {
		log.Fatal(err)
	}
	deploymentsClient := client.AppsV1().Deployments(apiv1.NamespaceDefault)
	return deploymentsClient
}

