package types

type ObserverConfig struct {
	Namespaces []string
	Interval   int
	KubeConfig string
}
