package config

// IRaftNodeConfig 节点配置接口
type IRaftNodeConfig interface {
	ID() string
	Endpoint() string
}
