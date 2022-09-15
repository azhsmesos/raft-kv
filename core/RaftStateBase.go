package core

import (
	"raft-kv/config"
	"raft-kv/roles"
)

// RaftStateBase 基本状态数据
type RaftStateBase struct {
	// role 当前角色
	role roles.RaftRole

	term int

	leaderID string

	// cfg 集群配置
	cfg config.IRaftConfig
}

// newRaftStateBase init initialize self, with term and config specified
func newRaftStateBase(term int, cfg config.IRaftConfig) *RaftStateBase {
	return &RaftStateBase{
		role:     roles.Follower,
		term:     term,
		leaderID: "",
		cfg:      cfg,
	}
}

func (rsb *RaftStateBase) Role() roles.RaftRole {
	return rsb.role
}
