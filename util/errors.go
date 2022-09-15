package util

import "errors"

var (
	ErrorCandidateWontReplyRequestVote   = errors.New("candidate won't reply RequestVote RPC")
	ErrorCandidateWontReplyAppendEntries = errors.New("candidate won't reply AppendEntries RPC")
	ErrorCommitLogNotFound               = errors.New("committing log not found")
)
