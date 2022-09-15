package core

import (
	"raft-kv/config"
	"raft-kv/roles"
	"raft-kv/rpc"
	"raft-kv/util"
	"sync"
	"time"
)

// FollowerState follower 状态实现
type FollowerState struct {
	RaftStateBase

	mInitOnce  sync.Once
	mStartOnce sync.Once

	mVotedLeaderID        string
	mLeaderHeartbeatClock int64
	mStateChangeHandler   StateChangedHandleFunc
	mEventMap             map[FollowerEvent][]FollowerEventHandler
}

type FollowerEvent int
type FollowerEventHandler func(e FollowerEvent, args ...interface{})
type JobFunc func()

const (
	evFollowerStart                  FollowerEvent = iota
	evFollowerLeaderHeartbeatTimeout FollowerEvent = iota
)

func newFollowerState(term int, cfg config.IRaftConfig, handler StateChangedHandleFunc) *FollowerState {
	fs := new(FollowerState)
	fs.init(term, cfg, handler)
	return fs
}

func (fs *FollowerState) init(term int, cfg config.IRaftConfig, handler StateChangedHandleFunc) {
	fs.mInitOnce.Do(func() {
		fs.RaftStateBase = *newRaftStateBase(term, cfg)
		fs.role = roles.Follower
		fs.mStateChangeHandler = handler
		fs.mEventMap = make(map[FollowerEvent][]FollowerEventHandler)
		fs.registerEventHandlers()
	})
}

func (fs *FollowerState) registerEventHandlers() {
	fs.mEventMap[evFollowerStart] = []FollowerEventHandler{
		fs.afterStartThenBeginWatchLeaderTimeout,
	}
	fs.mEventMap[evFollowerLeaderHeartbeatTimeout] = []FollowerEventHandler{}
}

func (fs *FollowerState) Start() {
	fs.mStartOnce.Do(func() {
		fs.raise(evFollowerStart)
	})
}

func (fs *FollowerState) raise(fe FollowerEvent, args ...interface{}) {
	if handlers, ok := fs.mEventMap[fe]; ok {
		for _, it := range handlers {
			it(fe, args)
		}
	}
}

// afterStartThenBeginWatchLeaderTimeout 监控leader timeout
func (fs *FollowerState) afterStartThenBeginWatchLeaderTimeout(fe FollowerEvent, args ...interface{}) {
	go func() {
		checkTimeoutInterval := util.HeartbeatTimeout / 3
		for range time.Tick(checkTimeoutInterval) {
			// todo watch leader AppendEntries rpc timeout
		}
	}()
}

// whenLeaderHeartbeatTimeoutThenSwitchToCandidateState 当leader timeout 就切换到candidate
func (fs *FollowerState) whenLeaderHeartbeatTimeoutThenSwitchToCandidateState(_ FollowerEvent, args ...interface{}) {
	fn := fs.mStateChangeHandler
	if fn == nil {
		return
	}
	state := newCandidateState(fs.cfg, fs.term, fs.mStateChangeHandler)
	fn(state)
}

func (fs *FollowerState) Role() roles.RaftRole {
	return roles.Follower
}

func (fs *FollowerState) RequestVote(cmd *rpc.RequestVoteCmd, ret *rpc.RequestVoteRet) error {
	if cmd.Term <= fs.term {
		ret.Term = fs.term
		ret.VoteGranted = false
		return nil
	}
	if fs.mVotedLeaderID != "" && fs.mVotedLeaderID != cmd.CandidateID {
		ret.Term = fs.term
		ret.VoteGranted = false
		return nil
	}
	fs.mVotedLeaderID = cmd.CandidateID
	ret.Term = cmd.Term
	ret.VoteGranted = true
	return nil
}

func (fs *FollowerState) AppendEntries(cmd *rpc.AppendEntriesCmd, ret *rpc.AppendEntriesRet) error {
	if cmd.Term < fs.term {
		ret.Term = fs.term
		ret.Success = false
		return nil
	}
	fs.term = cmd.Term
	fs.leaderID = cmd.LeaderID
	fs.mLeaderHeartbeatClock = time.Now().UnixNano()
	if len(cmd.Entries) <= 0 {
		ret.Term = cmd.Term
		ret.Success = true
		return nil
	}
	// todo append logs
	return nil
}

func (fs *FollowerState) StateChangedHandler(handler StateChangedHandleFunc) {
	fs.mStateChangeHandler = handler
}
