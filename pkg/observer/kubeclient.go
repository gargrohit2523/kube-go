package observer

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeClient struct{}

func newKubeClient() *kubeClient {
	return &kubeClient{}
}

func (k *kubeClient) get(config *Config) *kubernetes.Clientset {
	clientConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeConfig)

	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)

	if err != nil {
		panic(err)
	}

	return kubeClient
}
