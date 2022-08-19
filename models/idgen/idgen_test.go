package idgen

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	tm := time.Date(2018, time.January, 01, 0, 0, 0, 0, time.UTC)
	g := NewWithShard(2)
	prev := int64(-1)
	for i := 0; i < 37; i++ {
		next := g.NextWithTime(tm)
		if next <= prev {
			t.Errorf("%s: next=%d, prev=%d", tm, next, prev)
		} else {
			t.Logf("%20d \t %20x \t %14s \t%s", next, next, strconv.FormatInt(next, 36), tm)
		}
		prev = next
		tm = tm.AddDate(0, 1, 0)
	}
}

func TestSequence(t *testing.T) {
	g := NewWithShard(0)
	tm := time.Now()

	var prev int64
	for i := 0; i < 2048; i++ {
		next := g.NextWithTime(tm)
		if next <= prev {
			t.Errorf("iter %d: next=%d, prev=%d", i, next, prev)
		}
		prev = next
	}
}

func TestIdGen(t *testing.T) {
	N := 50000
	tm := time.Now()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	gen1 := NewWithShard(1)
	v1 := make([]int64, N)
	go func() {
		for i := 0; i < N; i++ {
			v1[i] = gen1.NextWithTime(tm)
		}
		wg.Done()
	}()

	gen2 := NewWithShard(2)
	v2 := make([]int64, N)
	go func() {
		for i := 0; i < N; i++ {
			v2[i] = gen2.NextWithTime(tm)
		}
		wg.Done()
	}()

	wg.Wait()

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if v1[i] == v2[j] {
				t.Fatalf("same numbers: %d and %d", v1[i], v2[j])
			}
		}
	}
}

func TestSplit(t *testing.T) {
	tm := time.Now()
	for shard := int64(0); shard < 1024; shard++ {
		gen := NewWithShard(shard)
		id := gen.NextWithTime(tm)
		gotTm, gotShard, gotSeq := SplitID(id)
		if gotTm.Unix() != tm.Unix() {
			t.Errorf("got %s, expected %s", gotTm, tm)
		}
		if gotShard != shard {
			t.Errorf("got %d, expected %d", gotShard, shard)
		}
		if gotSeq != 0 {
			t.Errorf("got %d, expected 1", gotSeq)
		}
	}
}
