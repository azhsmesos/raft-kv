package core

import (
	"raft-kv/config"
	"raft-kv/rpc"
	"raft-kv/util"
	"sync"
)

type CandidateState struct {
	RaftStateBase
	mInitOnce  sync.Once
	mStartOnce sync.Once

	mStateChangedHandler StateChangedHandleFunc
	mEventMap            map[CandidateEvent][]CandidateEventHandler
}

type CandidateEvent int

const (
	evCandidateStart           CandidateEvent = iota
	evCandidateElectionTimeout CandidateEvent = iota
	evCandidateGotEnoughVotes  CandidateEvent = iota
)

type CandidateEventHandler func(e CandidateEvent, args ...interface{})
type StateChangedHandleFunc func(state IRaftState)

func (cs *CandidateState) Start() {
	cs.mStartOnce.Do(func() {

	})
}

func (cs *CandidateState) raise(ce CandidateEvent, args ...interface{}) {
	if handlers, ok := cs.mEventMap[ce]; ok {
		for _, it := range handlers {
			it(ce, args)
		}
	}
}

func (cs *CandidateState) RequestVote(cmd *rpc.RequestVoteCmd, ret *rpc.RequestVoteRet) error {
	return util.ErrorCandidateWontReplyRequestVote
}

func (cs *CandidateState) AppendEntries(cmd *rpc.AppendEntriesCmd, ret *rpc.AppendEntriesRet) error {
	return util.ErrorCandidateWontReplyAppendEntries
}

func (cs *CandidateState) StateChangeHandler(handler StateChangedHandleFunc) {
	cs.mStateChangedHandler = handler
}

func newCandidateState(cfg config.IRaftConfig, term int, handler StateChangedHandleFunc) *CandidateState {
	cs := new(CandidateState)
	cs.init(cfg, term, handler)
	return cs
}

func (cs *CandidateState) init(cfg config.IRaftConfig, term int, handler StateChangedHandleFunc) {
	cs.mInitOnce.Do(func() {
		cs.cfg = cfg
		cs.term = term
		cs.mStateChangedHandler = handler

		cs.mEventMap = make(map[CandidateEvent][]CandidateEventHandler)
		cs.registerEventHandlers()
	})
}

func (cs *CandidateState) registerEventHandlers() {
	cs.mEventMap[evCandidateStart] = []CandidateEventHandler{
		cs.whenStartThenRequestVote,
		cs.whenStartThenWatchElectionTimeout,
	}
	cs.mEventMap[evCandidateElectionTimeout] = []CandidateEventHandler{
		cs.whenElectionTimeoutThenRequestVoteAgain,
	}
	cs.mEventMap[evCandidateGotEnoughVotes] = []CandidateEventHandler{
		cs.whenGotEnoughVotesThenSwitchToLeader,
	}
}

// whenStartThenRequestVote candidate启动后就请求投票
func (cs *CandidateState) whenStartThenRequestVote(_ CandidateEvent, _ ...interface{}) {

}

// whenStartThenWatchElectionTimeout watch leader 选举超时
func (cs *CandidateState) whenStartThenWatchElectionTimeout(_ CandidateEvent, _ ...interface{}) {

}

// whenElectionTimeoutThenRequestVoteAgain leader选举超时重新选举
func (cs *CandidateState) whenElectionTimeoutThenRequestVoteAgain(_ CandidateEvent, _ ...interface{}) {

}

// whenGotEnoughVotesThenSwitchToLeader 当获取到足够票数时，就切换到leader
func (cs *CandidateState) whenGotEnoughVotesThenSwitchToLeader(_ CandidateEvent, _ ...interface{}) {

}
