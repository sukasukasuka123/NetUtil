

# JWTUtil

`JWTUtil` 是一个基于 Go 的轻量级 JWT 工具库，用于简化 **令牌生成、解析、Cookie 存储与读取** 的开发流程。适用于用户认证、Session 维护、前后端分离登录系统等场景。

---

## 特性

* 封装 JWT 的生成与解析流程
* 支持自定义过期时间
* 自动写入 HttpOnly Cookie
* 从 Cookie 读取 JWT
* 基于 `HS256` 的对称加密
* 零依赖复杂结构，简单易集成到任意项目

---

## 安装

```bash
go get github.com/sukasukasuka123/NetUtil/jwtutil
```

在代码中引入：

```go
import "github.com/sukasukasuka123/NetUtil/jwtutil"
```

---

## 核心结构

`JWTManager` 提供完整的 Token 生成、解析、Cookie 读写功能：

```go
type JWTManager struct {
    Secret        []byte
    TokenDuration time.Duration
    CookieName    string
}
```

三个核心参数：

* **Secret**：签名密钥
* **TokenDuration**：令牌有效期
* **CookieName**：保存到浏览器时使用的 Cookie 名

---

## 使用示例

下面示例展示如何在登录时签发 Token、写入 Cookie，并在后续请求中读取和校验 Token。

### 生成 Token 并写入 Cookie

```go
package main

import (
	"net/http"
	"time"

	"github.com/sukasukasuka123/NetUtil/jwtutil"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	jm := jwtutil.NewJWTManager("my-secret-key", 24*time.Hour, "auth_token")

	claims := map[string]interface{}{
		"user_id": 123,
		"email":   "abc@test.com",
	}

	token, err := jm.GenerateToken(claims)
	if err != nil {
		http.Error(w, "failed to generate token", 500)
		return
	}

	// 写入 HttpOnly Cookie
	jm.SetTokenCookie(w, token)

	w.Write([]byte("login success"))
}
```

---

### 从 Cookie 中读取 Token 并解析

```go
func authHandler(w http.ResponseWriter, r *http.Request) {
	jm := jwtutil.NewJWTManager("my-secret-key", 24*time.Hour, "auth_token")

	tokenStr, err := jm.ReadTokenFromCookie(r)
	if err != nil {
		http.Error(w, "no token", 401)
		return
	}

	claims, err := jm.ParseToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", 401)
		return
	}

	w.Write([]byte("welcome user: " + claims["email"].(string)))
}
```

---

## 方法说明

### `GenerateToken(claims jwt.MapClaims)`

生成带自定义字段的 Token，会自动写入 `exp`（过期时间）。

### `ParseToken(tokenStr string)`

解析 Token，如果无效或过期会返回错误。

### `SetTokenCookie(w http.ResponseWriter, token string)`

将 Token 写入 HttpOnly Cookie，用于安全储存。

### `ReadTokenFromCookie(r *http.Request)`

从 Cookie 中读取 Token。

---

## 使用建议

* **Secret 需要足够复杂**，避免暴力破解
* 生产环境建议将 `Secure` 设置为 `true`（强制 HTTPS）
* Claims 可根据业务扩展，如 `user_id`、`roles`、`permissions`
* Cookie 模式适用于前后端分离但需要保持安全性的登录场景


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

