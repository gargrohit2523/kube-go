package chaos

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	kube "github.com/rogarg19/kube-go/pkg/kubeclient"
	"github.com/rogarg19/kube-go/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeChaos struct {
	config *types.ChaosConfig
}

func New(config *types.ChaosConfig) *KubeChaos {
	return &KubeChaos{config: config}
}

func (c *KubeChaos) Do() {

	if c.config.Settings.Duration == 0 {
		c.config.Settings.Duration = 10
	}

	var done = make(chan struct{}, 1)

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go start(c, &wg, ctx)

	<-done

	cancel()

	log.Println("Waiting for all goroutines to finish...")
	wg.Wait()
	log.Println("Monitoring stopped.")
}

func start(this *KubeChaos, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	ns := this.config.Namespace
	deploy := this.config.Deployment

	clientset := getclientset(
		&types.KubeClientConfig{
			KubeConfig: this.config.KubeConfig,
		},
	)

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			deployPods, err := clientset.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=" + deploy})

			if len(deployPods.Items) == 0 || err != nil {
				log.Println("no pods found for deployment : ", deploy)
			}

			for _, pod := range deployPods.Items {
				err := clientset.CoreV1().Pods(ns).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Fatalln("error killing pod : ", pod.Name)
				}
				log.Println("killed pod: ", pod.Name)
			}
		case <-ctx.Done():
			log.Println("stopping gorutine...")
			return
		}
	}
}

func getclientset(config *types.KubeClientConfig) *kubernetes.Clientset {
	kClient := kube.NewKubeClient()

	if kClient == nil {
		panic("kube client nil")
	}

	return kClient.Get(config)
}
