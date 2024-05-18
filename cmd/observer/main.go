package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var done = chan struct{}(nil)

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

	var wg sync.WaitGroup

	for _, namespace := range namespaces {
		wg.Add(1)
		go startWatcher(clientset, &namespace, &wg)
	}
	fmt.Println("before wait")

	wg.Wait()

	fmt.Println("after wait")
}

func startWatcher(clientset *kubernetes.Clientset, namespace *string, wg *sync.WaitGroup) {
	defer wg.Done()
	deployments, err := clientset.AppsV1().Deployments(*namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}

	for _, deploy := range deployments.Items {
		deployVar := deploy
		wg.Add(1)
		go monitorPodStatus(clientset, namespace, &deployVar, wg)
		//go monitorEvents(clientset, namespace, &deployVar, wg)
		//go monitorLogs(clientset, namespace, &deployVar, wg)
	}
}

func monitorPodStatus(clientset *kubernetes.Clientset, namespace *string, deployment *v1.Deployment, wg *sync.WaitGroup) {
	defer wg.Done()
	deployPods, err := clientset.CoreV1().Pods(*namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=" + deployment.Name + "-manipura"})

	if err != nil {
		panic(err)
	}

deployLoop:
	for _, pod := range deployPods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if !container.Ready {
				fmt.Println("Deployment not ready ready: ", deployment.Name, "\n Namespace: ", *namespace)
				break deployLoop
			}
		}
	}
}
