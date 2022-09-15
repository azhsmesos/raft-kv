package store

import "encoding/json"

type CmdBase struct {
}

func (cb *CmdBase) Marshal() []byte {
	data, err := json.Marshal(cb)
	if err != nil {
		return nil
	}
	return data
}

func (cb *CmdBase) Unmarshal(data []byte) {
	_ = json.Unmarshal(data, cb)
}