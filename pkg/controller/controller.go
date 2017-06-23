/*
Copyright 2016 Skippbox, Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubewatch/config"
	"github.com/skippbox/kubewatch/pkg/handlers"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func Controller(conf *config.Config, eventHandler handlers.Handler) {

	rawConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		logrus.Fatal(err)
	}

	restConfig, err := clientcmd.NewDefaultClientConfig(*rawConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	kubeClient := kubernetes.NewForConfigOrDie(restConfig)

	if conf.Resource.Pod {
		watchPods(kubeClient, eventHandler)
	}

	if conf.Resource.Services {
		watchServices(kubeClient, eventHandler)
	}

	if conf.Resource.ReplicationController {
		watchReplicationControllers(kubeClient, eventHandler)
	}

	if conf.Resource.Deployment {
		watchDeployments(kubeClient, eventHandler)
	}

	if conf.Resource.Job {
		watchJobs(kubeClient, eventHandler)
	}

	if conf.Resource.PersistentVolume {
		var servicesStore cache.Store
		servicesStore = watchPersistenVolumes(kubeClient, servicesStore, eventHandler)
	}

	logrus.Fatal(http.ListenAndServe(":8081", nil))
}

func watchPods(clientset *kubernetes.Clientset, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (Pods)
	watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "pods", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&api.Pod{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchServices(clientset *kubernetes.Clientset, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (Services)
	watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "services", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&api.Service{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
			UpdateFunc: eventHandler.ObjectUpdated,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchReplicationControllers(clientset *kubernetes.Clientset, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (ReplicationControllers)
	watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "replicationcontrollers", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&api.ReplicationController{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchDeployments(clientset *kubernetes.Clientset, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (Deployments)
	watchlist := cache.NewListWatchFromClient(clientset.ExtensionsV1beta1().RESTClient(), "deployments", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&v1beta1.Deployment{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchJobs(clientset *kubernetes.Clientset, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (Jobs)
	watchlist := cache.NewListWatchFromClient(clientset.BatchV1().RESTClient(), "jobs", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&batchv1.Job{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchPersistenVolumes(clientset *kubernetes.Clientset, store cache.Store, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (PersistenVolumes)
	watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "persistentvolumes", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&api.PersistentVolume{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}
