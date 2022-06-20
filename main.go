package main

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var rdsInstanceLabel string = "fiware.rds-instance"
var labelToWatch string = "fiware.service-to-cm"
var createdLabel string = "fiware.created-by"
var createdLabelValue string = "k8s-cm-to-service"
var namespaceToWatch string = ""

func main() {
	labelToWatch = os.Getenv("LABEL_TO_WATCH")
	namespaceToWatch = os.Getenv("NAMESPACE_TO_WATCH")
	createdLabelValueEnv := os.Getenv("CREATED_LABEL_VALUE")

	if labelToWatch == "" {
		log.Fatal("No label was provided.")
		return
	}
	if namespaceToWatch == "" {
		log.Info("Will watch all namespaces.")
		namespaceToWatch = metav1.NamespaceAll
	}
	if createdLabelValueEnv == "" {
		log.Info("Will use default label value: %s.", createdLabelValue)
	} else {
		createdLabelValue = createdLabelValueEnv
	}

	// creates the in-cluster config
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Was not able to create an in-cluster config. %v", err)
		return
	}
	// creates the client
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	labelExists, err := labels.NewRequirement(labelToWatch, selection.Exists, []string{})
	if err != nil {
		log.Fatalf("Was not able to build label selector: %v", err)
		return
	}

	optionsModifier := func(options *metav1.ListOptions) {
		options.LabelSelector = labelExists.String()
	}

	watchList := cache.NewFilteredListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", namespaceToWatch, optionsModifier)

	_, controller := cache.NewInformer(
		watchList,
		&v1.ConfigMap{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cm := obj.(*v1.ConfigMap)
				log.Debugf("ConfigMap added: %s.", cm.Name)
				createOrUpdateService(cm, clientset, true)
			},
			DeleteFunc: func(obj interface{}) {
				cm := obj.(*v1.ConfigMap)
				log.Debugf("ConfigMap deleted: %s.", cm.Name)
				deleteService(cm, clientset)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				cm := newObj.(*v1.ConfigMap)
				log.Debugf("ConfigMap added: %s.", cm.Name)
				createOrUpdateService(cm, clientset, false)
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

func deleteService(configMap *v1.ConfigMap, clientset *kubernetes.Clientset) {
	serviceName := configMap.Labels[labelToWatch]
	ctx := context.TODO()

	serviceClient := clientset.CoreV1().Services(configMap.Namespace)
	err := serviceClient.Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("Service %s was not deleted. Err: %v", serviceName, err)
	} else {
		log.Infof("Service %s was deleted.", serviceName)
	}
}

func createOrUpdateService(configMap *v1.ConfigMap, clientset *kubernetes.Clientset, create bool) {
	serviceName := configMap.Labels[labelToWatch]
	rdsInstanceName := configMap.Labels[rdsInstanceLabel]
	if rdsInstanceName == "" {
		log.Errorf("Configmap %s : %s does not provide the rds instance name.", configMap.Namespace, configMap.Name)
		return
	}
	host := configMap.Data["."+rdsInstanceName+"-host"]
	if host == "" {
		log.Errorf("Configmap %s : %s does not provide the rds host.", configMap.Namespace, configMap.Name)
	}
	port := configMap.Data["."+rdsInstanceName+"-port"]
	if port == "" {
		log.Errorf("Configmap %s : %s does not provide the rds port.", configMap.Namespace, configMap.Name)
	}

	serviceSpec := v1.ServiceSpec{Type: v1.ServiceTypeExternalName, ExternalName: host}
	serviceMetadata := metav1.ObjectMeta{Name: serviceName, Namespace: configMap.Namespace, Labels: map[string]string{createdLabel: createdLabelValue}}
	service := v1.Service{Spec: serviceSpec, ObjectMeta: serviceMetadata}

	ctx := context.TODO()

	serviceClient := clientset.CoreV1().Services(configMap.Namespace)

	if create {
		_, err := serviceClient.Create(ctx, &service, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Was not able to create service %v from configmap %v. Err: %v", service, configMap, err)
		} else {
			log.Infof("Created service %v.", service)
		}
	} else {
		_, err := serviceClient.Update(ctx, &service, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("Was not able to update service %v from configmap %v. Err: %v", service, configMap, err)
		} else {
			log.Infof("Updated service %v.", service)
		}
	}

}
