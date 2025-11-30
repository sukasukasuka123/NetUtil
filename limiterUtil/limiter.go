package limiterUtil

import (
	"context"
	"sync"
	"time"
)

// ---------------------------
// 状态定义
// ---------------------------
type State int

const (
	StateTaken     State = iota // 直接拿到 token，立即执行业务
	StateQueued                 // 进入队列等待
	StateDiscarded              // 被丢弃（队列满或超时或 reject）
)

// ---------------------------
// 限流配置
// ---------------------------
type LimiterConfig struct {
	StableCap        float64       // 稳定桶容量
	StableRate       float64       // 稳定桶每秒速率
	BurstCap         float64       // 突发桶容量
	BurstRate        float64       // 突发桶速率
	FailThreshold    int           // 连续失败超过该值触发 reject
	RejectDur        time.Duration // 熔断冷却时间
	QueueMaxLen      int           // 队列最大长度（0 表示无上限）
	QueueCleanup     time.Duration // 队列清理周期
	WorkerInterval   time.Duration // worker 尝试从队列拉取的间隔
	QueueItemTTL     time.Duration // 队列项在队列中的最大等待时间
	TokenWaitTimeout time.Duration // worker 等待令牌的最大时间（独立于队列TTL）
}

// ---------------------------
// Limiter 泛型结构体
// ---------------------------
type Limiter[T any] struct {
	triple    *TripleBucket  // 双桶限流器
	queue     *Queue[T]      // 队列
	cfg       LimiterConfig  // 配置
	onProcess func(T)        // 拿到 token 后的回调
	stopCh    chan struct{}  // worker 停止信号
	wg        sync.WaitGroup // 等待 worker 停止
}

// ---------------------------
// 构造函数
// ---------------------------
func NewLimiter[T any](cfg LimiterConfig) *Limiter[T] {
	// 设置默认值
	if cfg.TokenWaitTimeout == 0 {
		cfg.TokenWaitTimeout = 30 * time.Second
	}

	tb := NewTripleBucket(cfg.StableCap, cfg.StableRate, cfg.BurstCap, cfg.BurstRate, cfg.FailThreshold, cfg.RejectDur)

	q := NewQueue[T]()
	q.SetMaxLen(cfg.QueueMaxLen)
	q.SetCleanupInterval(cfg.QueueCleanup)

	return &Limiter[T]{
		triple:    tb,
		queue:     q,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
		onProcess: nil,
	}
}

// ---------------------------
// 设置处理回调
// ---------------------------
func (l *Limiter[T]) SetOnProcess(fn func(T)) {
	l.onProcess = fn
}

// ---------------------------
// 启动 worker
// ---------------------------
func (l *Limiter[T]) Start() {
	l.wg.Add(1)
	go l.workerLoop()
}

// ---------------------------
// 停止 worker
// ---------------------------
func (l *Limiter[T]) Stop() {
	close(l.stopCh)
	l.queue.Stop() // 停止队列后台清理
	l.wg.Wait()
}

// ---------------------------
// 提交 payload
// ---------------------------
func (l *Limiter[T]) Submit(payload T) State {
	if l.triple.IsRejected() {
		return StateDiscarded
	}

	if l.triple.TryTake() {
		if l.onProcess != nil {
			go l.onProcess(payload)
		}
		return StateTaken
	}

	if err := l.queue.Enqueue(payload, l.cfg.QueueItemTTL); err != nil {
		return StateDiscarded
	}
	return StateQueued
}

// ---------------------------
// worker 循环（修复版）
// ---------------------------
func (l *Limiter[T]) workerLoop() {
	defer l.wg.Done()

	// 使用更小的检查间隔，提高响应速度
	checkInterval := l.cfg.WorkerInterval
	if checkInterval > 50*time.Millisecond {
		checkInterval = 50 * time.Millisecond
	}

	for {
		select {
		case <-l.stopCh:
			return
		default:
		}

		// 检查熔断状态
		if l.triple.IsRejected() {
			time.Sleep(checkInterval)
			continue
		}

		// 尝试从队列获取 item
		ctx := context.Background()
		item, ok := l.queue.Dequeue(ctx)
		if !ok {
			// 队列为空，短暂等待后重试
			select {
			case <-l.stopCh:
				return
			case <-time.After(checkInterval):
				continue
			}
		}

		// 【关键修复】持续尝试获取令牌，直到成功或被停止
		// 不设置超时，让队列自己的TTL机制来控制过期

		for {
			// 检查停止信号（高优先级）
			select {
			case <-l.stopCh:
				return
			default:
			}

			// 检查熔断
			if l.triple.IsRejected() {
				// 熔断期间等待一下再重试，避免忙等待
				select {
				case <-l.stopCh:
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			// 尝试获取令牌
			if l.triple.TryTake() {
				if l.onProcess != nil {
					go l.onProcess(item)
				}
				break
			}

			// 使用更短的重试间隔提高吞吐量
			select {
			case <-l.stopCh:
				return
			case <-time.After(10 * time.Millisecond):
				// 继续重试
			}
		}

		// 如果worker被停止才会走到这里且processed=false
		// 正常情况下processed必定为true
	}
}
