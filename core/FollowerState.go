package core

import (
	"raft-kv/model"
	"raft-kv/roles"
	"raft-kv/rpc"
	"raft-kv/util"
	"sync"
	"time"
)

// FollowerState follower 状态实现
type FollowerState struct {
	model.EventDrivenModel

	context IRaft

	mInitOnce    sync.Once
	mStartOnce   sync.Once
	mDisposeOnce sync.Once

	// init时，设置term = store.lastCommitTerm 更新leader heartbeat
	mTerm                     int64
	mLeaderHeartbeatTimestamp int64
	mLeaderID                 string
	mLastVotedTerm            int64
	mLastVotedCandidateID     string
	mLastVotedTimestamp       int64

	mDiseposedFlag bool
}

const (
	feInit                          = "follower.init"
	feStart                  string = "follower.start"
	feLeaderHeartbeat               = "follower.LeaderHeartbeat"
	feLeaderHeartbeatTimeout string = "follower.leaderHeartbeatTimeout"
	feCandidateRequestVote          = "candidate.RequestVote"
	feVoteToCandidate               = "follower.CandidateRequestVote"
	feDisposing                     = "follower.Disposing"
)

func newFollowerState(iraft IRaft) IRaftState {
	fs := new(FollowerState)
	fs.init(iraft)
	return fs
}

func (fs *FollowerState) init(iraft IRaft) {
	fs.mInitOnce.Do(func() {
		fs.context = iraft
		fs.initEventHandlers()
	})
}

func (fs *FollowerState) initEventHandlers() {
	fs.hookEventsForTerm()
	fs.hookEventsForLeaderHeartbeatTiemstamp()
	fs.hookEventsForLeaderID()
	fs.hookEventsForLastVotedTerm()
	fs.hookEventsForLastVotedCandidateID()
	fs.hookEventsForLastVotedTimestamp()
	fs.hookEventsForDisposedFlag()

	fs.Hook(feStart, fs.whenStartThenBeginWatchLeaderTimeout)
	fs.Hook(feLeaderHeartbeatTimeout, fs.whenLeaderHeartbeatTimeoutThenSwitchToCandidateState)
}

// hookEventsForTerm 维护term字段
func (fs *FollowerState) hookEventsForTerm() {
	fs.Hook(feInit, func(e string, args ...interface{}) {
		fs.mTerm = fs.context.store().LastCommitTerm()
	})

	fs.Hook(feLeaderHeartbeat, func(e string, args ...interface{}) {
		cmd := args[0].(*rpc.HeartbeatCmd)
		fs.mTerm = cmd.Term
	})
}

// hookEventsForLeaderHeartbeatTiemstamp 维护 mLeaderHeartbeatClock字段
func (fs *FollowerState) hookEventsForLeaderHeartbeatTiemstamp() {
	fs.Hook(feInit, func(e string, args ...interface{}) {
		fs.mLeaderHeartbeatTimestamp = time.Now().UnixNano()
	})

	fs.Hook(feLeaderHeartbeat, func(s string, args ...interface{}) {
		fs.mLeaderHeartbeatTimestamp = time.Now().UnixNano()
	})

	fs.Hook(feLeaderHeartbeatTimeout, func(e string, args ...interface{}) {
		fs.mLeaderHeartbeatTimestamp = 0
	})
}

func (fs *FollowerState) hookEventsForLastVotedTerm() {
	fs.Hook(feCandidateRequestVote, func(e string, args ...interface{}) {
		now := time.Now().UnixNano()
		// 在投票时得检测上次投票是否超时
		if time.Duration(now-fs.mLastVotedTimestamp)*time.Nanosecond >= util.RandomizeDuration(util.ElectionTimeout) {
			// timeout, reset to empty
			fs.mLastVotedTerm = 0
			fs.mLastVotedCandidateID = ""
			fs.mLastVotedTimestamp = 0
		}
	})

	fs.Hook(feVoteToCandidate, func(e string, args ...interface{}) {
		cmd := args[0].(*rpc.RequestVoteCmd)
		fs.mLastVotedTerm = cmd.Term
	})
}

func (fs *FollowerState) hookEventsForLastVotedCandidateID() {
	fs.Hook(feCandidateRequestVote, func(e string, args ...interface{}) {
		now := time.Now().UnixNano()
		if time.Duration(now-fs.mLastVotedTimestamp)*time.Nanosecond >= util.RandomizeDuration(util.ElectionTimeout) {
			fs.mLastVotedTerm = 0
			fs.mLastVotedCandidateID = ""
			fs.mLastVotedTimestamp = 0
		}
	})

	fs.Hook(feVoteToCandidate, func(e string, args ...interface{}) {
		cmd := args[0].(*rpc.RequestVoteCmd)
		fs.mLastVotedCandidateID = cmd.CandidateID
	})
}

func (fs *FollowerState) hookEventsForLeaderID() {
	fs.Hook(feLeaderHeartbeat, func(e string, args ...interface{}) {
		cmd := args[0].(rpc.HeartbeatCmd)
		fs.mLeaderID = cmd.LeaderID
	})

	fs.Hook(feLeaderHeartbeatTimeout, func(e string, args ...interface{}) {
		fs.mLeaderID = ""
	})
}

func (fs *FollowerState) hookEventsForLastVotedTimestamp() {
	fs.Hook(feCandidateRequestVote, func(e string, args ...interface{}) {
		now := time.Now().UnixNano()
		if time.Duration(now-fs.mLastVotedTimestamp)*time.Nanosecond >= util.RandomizeDuration(util.ElectionTimeout) {
			fs.mLastVotedTerm = 0
			fs.mLastVotedCandidateID = ""
			fs.mLastVotedTimestamp = 0
		}
	})

	fs.Hook(feVoteToCandidate, func(e string, args ...interface{}) {
		fs.mLastVotedTimestamp = time.Now().UnixNano()
	})
}

func (fs *FollowerState) hookEventsForDisposedFlag() {
	fs.Hook(feInit, func(e string, args ...interface{}) {
		fs.mDiseposedFlag = false
	})

	fs.Hook(feDisposing, func(e string, args ...interface{}) {
		fs.mDiseposedFlag = true
	})
}

func (fs *FollowerState) Start() {
	fs.mStartOnce.Do(func() {
		fs.Raise(feStart)
	})
}

func (fs *FollowerState) Role() roles.RaftRole {
	return roles.Follower
}

// whenStartThenBeginWatchLeaderTimeout 监听leader是否超时
func (fs *FollowerState) whenStartThenBeginWatchLeaderTimeout(e string, args ...interface{}) {
	go func() {
		iCheckTimeoutInterval := util.RandomizeDuration(util.HeartbeatTimeout / 3)
		for range time.Tick(iCheckTimeoutInterval) {
			if fs.mDiseposedFlag {
				return
			}
			now := time.Now().UnixNano()
			iHeartbeatTimeoutNanos := util.RandomizeInt64(int64(util.HeartbeatTimeout / time.Nanosecond))
			if now-fs.mLeaderHeartbeatTimestamp >= iHeartbeatTimeoutNanos {
				fs.Raise(feLeaderHeartbeatTimeout)
				return
			}
		}
	}()
}

// whenLeaderHeartbeatTimeoutThenSwitchToCandidateState 当leader心跳超时就切换为cadidate状态
func (fs *FollowerState) whenLeaderHeartbeatTimeoutThenSwitchToCandidateState(_ string, args ...interface{}) {
	fs.Raise(feDisposing)
	fs.context.handleStateChanged(newCandidateState(fs.context, fs.mTerm+1))
}

func (fs *FollowerState) Heartbeat(cmd *rpc.HeartbeatCmd, ret *rpc.HeartbeatRet) error {
	if cmd.Term < fs.mTerm {
		ret.Code = rpc.HBTermMismatch
		ret.Term = fs.mTerm
		return nil
	}
	fs.Raise(feLeaderHeartbeat, cmd)
	ret.Code = rpc.HBOK
	return nil
}

func (fs *FollowerState) AppendLog(cmd *rpc.AppendLogCmd, ret *rpc.AppendLogRet) error {
	ret.Term = fs.mTerm
	if cmd.Term < fs.mTerm {
		ret.Code = rpc.ALTermMismatch
		return nil
	}
	store := fs.context.store()
	entry := cmd.Entry

	// check log, 追加操作必须紧随前一个提交操作
	if entry.PrevIndex != store.LastCommitIndex() || entry.PrevTerm != store.LastCommitTerm() {
		err, log := store.GetLog(entry.Index)
		if err != nil {
			ret.Code = rpc.ALInternalError
			return nil
		}
		if log == nil || log.PrevIndex != entry.PrevIndex || log.PrevTerm != log.PrevTerm {
			// bad log
			ret.Code = rpc.ALIndexMismatch
			ret.PrevLogIndex = store.LastCommitIndex()
			ret.PrevLogTerm = store.LastCommitTerm()
			return nil
		}

		// good log, but old, just ignore it
		ret.Code = rpc.ALOK
		return nil
	}
	// good log
	err := store.Append(entry)
	if err != nil {
		ret.Code = rpc.ALInternalError
		return nil
	} else {
		ret.Code = rpc.ALOK
		return nil
	}
}

func (fs *FollowerState) CommitLog(cmd *rpc.CommitLogCmd, ret *rpc.CommitLogRet) error {
	store := fs.context.store()
	if cmd.Index != store.LastAppendIndex() || cmd.Term != store.LastAppendTerm() {
		ret.Code = rpc.CLLogNotFound
		return nil
	}
	err := store.Commit(cmd.Index)
	if err != nil {
		ret.Code = rpc.CLInternalError
		return nil
	}
	ret.Code = rpc.CLOK
	return nil
}

func (fs *FollowerState) RequestVote(cmd *rpc.RequestVoteCmd, ret *rpc.RequestVoteRet) error {
	// before voting
	fs.Raise(feCandidateRequestVote, cmd)
	if cmd.Term <= fs.mTerm {
		ret.Term = fs.mTerm
		ret.Code = rpc.RVTermMismatch
		return nil
	}
	// check if already voted another
	if fs.mLastVotedTerm >= cmd.Term && fs.mLastVotedCandidateID != "" && fs.mLastVotedCandidateID != cmd.CandidateID {
		ret.Code = rpc.RVVotedAnother
		return nil
	}
	if cmd.LastLogIndex < fs.context.store().LastCommitIndex() {
		ret.Code = rpc.RVLogMismatch
		return nil
	}
	// vote ok
	fs.Raise(feVoteToCandidate, cmd)
	ret.Term = cmd.Term
	ret.Code = rpc.RVOK
	return nil
}
