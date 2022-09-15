package model

import "encoding/json"

type LogEntry struct {
	Tag       int
	Term      int64
	Index     int64
	PrevTerm  int64
	PrevIndex int64
	Command   []byte
}

func (le *LogEntry) Marshal() (error, []byte) {
	data, err := json.Marshal(le)
	if err != nil {
		return err, nil
	}
	return nil, data
}

func (le *LogEntry) Unmarshal(data []byte) error {
	return json.Unmarshal(data, le)
}
