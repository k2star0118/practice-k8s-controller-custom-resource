package worker

import (
	log "github.com/Sirupsen/logrus"
	"k8s-controller-custom-resource/service"
)

// Handler interface contains the methods that are required
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// MyResourceHandler is a sample implementation of Handler
type MyResourceHandler struct{}

// Init handles any Handler initialization
func (t *MyResourceHandler) Init() error {
	log.Info("MyResourceHandler.Init")
	return nil
}

// ObjectCreated is called when an object is created
func (t *MyResourceHandler) ObjectCreated(obj interface{}) {
	log.Info("MyResourceHandler.ObjectCreated")
	// log.Info("MyResource is: %v", obj.(*v1.MyResource).Spec.Message)
	service.CreateHttp(obj)
}

// ObjectDeleted is called when an object is deleted
func (t *MyResourceHandler) ObjectDeleted(obj interface{}) {
	log.Info("MyResourceHandler.ObjectDeleted")
	service.DeleteHttp(obj)
}

// ObjectUpdated is called when an object is updated
func (t *MyResourceHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("MyResourceHandler.ObjectUpdated")
	service.UpdateHttp(objOld, objNew)
}
