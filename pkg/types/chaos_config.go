package types

type ChaosConfig struct {
	Deployment string
	Namespace  string
	KubeConfig string

	Settings ChaosSettings
}

type ChaosSettings struct {
	Duration int
}
