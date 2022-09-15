package util

import "time"

const (
	HeartbeatInterval = 150 * time.Millisecond
	HeartbeatTimeout  = 5 * HeartbeatInterval
	ElectionTimeout   = HeartbeatTimeout
)
