package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

func main() {

	client, err := getKubernetesClient()
	if err != nil {
		panic(err.Error())
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Pods(meta_v1.NamespaceDefault).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Pods(meta_v1.NamespaceDefault).Watch(options)
			},
		},
		&api_v1.Pod{},
		0,
		cache.Indexers{},
	)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(newObj); err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err == nil {
				queue.Add(key)
			}
		},
	})

	controller := Controller{
		logger:    log.NewEntry(log.New()),
		clientset: client,
		informer:  informer,
		queue:     queue,
		handler:   &TestHandler{},
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	go controller.Run(stopCh)

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm

}

func getKubernetesClient() (kubernetes.Interface, error) {

	config, err := getKubernetesClientConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)

}

func getKubernetesClientConfig() (*rest.Config, error) {

	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) > 0 {
		return rest.InClusterConfig()
	}

	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)

}
