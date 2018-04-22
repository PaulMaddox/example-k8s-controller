package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	logger    *log.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	handler   Handler
}

func (c *Controller) Run(stopCh <-chan struct{}) {

	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		runtime.HandleError(fmt.Errorf("Error syncing cache"))
	}

	wait.Until(c.runWorker, time.Millisecond, stopCh)

}

func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c Controller) runWorker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {

	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	keyRaw := key.(string)

	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(key) < 5 {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", key, err)
			c.queue.AddRateLimited(key)
		} else {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
			c.queue.Forget(key)
			runtime.HandleError(err)
		}
	}

	if !exists {
		c.logger.Infof("Controller.processNextItem: object deleted detected: %s", keyRaw)
		c.handler.ObjectDeleted(item)
		c.queue.Forget(key)
	} else {
		c.logger.Infof("Controller.processNextItem: object created detected: %s", keyRaw)
		c.handler.ObjectCreated(item)
		c.queue.Forget(key)
	}

	// keep the worker loop running by returning true
	return true

}
