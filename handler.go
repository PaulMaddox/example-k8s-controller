package main

import (
	log "github.com/Sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
)

type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

type TestHandler struct{}

func (t *TestHandler) Init() error {
	return nil
}

func (t *TestHandler) ObjectCreated(obj interface{}) {
	// assert the type to a Pod object to pull out relevant data
	pod := obj.(*core_v1.Pod)
	log.Infof("    ResourceVersion: %s", pod.ObjectMeta.ResourceVersion)
	log.Infof("    NodeName: %s", pod.Spec.NodeName)
	log.Infof("    Phase: %s", pod.Status.Phase)
}

// ObjectDeleted is called when an object is deleted
func (t *TestHandler) ObjectDeleted(obj interface{}) {
	log.Info("TestHandler.ObjectDeleted")
}

// ObjectUpdated is called when an object is updated
func (t *TestHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("TestHandler.ObjectUpdated")
}
