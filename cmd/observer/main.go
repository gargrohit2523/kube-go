package main

import (
	"flag"
	"path/filepath"
	"strings"

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

	ob := observer.New(&observer.Config{
		Namespaces: namespaces,
		Interval:   *interval,
		KubeConfig: *kubeconfig,
	})

	ob.Start()
}
