package breaker

import (
	"sync"
	"sync/atomic"
	"time"
)

// ========== 滑动窗口（时间轮） ==========

type bucket struct {
	accepts atomic.Int64 // 成功数
	total   atomic.Int64 // 总请求数
}

type rollingWindow struct {
	buckets    []bucket
	size       int
	interval   time.Duration // 每个桶的时间跨度
	lastUpdate int64         // 上次更新时间戳（毫秒）
	mu         sync.Mutex
}

func newRollingWindow(window time.Duration, buckets int) *rollingWindow {
	if buckets <= 0 {
		buckets = 40
	}
	return &rollingWindow{
		buckets:  make([]bucket, buckets),
		size:     buckets,
		interval: window / time.Duration(buckets),
	}
}

// add 写入一次请求结果
func (rw *rollingWindow) add(accept bool) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	now := time.Now().UnixMilli()
	idx, offset := rw.offset(now)
	if offset > 0 {
		rw.resetBuckets(idx, offset)
	}
	b := &rw.buckets[idx]
	b.total.Add(1)
	if accept {
		b.accepts.Add(1)
	}
	rw.lastUpdate = now
}

// history 返回当前窗口内的成功数与总数
func (rw *rollingWindow) history() (accepts, total int64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	now := time.Now().UnixMilli()
	idx, offset := rw.offset(now)
	if offset > 0 {
		rw.resetBuckets(idx, offset)
	}
	for i := range rw.buckets {
		accepts += rw.buckets[i].accepts.Load()
		total += rw.buckets[i].total.Load()
	}
	return
}

// reset 重置窗口
func (rw *rollingWindow) reset() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	for i := range rw.buckets {
		rw.buckets[i].accepts.Store(0)
		rw.buckets[i].total.Store(0)
	}
	rw.lastUpdate = 0
}

// offset 计算当前时间对应的桶下标，以及距离上次更新跨越的桶数
func (rw *rollingWindow) offset(now int64) (idx, offset int) {
	last := rw.lastUpdate
	if last == 0 {
		last = now
	}
	offset = int((now - last) / rw.interval.Milliseconds())
	if offset > rw.size {
		offset = rw.size
	}
	idx = int((now / rw.interval.Milliseconds()) % int64(rw.size))
	return
}

// resetBuckets 重置 lastUpdate 到 idx 之间的桶
func (rw *rollingWindow) resetBuckets(idx, offset int) {
	for i := 1; i <= offset; i++ {
		pos := (idx - i + rw.size) % rw.size
		rw.buckets[pos].accepts.Store(0)
		rw.buckets[pos].total.Store(0)
	}
}
