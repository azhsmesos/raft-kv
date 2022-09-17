package core

import (
	"raft-kv/model"
	"raft-kv/roles"
	"raft-kv/rpc"
	"sync"
	"time"
)

type CandidateState struct {
	model.EventDrivenModel
	context IRaft

	mInitOnce  sync.Once
	mStartOnce sync.Once

	mTerm int64

	mVotedTerm int64

	mVotedCandidateID string

	mVotedTimestamp int64
}

const (
	ceInit            = "candidate.init"
	ceStart           = "candidate.Start"
	ceElectionTimeout = "cadidate.ElectionTimeout"
	ceLeaderAnnounced = "candidate.LeaderAnounced"
	ceVoteToCandidate = "candidate.VoteToCandidate"
	ceDisposing       = "candidate.Disposing"
)

func newCandidateState(ctx IRaft, term int64) IRaftState {
	cds := new(CandidateState)
	cds.init(ctx, term)
	return cds
}

func (cs *CandidateState) init(ctx IRaft, term int64) {
	cs.mInitOnce.Do(func() {
		cs.context = ctx
		cs.mTerm = term
		cs.initEventHandlers()
		cs.Raise(ceInit)
	})
}

func (cs *CandidateState) initEventHandlers() {
	cs.hookEventsForTerm()
	cs.hookEventsForVotedTerm()
	cs.hookEventsForVotedCandidateID()
	cs.hookEventsForVotedTimestamp()

	cs.Hook(ceLeaderAnnounced, cs.whenLeaderAnnouncedThenSwitchToFollower)
	cs.Hook(ceElectionTimeout, cs.whenElectionTimeoutThenRequestVoteAgain)
}

func (cs *CandidateState) hookEventsForTerm() {
	cs.Hook(ceElectionTimeout, func(e string, args ...interface{}) {
		// when election timeout, term++ and request vote again
		cs.mTerm++
	})
}

func (cs *CandidateState) hookEventsForVotedTerm() {
	cs.Hook(ceInit, func(e string, args ...interface{}) {
		// 第一票投递给自己
		cs.mVotedTerm = cs.mTerm
	})

	cs.Hook(ceElectionTimeout, func(e string, args ...interface{}) {
		cs.mVotedTerm = cs.mTerm
	})

	cs.Hook(ceVoteToCandidate, func(e string, args ...interface{}) {
		cmd := args[0].(*rpc.RequestVoteCmd)
		cs.mVotedTerm = cmd.Term
	})
}

func (cs *CandidateState) hookEventsForVotedCandidateID() {
	cs.Hook(ceInit, func(e string, args ...interface{}) {
		cs.mVotedCandidateID = cs.context.config().ID()
	})

	cs.Hook(ceElectionTimeout, func(e string, args ...interface{}) {
		cs.mVotedCandidateID = cs.context.config().ID()
	})

	cs.Hook(ceVoteToCandidate, func(e string, args ...interface{}) {
		cmd := args[0].(*rpc.RequestVoteCmd)
		cs.mVotedCandidateID = cmd.CandidateID
	})
}

func (cs *CandidateState) hookEventsForVotedTimestamp() {
	cs.Hook(ceInit, func(e string, args ...interface{}) {
		cs.mVotedTimestamp = time.Now().UnixNano()
	})

	cs.Hook(ceElectionTimeout, func(e string, args ...interface{}) {
		cs.mVotedTimestamp = time.Now().UnixNano()
	})

	cs.Hook(ceVoteToCandidate, func(e string, args ...interface{}) {
		cs.mVotedTimestamp = time.Now().UnixNano()
	})
}

// whenLeaderAnnouncedThenSwitchToFollower 当leader确定后就切换成follower
func (cs *CandidateState) whenLeaderAnnouncedThenSwitchToFollower(_ string, _ ...interface{}) {
	cs.Raise(ceDisposing)
	cs.context.handleStateChanged(newFollowerState(cs.context))
}

func (cs *CandidateState) whenElectionTimeoutThenRequestVoteAgain(_ string, _ ...interface{}) {
	panic("implements me")
}

func (cs *CandidateState) Start() {
	cs.mStartOnce.Do(func() {
		cs.Raise(feStart)
	})
}

func (cs *CandidateState) Role() roles.RaftRole {
	return roles.Candidate
}

func (cs *CandidateState) Heartbeat(cmd *rpc.HeartbeatCmd, ret *rpc.HeartbeatRet) error {
	if cmd.Term <= cs.mTerm {
		ret.Code = rpc.HBTermMismatch
		return nil
	}
	cs.Raise(ceLeaderAnnounced)
	ret.Code = rpc.HBOK
	return nil
}

func (cs *CandidateState) AppendLog(cmd *rpc.AppendLogCmd, ret *rpc.AppendLogRet) error {
	if cmd.Term <= cs.mTerm {
		ret.Code = rpc.ALTermMismatch
		return nil
	}
	cs.Raise(ceLeaderAnnounced)
	ret.Code = rpc.ALInternalError
	return nil
}

func (cs *CandidateState) CommitLog(cmd *rpc.CommitLogCmd, ret *rpc.CommitLogRet) error {
	// ignore and return
	ret.Code = rpc.CLInternalError
	return nil
}

func (cs *CandidateState) RequestVote(cmd *rpc.RequestVoteCmd, ret *rpc.RequestVoteRet) error {
	panic("")
}
