package rpc

type IRaftRPCServer interface {
	BeginServerTCP(port int, r IRaftRPC)
}
