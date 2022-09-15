package store

import "github.com/boltdb/bolt"

type PutCmd struct {
	CmdBase
	Key   string
	Value []byte
}

func (pc *PutCmd) Apply(tx *bolt.Tx) error {
	bucket := tx.Bucket(dataBucket)
	return bucket.Put([]byte(pc.Key), pc.Value)
}
