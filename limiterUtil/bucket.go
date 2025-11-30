package limiterUtil

import (
	"sync"
	"time"
)

type Bucket struct {
	capacity   float64   // 最大令牌数
	tokens     float64   // 当前令牌数（可为小数，支持浮点速率）
	rate       float64   // 每秒生成令牌数（可以是小数）
	lastUpdate time.Time // 上次更新时间
	mu         sync.Mutex
}

func NewBucket(capacity, rate float64) *Bucket {
	if capacity < 0 {
		capacity = 0
	}
	if rate < 0 {
		rate = 0
	}
	return &Bucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastUpdate: time.Now(),
	}
}

// refillLocked 假设已经持有 mu
func (b *Bucket) refillLocked(now time.Time) {
	if b.rate <= 0 {
		b.lastUpdate = now
		return
	}
	delta := now.Sub(b.lastUpdate).Seconds() * b.rate
	if delta <= 0 {
		return
	}
	b.tokens += delta
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastUpdate = now
}

// Refill 手动补令牌（线程安全）
func (b *Bucket) Refill() {
	b.mu.Lock()
	now := time.Now()
	b.refillLocked(now)
	b.mu.Unlock()
}

// TryTake 尝试拿 count 个令牌（count >= 1）
// 返回是否成功（拿到）
// 线程安全
func (b *Bucket) TryTake(count float64) bool {
	if count <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.refillLocked(now)
	if b.tokens >= count {
		b.tokens -= count
		return true
	}
	return false
}

// TakeOne 便捷：拿1个
func (b *Bucket) TakeOne() bool {
	return b.TryTake(1.0)
}

// Tokens 当前令牌（线程安全）
func (b *Bucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.refillLocked(now)
	return b.tokens
}

// Capacity 返回桶容量
func (b *Bucket) Capacity() float64 {
	return b.capacity
}

// SetRate 修改生成速率（线程安全）
func (b *Bucket) SetRate(rate float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refillLocked(time.Now())
	b.rate = rate
}

// StartAutoRefill 启动后台周期 refill（可选，桶本身为惰性 refill，不强制需要该协程）
func (b *Bucket) StartAutoRefill(interval time.Duration, stop <-chan struct{}) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				b.Refill()
			case <-stop:
				return
			}
		}
	}()
}
