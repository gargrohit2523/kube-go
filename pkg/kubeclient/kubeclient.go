package kubeclient

import (
	"github.com/rogarg19/kube-go/pkg/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeClient struct{}

func NewKubeClient() *kubeClient {
	return &kubeClient{}
}

func (k *kubeClient) Get(config *types.KubeClientConfig) *kubernetes.Clientset {
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
