package ratelimiter

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/kkonst40/chat-service/internal/config"
	"golang.org/x/time/rate"
)

const shardCount = 64

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type shard struct {
	mu  sync.Mutex
	ips map[string]*client
}

type IPRateLimiter struct {
	shards   []*shard
	limit    rate.Limit
	maxBurst int
}

func New(cfg *config.Config) *IPRateLimiter {
	limiter := &IPRateLimiter{
		shards:   make([]*shard, shardCount),
		limit:    rate.Limit(cfg.RateLimiter.Limit),
		maxBurst: cfg.RateLimiter.MaxBurst,
	}

	cleanupInterval := time.Duration(cfg.RateLimiter.CleanupIntervalSeconds) * time.Second
	ipIdleLifetime := time.Duration(cfg.RateLimiter.IPIdleLifetimeSeconds) * time.Second

	for i := range limiter.shards {
		limiter.shards[i] = &shard{ips: make(map[string]*client)}
		go limiter.cleanupVisitors(i, cleanupInterval, ipIdleLifetime)
	}

	return limiter
}

func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	shard := l.getShard(ip)

	v, exists := shard.ips[ip]
	if !exists {
		limiter := rate.NewLimiter(l.limit, l.maxBurst)
		shard.ips[ip] = &client{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (l *IPRateLimiter) cleanupVisitors(shardIndex int, d, ipIdleLifetime time.Duration) {
	shard := l.shards[shardIndex]
	ticker := time.NewTicker(d)
	for range ticker.C {
		shard.mu.Lock()
		for ip, v := range shard.ips {
			if time.Since(v.lastSeen) > ipIdleLifetime {
				delete(shard.ips, ip)
			}
		}
		shard.mu.Unlock()
	}
}

func (l *IPRateLimiter) getShard(ip string) *shard {
	h := fnv.New32a()
	h.Write([]byte(ip))
	return l.shards[h.Sum32()%shardCount]
}
