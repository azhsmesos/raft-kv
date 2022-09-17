package rpc

import "raft-kv/model"

type RequestVoteCmd struct {
	CandidateID  string
	Term         int64
	LastLogIndex int64
	LastLogTerm  int64
}

type RequestVoteRet struct {
	Code RVCode
	Term int64
}

type HeartbeatCmd struct {
	LeaderID string
	Term     int64
}

type HeartbeatRet struct {
	Code HBCode
	Term int64
}

type AppendLogCmd struct {
	LeaderID string
	Term     int64
	Entry    *model.LogEntry
}

type AppendLogRet struct {
	Code         ALCode
	Term         int64
	PrevLogIndex int64
	PrevLogTerm  int64
}

type CommitLogCmd struct {
	LeaderID string
	Term     int64
	Index    int64
}

type CommitLogRet struct {
	Code CLCode
}

type HBCode int
type RVCode int
type ALCode int
type CLCode int

const (
	HBOK           HBCode = iota
	HBTermMismatch HBCode = iota

	RVOK           RVCode = iota
	RVLogMismatch  RVCode = iota
	RVTermMismatch RVCode = iota
	RVVotedAnother RVCode = iota

	ALOK            ALCode = iota
	ALTermMismatch  ALCode = iota
	ALIndexMismatch ALCode = iota
	ALInternalError ALCode = iota

	CLOK            CLCode = iota
	CLLogNotFound   CLCode = iota
	CLInternalError CLCode = iota
)

// IRaftRPC raft 基本RPC接口定义
type IRaftRPC interface {
	Heartbeat(cmd *HeartbeatCmd, ret *HeartbeatRet) error
	AppendLog(cmd *AppendLogCmd, ret *AppendLogRet) error
	CommitLog(cmd *CommitLogCmd, ret *CommitLogRet) error
	RequestVote(cmd *RequestVoteCmd, ret *RequestVoteRet) error
}
