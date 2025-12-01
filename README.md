
# LimiterUtil

`LimiterUtil` 是一个基于 Go 的泛型限流库，集成了 **双/三桶限流**、**队列排队**、**熔断机制** 和 **失败重试控制**。适用于高并发场景下对任务执行速率和稳定性进行控制。

---

## 特性

* 双/三桶限流（稳定桶 + 突发桶）
* 支持任务队列（可配置最大长度和 TTL）
* 支持连续失败触发熔断（Reject）
* 泛型支持任意类型的 payload
* 异步回调处理任务
* Worker 自动拉取队列任务并执行

---

## 安装

```bash
go get github.com/sukasukasuka123/NetUtil/limiterUtil
```

在代码中引入：

```go
import "github.com/sukasukasuka123/NetUtil/limiterUtil"
```

---

## 核心概念

* **LimiterConfig**：限流配置
* **State**：提交任务后的状态，包含：

  * `StateTaken`：立即获得令牌执行
  * `StateQueued`：进入队列等待
  * `StateDiscarded`：任务被丢弃（队列满或熔断）
* **Limiter[T]**：泛型限流器，T 为 payload 类型
* **onProcess**：获取令牌后回调处理任务的函数

---

## 使用示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/sukasukasuka123/NetUtil/limiterUtil"
)

func main() {
	cfg := limiterUtil.LimiterConfig{
		StableCap:        5,
		StableRate:       1,
		BurstCap:         10,
		BurstRate:        5,
		FailThreshold:    3,
		RejectDur:        5 * time.Second,
		QueueMaxLen:      100,
		QueueCleanup:     1 * time.Second,
		WorkerInterval:   50 * time.Millisecond,
		QueueItemTTL:     30 * time.Second,
		TokenWaitTimeout: 10 * time.Second,
	}

	limiter := limiterUtil.NewLimiter[string](cfg)

	limiter.SetOnProcess(func(payload string) {
		fmt.Println("Processing:", payload)
	})

	limiter.Start()
	defer limiter.Stop()

	for i := 0; i < 20; i++ {
		state := limiter.Submit(fmt.Sprintf("task-%d", i))
		switch state {
		case limiterUtil.StateTaken:
			fmt.Println("Taken immediately:", i)
		case limiterUtil.StateQueued:
			fmt.Println("Queued:", i)
		case limiterUtil.StateDiscarded:
			fmt.Println("Discarded:", i)
		}
	}

	time.Sleep(5 * time.Second)
}
```

---

## 配置说明

| 配置项              | 说明                |
| ---------------- | ----------------- |
| StableCap        | 稳定桶容量，限制稳定速率下的并发量 |
| StableRate       | 稳定桶每秒填充速率         |
| BurstCap         | 突发桶容量，用于短时间高峰流量   |
| BurstRate        | 突发桶填充速率           |
| FailThreshold    | 连续失败次数超过该值触发熔断    |
| RejectDur        | 熔断冷却时间            |
| QueueMaxLen      | 队列最大长度，0 表示无限制    |
| QueueCleanup     | 队列后台清理周期          |
| WorkerInterval   | Worker 从队列获取任务的间隔 |
| QueueItemTTL     | 队列中任务最大等待时间       |
| TokenWaitTimeout | Worker 等待令牌的最大时间  |

---

## 使用建议

1. **高并发场景**：结合稳定桶和突发桶，防止瞬时请求打爆系统。
2. **队列处理**：设置合理的 `QueueMaxLen` 和 `QueueItemTTL`，防止队列堆积过多。
3. **熔断策略**：通过 `FailThreshold` 和 `RejectDur` 控制连续失败后的短暂拒绝，提高系统稳定性。
4. **泛型 payload**：可以使用任意类型，例如 `struct` 或基本类型。

