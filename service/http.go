package service

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"k8s-controller-custom-resource/pkg/apis/myresource/v1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s-controller-custom-resource/util"
)

func int32Ptr(i int32) *int32 { return &i }

func createHttpServiceSpec(resource *v1.MyResource) (*appsv1.Deployment) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: resource.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
}

func CreateHttp(obj interface{}) {
	log.Infof("Create http service")
	deploymentsClient := util.GetDeploymentClient()
	myResource := obj.(*v1.MyResource)

	executingDeployment, err := deploymentsClient.Get(myResource.Name, metav1.GetOptions{})

	if err == nil {
		log.Infof("Pods (%s) already created. \n", myResource.Name)
		log.Infof("Pods information:\n%v", executingDeployment)
	} else {
		if errors.IsNotFound(err) {
			log.Infof("Creating deployment (%s) \n", myResource.Name)
			deploymentConfig := createHttpServiceSpec(myResource)
			result, err := deploymentsClient.Create(deploymentConfig)
			if err != nil {
				panic(err)
			}
			log.Infof("Created deployment %s.\n", result.GetObjectMeta().GetName())
		} else {
			log.Errorf("Failed to query resource (%s)\n", myResource.Name)
			panic(err)
		}
	}
}

//TODO: need to change
func UpdateHttp(objOld interface{}, objNew interface{}) {
	deploymentsClient := util.GetDeploymentClient()
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(objOld.(*v1.MyResource).Name, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		//result.Spec.Replicas = int32Ptr(1)                           // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})

	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
}

func DeleteHttp(obj interface{}) {
	deploymentsClient := util.GetDeploymentClient()
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(obj.(*v1.MyResource).Name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
}

func GetHttp() {
	/*list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}*/
}
