package worker

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

// Event indicate the informerEvent
type Event struct {
	Key       string
	EventType string
	OldObj    interface {}
}

// Controller struct defines how a controller should encapsulate
// logging, client connectivity, informing (list and watching)
// queueing, and handling of resource changes
type Controller struct {
	Logger    *log.Entry
	Clientset kubernetes.Interface
	Queue     workqueue.RateLimitingInterface
	Informer  cache.SharedIndexInformer
	Handler   Handler
}

// Run is the main path of execution for the controller loop
func (c *Controller) Run(stopCh <-chan struct{}) {
	// handle a panic with logging and exiting
	defer utilruntime.HandleCrash()
	// ignore new items in the Queue but when all goroutines
	// have completed existing items then shutdown
	defer c.Queue.ShutDown()

	c.Logger.Info("Controller.Run: initiating")

	// run the Informer to start listing and watching resources
	go c.Informer.Run(stopCh)

	// do the initial synchronization (one time) to populate resources
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("error syncing cache"))
		return
	}
	c.Logger.Info("Controller.Run: cache sync complete")

	// run the runWorker method every second with a stop channel
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced allows us to satisfy the Controller interface
// by wiring up the Informer's HasSynced method to it
func (c *Controller) HasSynced() bool {
	return c.Informer.HasSynced()
}

// runWorker executes the loop to process new items added to the Queue
func (c *Controller) runWorker() {
	log.Info("Controller.runWorker: starting")

	// invoke processNextItem to fetch and consume the next change
	// to a watched or listed resource
	for c.processNextItem() {
		log.Info("Controller.runWorker: processing next item")
	}

	log.Info("Controller.runWorker: completed")
}

// processNextItem retrieves each queued item and takes the
// necessary Handler action based off of if the item was
// created or deleted
func (c *Controller) processNextItem() bool {
	log.Info("Controller.processNextItem: start")

	// fetch the next item (blocking) from the Queue to process or
	// if a shutdown is requested then return out of this to stop
	// processing
	newEvent, quit := c.Queue.Get()

	// stop the worker loop from running as this indicates we
	// have sent a shutdown message that the Queue has indicated
	// from the Get method
	if quit {
		return false
	}
	defer c.Queue.Done(newEvent)
	err := c.processItem(newEvent.(Event))
	if err == nil {
		// No error, reset the ratelimit counters
		c.Queue.Forget(newEvent)
	} else if c.Queue.NumRequeues(newEvent) < maxRetries {
		c.Logger.Errorf("Error processing %s (will retry):\n%v", newEvent.(Event).Key, err)
		c.Queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		c.Logger.Errorf("Error processing %s (giving up):\n%v", newEvent.(Event).Key, err)
		c.Queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(newEvent Event) error {
	item, _, err := c.Informer.GetIndexer().GetByKey(newEvent.Key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store:\n%v", newEvent.Key, err)
	}

	// process events based on its type
	switch newEvent.EventType {
	case "create":
		c.Handler.ObjectCreated(item)
		return nil
	case "update":
		c.Handler.ObjectUpdated(newEvent.OldObj, item)
		return nil
	case "delete":
		log.Infof("Old obj is:\n%v", newEvent.OldObj)
		c.Handler.ObjectDeleted(newEvent.OldObj)
		return nil
	}

	return nil
}
