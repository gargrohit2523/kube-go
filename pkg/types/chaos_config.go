package types

type ChaosConfig struct {
	TargetObject TargetChaosObject
}

type TargetChaosObject struct {
	ObjectType string
	Namespace  string
}
