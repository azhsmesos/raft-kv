package store

import "github.com/boltdb/bolt"

// ICmd 操作指令接口
type ICmd interface {
	Marshal() []byte
	Unmarshal(data []byte)
	Apply(tx *bolt.Tx) error
}
