package core

import (
	"raft-kv/config"
	"raft-kv/rpc"
	"raft-kv/store"
)

// IRaft 有限状态机
type IRaft interface {
	rpc.IRaftRPC

	State() IRaftState

	config() config.IRaftConfig
	store() store.ILogStore
	handleStateChanged(state IRaftState)
}
