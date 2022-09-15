package store

import "github.com/boltdb/bolt"

type DelCmd struct {
	CmdBase
	Key string
}

func (dc *DelCmd) Apply(tx *bolt.Tx) error {
	bucket := tx.Bucket(dataBucket)
	return bucket.Delete([]byte(dc.Key))
}
