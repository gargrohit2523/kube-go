package observer

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Observer struct {
	Config *Config
	client *kubernetes.Clientset
}

func New(config *Config) *Observer {
	return &Observer{Config: config}
}

func (ob *Observer) Start() {

	kubeClient := newKubeClient()

	ob.client = kubeClient.get(ob.Config)

	var done = make(chan struct{}, 1)

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

	var wg sync.WaitGroup

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, ns := range ob.Config.Namespaces {
		wg.Add(1)
		go startObserver(ob, ns, ctx, &wg)
	}

	<-done

	cancel()

	log.Println("Waiting for all goroutines to finish...")
	wg.Wait()
	log.Println("Monitoring stopped.")
}

func startObserver(ob *Observer, namespace string, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(ob.Config.Interval) * time.Minute)

	for {
		select {
		case <-ticker.C:
			deployments, err := ob.client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})

			if err != nil {
				panic(err)
			}

			for _, deploy := range deployments.Items {
				deployVar := deploy
				wg.Add(1)
				go monitorDeployment(ob.client, namespace, &deployVar, wg)
			}
		case <-ctx.Done():
			log.Println("Exiting goroutine for " + namespace + " gracefully...")
			return
		}
	}
}

func monitorDeployment(clientset *kubernetes.Clientset, namespace string, deployment *v1.Deployment, wg *sync.WaitGroup) {
	defer wg.Done()

	deployPods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=" + deployment.Name})

	if err != nil {
		panic(err)
	}

	for _, pod := range deployPods.Items {
		if podHealthy := checkPodState(pod); !podHealthy {
			log.Printf("Deployment: %s in Namespace: %s unhealthy \n", deployment.Name, namespace)
		}
	}
}

func checkPodState(pod coreV1.Pod) bool {
	podReady := true
	if pod.Status.Phase != "Running" && time.Since(pod.CreationTimestamp.Time) > time.Duration(2*time.Minute) {
		podReady = false
	} else if len(pod.Status.ContainerStatuses) > 0 {
		for _, container := range pod.Status.ContainerStatuses {
			if !container.Ready && time.Since(pod.CreationTimestamp.Time) > time.Duration(2*time.Minute) {
				podReady = false
				break
			}
		}
	} else {
		podReady = false
	}
	return podReady
}
