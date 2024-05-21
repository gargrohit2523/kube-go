package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var done = make(chan struct{}, 1)
var monitorStopped = make(chan struct{}, 1)

func main() {
	var ns *string = flag.String("ns", "default", "| separated namespaces")
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	namespaces := strings.Split(*ns, "|")

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

	for _, namespace := range namespaces {
		go startWatcher(clientset, &namespace)
	}

	<-monitorStopped
	fmt.Fprintf(os.Stdout, "Monitoring stopped")
}

func startWatcher(clientset *kubernetes.Clientset, namespace *string) {
	deployments, err := clientset.AppsV1().Deployments(*namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}

	for _, deploy := range deployments.Items {
		deployVar := deploy
		go monitorDeployment(clientset, namespace, &deployVar)
	}
}

func monitorDeployment(clientset *kubernetes.Clientset, namespace *string, deployment *v1.Deployment) {
	defer func() {
		monitorStopped <- struct{}{}
	}()

	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-tick:
			deployPods, err := clientset.CoreV1().Pods(*namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=" + deployment.Name})

			if err != nil {
				panic(err)
			}

		deployLoop:
			for _, pod := range deployPods.Items {
				if pod.Status.Phase != "Running" && time.Since(pod.CreationTimestamp.Time) > time.Duration(2*time.Minute) {
					fmt.Fprintf(os.Stderr, "Deployment: %s in Namespace: %s not ready \n", deployment.Name, *namespace)
				} else if len(pod.Status.ContainerStatuses) > 0 {
					for _, container := range pod.Status.ContainerStatuses {
						if !container.Ready && time.Since(pod.CreationTimestamp.Time) > time.Duration(2*time.Minute) {
							fmt.Fprintf(os.Stderr, "Deployment: %s in Namespace: %s not ready \n", deployment.Name, *namespace)
							break deployLoop
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "No container running in Deployment: %s in Namespace: %s not ready \n", deployment.Name, *namespace)
				}
			}
		case <-done:
			fmt.Fprintf(os.Stdout, "Exiting...\n")
			return
		}
	}
}
