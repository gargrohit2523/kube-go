package main

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/rogarg19/kube-go/pkg/chaos"
	"github.com/rogarg19/kube-go/pkg/types"

	"k8s.io/client-go/util/homedir"
)

func main() {

	var ns *string = flag.String("ns", "default", "| separated namespaces")
	var kubeconfig *string
	var duration = flag.Int("duration", 5, "Time interval in minutes to monitor")

	defaultKubeConfig := ""

	if home := homedir.HomeDir(); home != "" {
		defaultKubeConfig = filepath.Join(home, ".kube", "config")
	}

	kubeconfig = flag.String("kubeconfig", defaultKubeConfig, "Path to kube config file. By default picks from /home/.kubeconfig if available")

	flag.Parse()

	namespaces := strings.Split(*ns, "|")

	kubeChaos := chaos.New(&types.ChaosConfig{
		Namespace:  namespaces[0],
		Deployment: "err-nginx",
		KubeConfig: *kubeconfig,
		Settings: types.ChaosSettings{
			Duration: *duration,
		},
	})

	kubeChaos.Do()
}
