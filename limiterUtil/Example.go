package limiterUtil

import (
	"fmt"
	"time"
)

func ExampleUsage() {
	// create limiter for string payload
	cfg := LimiterConfig{
		StableCap:      5,   // 稳定桶最多 5 个 token
		StableRate:     1.0, // 每秒填 1 个
		BurstCap:       10,  // 突发桶最多 10 个
		BurstRate:      5.0, // 突发桶速率 5/s (只是补偿用)
		FailThreshold:  20,  // 连续失败阈值
		RejectDur:      5 * time.Second,
		QueueMaxLen:    100,
		QueueCleanup:   1 * time.Second,
		WorkerInterval: 200 * time.Millisecond,
		QueueItemTTL:   10 * time.Second,
	}
	lim := NewLimiter[string](cfg)
	lim.SetOnProcess(func(s string) {
		fmt.Println("processing:", s, "at", time.Now())
		// 模拟业务耗时
		time.Sleep(10 * time.Millisecond)
	})
	lim.Start()
	defer lim.Stop()

	// submit some requests
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("req-%02d", i)
		state := lim.Submit(id)
		switch state {
		case StateTaken:
			fmt.Println(id, "taken immediately")
		case StateQueued:
			fmt.Println(id, "queued")
		case StateDiscarded:
			fmt.Println(id, "discarded")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
