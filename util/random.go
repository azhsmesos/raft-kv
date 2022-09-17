package util

import (
	"math/rand"
	"time"
)

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func RandomizeInt64(value int64) int64 {
	return value + value*random.Int63n(30)/100
}

func RandomizeDuration(value time.Duration) time.Duration {
	i := int64(value)
	return time.Duration(RandomizeInt64(i))
}
