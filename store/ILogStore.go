package store

import "raft-kv/model"

// ILogStore 日志存储接口
type ILogStore interface {
	LastAppendTerm() int64
	LastAppendIndex() int64
	LastCommitTerm() int64
	LastCommitIndex() int64

	Append(entry *model.LogEntry) error
	Commit(index int64) error
	GetLog(index int64) (error, *model.LogEntry)
}
