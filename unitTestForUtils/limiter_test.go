package unitTestForUtils

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sukasukasuka123/NetUtil/limiterUtil"
)

// ---------------------------
// 测试 TripleBucket 模拟
// ---------------------------
type TripleBucket struct{}

func NewTripleBucket(stableCap, stableRate, burstCap, burstRate float64, failThreshold int, rejectDur time.Duration) *TripleBucket {
	return &TripleBucket{}
}

// 简单模拟：总是能拿到 token
func (tb *TripleBucket) TryTake() bool {
	return true
}

// 简单模拟：永不 reject
func (tb *TripleBucket) IsRejected() bool {
	return false
}

// ---------------------------
// 单元测试
// ---------------------------
func TestLimiter(t *testing.T) {
	cfg := limiterUtil.LimiterConfig{
		StableCap:      5,
		StableRate:     3,
		BurstCap:       5,
		BurstRate:      5,
		FailThreshold:  3,
		RejectDur:      time.Second,
		QueueMaxLen:    10, //  队列足够大
		QueueCleanup:   500 * time.Millisecond,
		WorkerInterval: 50 * time.Millisecond, //  加快处理
		QueueItemTTL:   30 * time.Second,      //  足够的等待时间
	}

	limiter := limiterUtil.NewLimiter[int](cfg)

	var mu sync.Mutex
	done := make(chan struct{})
	count := 0
	var processed []int

	limiter.SetOnProcess(func(i int) {
		mu.Lock()
		defer mu.Unlock()
		fmt.Printf("[%s] 处理: %d\n", time.Now().Format("15:04:05.000"), i)
		processed = append(processed, i)
		count++
		if count == 10 {
			close(done)
		}
	})

	limiter.Start()
	defer limiter.Stop()

	// 提交 10 个请求
	for i := 1; i <= 10; i++ {
		state := limiter.Submit(i)
		fmt.Printf("提交 %d 状态: %v\n", i, state)
	}

	// 等待处理完成
	select {
	case <-done:
		fmt.Printf("\n O 单元测试完成，共处理 %d 个请求\n", len(processed))
		fmt.Printf("处理顺序: %v\n", processed)
	case <-time.After(10 * time.Second):
		mu.Lock()
		t.Fatalf("✗ 测试超时！只处理了 %d/%d 个请求: %v", len(processed), 10, processed)
		mu.Unlock()
	}
}
