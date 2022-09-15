package rpc

type RequestVoteCmd struct {
	CandidateID  string
	Term         int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteRet struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesCmd struct {
	Term         int
	LeaderID     string
	PrevLogTerm  int
	PrevLogIndex int
	LeaderCommit int
	Entries      []*LogEntry
}

type LogEntry struct {
	Tag     int
	Content []byte
}

type AppendEntriesRet struct {
	Term    int
	Success bool
}

// IRaftRPC raft 基本RPC接口定义
type IRaftRPC interface {
	RequestVote(cmd *RequestVoteCmd, ret *RequestVoteRet) error
	AppendEntries(cmd *AppendEntriesCmd, ret *AppendEntriesRet) error
}
