package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gargrohit2523/kube-observer/pkg/observer"

	"k8s.io/client-go/util/homedir"
)

func main() {

	var ns *string = flag.String("ns", "default", "| separated namespaces")
	var kubeconfig *string
	var interval = flag.Int("interval", 5, "Time interval in minutes to monitor")

	defaultKubeConfig := ""

	if home := homedir.HomeDir(); home != "" {
		defaultKubeConfig = filepath.Join(home, ".kube", "config")
	}

	kubeconfig = flag.String("kubeconfig", defaultKubeConfig, "Path to kube config file. By default picks from /home/.kubeconfig if available")

	flag.Parse()

	namespaces := strings.Split(*ns, "|")

	var done = make(chan struct{}, 1)

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

	var wg sync.WaitGroup

	ob := observer.New(&observer.Config{
		Namespaces: namespaces,
		Interval:   *interval,
		KubeConfig: *kubeconfig,
	})

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ob.Start(ctx, &wg)

	<-done

	cancel()

	log.Println("Waiting for all goroutines to finish...")
	wg.Wait()
	log.Println("Monitoring stopped.")
}
