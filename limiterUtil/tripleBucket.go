package limiterUtil

import (
	"sync"
	"time"
)

type TripleBucket struct {
	Stable *Bucket
	Burst  *Bucket

	// 熔断策略
	failCount     int           // 连续失败计数
	failThreshold int           // 超过这个阈值触发 reject
	rejectUntil   time.Time     // 拒绝截止时间（在此之前所有请求被直接拒绝）
	rejectDur     time.Duration // 冷却时长
	mu            sync.Mutex
}

func NewTripleBucket(stableCap, stableRate, burstCap, burstRate float64, failThreshold int, rejectDur time.Duration) *TripleBucket {
	return &TripleBucket{
		Stable:        NewBucket(stableCap, stableRate),
		Burst:         NewBucket(burstCap, burstRate),
		failThreshold: failThreshold,
		rejectDur:     rejectDur,
	}
}

// TryTake 尝试拿 token：先检查是否处于 reject 状态
func (t *TripleBucket) TryTake() bool {
	t.mu.Lock()
	now := time.Now()
	if !t.rejectUntil.IsZero() && now.Before(t.rejectUntil) {
		// 当前处于拒绝期
		t.mu.Unlock()
		return false
	}
	t.mu.Unlock()

	// 正常尝试：stable -> burst
	if t.Stable.TryTake(1.0) {
		// success: reset failCount
		t.mu.Lock()
		t.failCount = 0
		t.mu.Unlock()
		return true
	}
	if t.Burst.TryTake(1.0) {
		t.mu.Lock()
		t.failCount = 0
		t.mu.Unlock()
		return true
	}

	// 两桶都拿不到，算一次失败
	t.mu.Lock()
	t.failCount++
	if t.failCount >= t.failThreshold {
		t.rejectUntil = time.Now().Add(t.rejectDur)
		// reset failCount to avoid repeated accumulation
		t.failCount = 0
	}
	t.mu.Unlock()
	return false
}

// IsRejected 当前是否处于 reject 状态
func (t *TripleBucket) IsRejected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.rejectUntil.IsZero() && time.Now().Before(t.rejectUntil)
}

// ResetReject 手动重置 reject 状态
func (t *TripleBucket) ResetReject() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rejectUntil = time.Time{}
	t.failCount = 0
}
