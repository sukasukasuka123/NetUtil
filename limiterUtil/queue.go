package limiterUtil

import (
	"context"
	"errors"
	"sync"
	"time"
)

// 泛型队列项
type queueItem[T any] struct {
	payload  T
	expireAt time.Time
}

// 泛型队列
type Queue[T any] struct {
	items           []queueItem[T]
	lock            sync.Mutex
	maxLen          int
	cleanupInterval time.Duration
	stopCh          chan struct{}
}

// 构造函数
func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{stopCh: make(chan struct{})}
	go q.cleanupLoop()
	return q
}

// 设置队列最大长度
func (q *Queue[T]) SetMaxLen(max int) {
	q.maxLen = max
}

// 设置队列清理周期
func (q *Queue[T]) SetCleanupInterval(d time.Duration) {
	q.cleanupInterval = d
}

// 入队操作（失败返回 error）
func (q *Queue[T]) Enqueue(payload T, ttl time.Duration) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.maxLen > 0 && len(q.items) >= q.maxLen {
		return errors.New("queue full")
	}

	q.items = append(q.items, queueItem[T]{payload: payload, expireAt: time.Now().Add(ttl)})
	return nil
}

// 出队操作（非阻塞，返回值 + 是否成功）
func (q *Queue[T]) Dequeue(ctx context.Context) (T, bool) {
	var zero T
	for {
		q.lock.Lock()
		if len(q.items) == 0 {
			q.lock.Unlock()
			return zero, false
		}

		item := q.items[0]
		if time.Now().After(item.expireAt) {
			// 丢弃过期项
			q.items = q.items[1:]
			q.lock.Unlock()
			continue
		}

		q.items = q.items[1:]
		q.lock.Unlock()
		return item.payload, true
	}
}

// 后台清理过期队列项
func (q *Queue[T]) cleanupLoop() {
	ticker := time.NewTicker(q.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			q.lock.Lock()
			filtered := q.items[:0]
			for _, item := range q.items {
				if item.expireAt.After(now) {
					filtered = append(filtered, item)
				}
			}
			q.items = filtered
			q.lock.Unlock()
		}
	}
}

// 停止队列后台清理
func (q *Queue[T]) Stop() {
	close(q.stopCh)
}
