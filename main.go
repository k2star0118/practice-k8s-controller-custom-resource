package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	myresourceinformer_v1 "k8s-controller-custom-resource/pkg/client/informers/externalversions/myresource/v1"
	"k8s-controller-custom-resource/util"
	"k8s-controller-custom-resource/worker"
)

// main code path
func main() {
	// get the Kubernetes client for connectivity
	client, myresourceClient := util.GetBothKubernetesClient()

	// retrieve our custom resource informer which was generated from
	// the code generator and pass it the custom resource client, specifying
	// we should be looking through all namespaces for listing and watching
	informer := myresourceinformer_v1.NewMyResourceInformer(
		myresourceClient,
		meta_v1.NamespaceAll,
		0,
		cache.Indexers{},
	)

	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// add event handlers to handle the three types of events for resources:
	//  - adding new resources
	//  - updating existing resources
	//  - deleting resources
	var newEvent worker.Event
	var err error
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			newEvent.Key, err = cache.MetaNamespaceKeyFunc(obj)
			newEvent.EventType = "create"
			log.Infof("Add myresource: %s", newEvent.Key)
			if err == nil {
				// add the key to the queue for the handler to get
				queue.Add(newEvent)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newEvent.Key, err = cache.MetaNamespaceKeyFunc(oldObj)
			newEvent.EventType = "update"
			newEvent.OldObj = oldObj
			log.Infof("Update myresource: %s", newEvent.Key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			newEvent.Key, err = cache.MetaNamespaceKeyFunc(obj)
			newEvent.EventType = "delete"
			newEvent.OldObj = obj
			log.Infof("Delete myresource: %s", newEvent.Key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
	})

	// construct the Controller object which has all of the necessary components to
	// handle logging, connections, informing (listing and watching), the queue,
	// and the handler

	controller := worker.Controller {
		Logger:    log.NewEntry(log.New()),
		Clientset: client,
		Informer:  informer,
		Queue:     queue,
		Handler:   &worker.TestHandler{},
	}

	// use a channel to synchronize the finalization for a graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	// run the controller loop to process items
	go controller.Run(stopCh)

	// use a channel to handle OS signals to terminate and gracefully shut
	// down processing
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
