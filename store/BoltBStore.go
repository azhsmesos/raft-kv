package store

import (
	"bytes"
	"encoding/binary"
	"github.com/boltdb/bolt"
	"raft-kv/model"
	"raft-kv/util"
)

var (
	// dataBucket kv键值数据存储
	dataBucket = []byte("data")
	// metaBucket 记录未提交的index和term
	metaBucket        = []byte("meta")
	keyCommittedTerm  = []byte("committed.term")
	keyCommittedIndex = []byte("committed.index")

	defaultTerm  int64 = 0
	defaultIndex int64 = 0

	// unstableBucket 记录收到未提交的日志，重启后会清空
	unstableBucket = []byte("unstable")
	// committedBucket 记录已提交的日志
	committedBucket = []byte("committed")
)

type boltDBStore struct {
	file            string
	lastAppendTerm  int64
	lastAppendIndex int64
	lastCommitTerm  int64
	lastCommitIndex int64
	db              bolt.DB
}

func NewBoltStore(file string) (error, ILogStore) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, nil
	}
	store := new(boltDBStore)
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(metaBucket)
		if err != nil {
			return err
		}
		value := bucket.Get(keyCommittedTerm)
		if value == nil {
			err = bucket.Put(keyCommittedTerm, int64ToBytes(defaultTerm))
			if err != nil {
				return err
			}
			store.lastCommitTerm = defaultTerm
		} else {
			store.lastCommitTerm = byteToInt64(value)
		}

		value = bucket.Get(keyCommittedIndex)
		if value == nil {
			err = bucket.Put(keyCommittedIndex, int64ToBytes(defaultIndex))
			if err != nil {
				return err
			}
			store.lastCommitIndex = defaultIndex
		} else {
			store.lastCommitIndex = byteToInt64(value)
		}

		bucket, err = tx.CreateBucketIfNotExists(dataBucket)
		if err != nil {
			return err
		}
		err = tx.DeleteBucket(unstableBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucket(unstableBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(committedBucket)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err, nil
	}
	return nil, store
}

func int64ToBytes(value int64) []byte {
	buf := bytes.NewBuffer(make([]byte, 8))
	_ = binary.Write(buf, binary.BigEndian, value)
	return buf.Bytes()
}

func byteToInt64(data []byte) int64 {
	var value int64
	buf := bytes.NewBuffer(data)
	_ = binary.Read(buf, binary.BigEndian, &value)
	return value
}

func (bt *boltDBStore) LastCommitTerm() int64 {
	return bt.lastCommitTerm
}

func (bt *boltDBStore) LastCommitIndex() int64 {
	return bt.lastCommitIndex
}

func (bt *boltDBStore) LastAppendTerm() int64 {
	return bt.lastAppendTerm
}

func (bt *boltDBStore) LastAppendIndex() int64 {
	return bt.lastAppendIndex
}

func (bt *boltDBStore) Append(entry *model.LogEntry) error {
	tag := cmdFactory.OfTag(entry.Tag)
	tag.Unmarshal(entry.Command)
	err, entryData := entry.Marshal()
	if err != nil {
		return err
	}
	return bt.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(unstableBucket)
		err = bucket.Put(int64ToBytes(entry.Index), entryData)
		if err != nil {
			return err
		}
		return nil
	})
}

func (bt *boltDBStore) Commit(index int64) error {
	return bt.db.Update(func(tx *bolt.Tx) error {
		ub := tx.Bucket(unstableBucket)
		key := int64ToBytes(index)
		data := ub.Get(key)
		if data == nil {
			return util.ErrorCommitLogNotFound
		}
		entry := new(model.LogEntry)
		err := entry.Unmarshal(data)
		if err != nil {
			return err
		}
		cmd := cmdFactory.OfTag(entry.Tag)
		err = cmd.Apply(tx)
		if err != nil {
			return err
		}

		cb := tx.Bucket(committedBucket)
		err = cb.Put(key, data)
		if err != nil {
			return err
		}

		// update committed.index committed.term
		mb := tx.Bucket(metaBucket)
		err = mb.Put(keyCommittedIndex, int64ToBytes(index))
		if err != nil {
			return err
		}
		err = mb.Put(keyCommittedTerm, int64ToBytes(entry.Term))
		if err != nil {
			return err
		}
		// del unstable.index
		err = ub.Delete(key)
		if err != nil {
			return err
		}
		bt.lastCommitIndex = entry.Index
		bt.lastCommitTerm = entry.Term
		return nil
	})
}

func (bt *boltDBStore) GetLog(index int64) (error, *model.LogEntry) {
	ret := []*model.LogEntry{
		nil,
	}
	err := bt.db.View(func(tx *bolt.Tx) error {
		key := int64ToBytes(index)
		value := tx.Bucket(committedBucket).Get(key)
		if value == nil {
			return nil
		}

		entry := new(model.LogEntry)
		err := entry.Unmarshal(value)
		if err != nil {
			return err
		}
		ret[0] = entry
		return nil
	})
	return err, ret[0]
}
