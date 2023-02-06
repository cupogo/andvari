// Package idgen 全局型 ID 生成
package idgen

import (
	"sync/atomic"
	"time"
)

const (
	epoch     = int64(1451606400000) // 2016-01-01 00:00:00 +0000 UTC
	shardMask = int64(1<<10) - 1     // 1023
	seqMask   = int64(1<<11) - 1     // 2047

	Min = 1 << 21 // 2097152
)

// IDGen generates sortable unique int64 numbers that consist of:
// - 43 bits for time in milliseconds.
// - 10 bits that represent the shard id.
// - 11 bits that represent an auto-incrementing sequence.
//
// That means that we can generate 2048 ids per
// millisecond for 1024 shards.
type IDGen struct {
	seq   int64
	shard int64
}

// NewWithShard returns id generator for the shard.
func NewWithShard(shard int64) *IDGen {
	return &IDGen{
		shard: shard % (shardMask + 1),
	}
}

// NextWithTime returns increasing id for the time. Note that you can only
// generate 2048 unique numbers per millisecond.
func (g *IDGen) NextWithTime(tm time.Time) int64 {
	seq := atomic.AddInt64(&g.seq, 1) - 1
	id := tm.UnixNano()/int64(time.Millisecond) - epoch
	id <<= 21
	id |= g.shard << 11
	id |= seq % (seqMask + 1)
	return id
}

// Next acts like NextWithTime, but returns id for current time.
func (g *IDGen) Next() int64 {
	return g.NextWithTime(time.Now())
}

// SplitID splits id into time, shard id, and sequence id.
func SplitID(id int64) (tm time.Time, shardID int64, seqID int64) {
	ms := int64(id>>21) + epoch
	sec := ms / 1000
	tm = time.Unix(sec, (ms-sec*1000)*int64(time.Millisecond))
	shardID = (id >> 11) & shardMask
	seqID = id & seqMask
	return
}
