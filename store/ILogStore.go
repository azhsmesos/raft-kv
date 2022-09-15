package store

import "raft-kv/model"

// ILogStore 日志存储接口
type ILogStore interface {
	Term() int64
	Index() int64
	Append(entry *model.LogEntry) error
	Commit(index int64) error
}
