package conntracker

import (
	"hash/fnv"
	"sync"
)

const shardCount = 64

type ConnTracker struct {
	shards   []*shard
	maxPerIP int
}

type shard struct {
	conns map[string]int
	mu    sync.Mutex
}

func New(maxPerIP int) *ConnTracker {
	ct := &ConnTracker{
		shards:   make([]*shard, shardCount),
		maxPerIP: maxPerIP,
	}
	for i := range ct.shards {
		ct.shards[i] = &shard{conns: make(map[string]int)}
	}
	return ct
}

func (t *ConnTracker) getShard(ip string) *shard {
	h := fnv.New32a()
	h.Write([]byte(ip))
	return t.shards[h.Sum32()%shardCount]
}

func (t *ConnTracker) Acquire(ip string) bool {
	shard := t.getShard(ip)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if shard.conns[ip] >= t.maxPerIP {
		return false
	}

	shard.conns[ip]++
	return true
}

func (t *ConnTracker) Release(ip string) {
	shard := t.getShard(ip)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.conns[ip]--
	if shard.conns[ip] <= 0 {
		delete(shard.conns, ip)
	}
}
