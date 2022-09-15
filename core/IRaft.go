package core

import "raft-kv/rpc"

// IRaft 有限状态机
type IRaft interface {
	rpc.IRaftRPC

	State() IRaftState
}
