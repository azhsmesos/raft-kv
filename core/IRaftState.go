package core

import (
	"raft-kv/roles"
	"raft-kv/rpc"
)

// IRaftState 状态接口
type IRaftState interface {
	rpc.IRaftRPC

	Role() roles.RaftRole

	Start()
}
