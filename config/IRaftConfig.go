package config

// IRaftConfig 集群配置接口
type IRaftConfig interface {
	ID() string
	Nodes() []IRaftNodeConfig
}
