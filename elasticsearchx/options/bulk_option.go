package options

import (
	"context"
	"time"

	"github.com/LouYuanbo1/go-webservice/elasticsearchx/config"
)

type bulk struct {
	numWorkers    int
	flushBytes    int
	flushInterval time.Duration
	onError       func(context.Context, error)          // Called for indexer errors.
	onFlushStart  func(context.Context) context.Context // Called when the flush starts.
	onFlushEnd    func(context.Context)                 // Called when the flush ends.
	timeout       time.Duration
}

func (b *bulk) GetNumWorkers() int {
	return b.numWorkers
}
func (b *bulk) GetFlushBytes() int {
	return b.flushBytes
}

func (b *bulk) GetFlushInterval() time.Duration {
	return b.flushInterval
}

func (b *bulk) GetOnError() func(context.Context, error) {
	return b.onError
}

func (b *bulk) GetOnFlushStart() func(context.Context) context.Context {
	return b.onFlushStart
}

func (b *bulk) GetOnFlushEnd() func(context.Context) {
	return b.onFlushEnd
}

func (b *bulk) GetTimeout() time.Duration {
	return b.timeout
}

func NewBulk() *bulk {
	return &bulk{}
}

func NewBulkWithOptions(opts ...BulkOption) *bulk {
	b := NewBulk()
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func NewBulkByConfig(cfg config.BulkIndexerConfig) *bulk {
	return &bulk{
		numWorkers:    cfg.NumWorkers,
		flushBytes:    cfg.FlushBytes,
		flushInterval: cfg.FlushInterval,
		onError:       cfg.OnError,
		onFlushStart:  cfg.OnFlushStart,
		onFlushEnd:    cfg.OnFlushEnd,
		timeout:       cfg.Timeout,
	}
}

func (b *bulk) WithNumWorkers(numWorkers int) *bulk {
	b.numWorkers = numWorkers
	return b
}

func (b *bulk) WithFlushBytes(flushBytes int) *bulk {
	b.flushBytes = flushBytes
	return b
}

func (b *bulk) WithFlushInterval(flushInterval time.Duration) *bulk {
	b.flushInterval = flushInterval
	return b
}

func (b *bulk) WithOnError(onError func(context.Context, error)) *bulk {
	b.onError = onError
	return b
}

func (b *bulk) WithOnFlushStart(onFlushStart func(context.Context) context.Context) *bulk {
	b.onFlushStart = onFlushStart
	return b
}

func (b *bulk) WithOnFlushEnd(onFlushEnd func(context.Context)) *bulk {
	b.onFlushEnd = onFlushEnd
	return b
}

func (b *bulk) WithTimeout(timeout time.Duration) *bulk {
	b.timeout = timeout
	return b
}

type BulkOption func(*bulk)

func WithNumWorkers(numWorkers int) BulkOption {
	return func(b *bulk) {
		b.numWorkers = numWorkers
	}
}

func WithFlushBytes(flushBytes int) BulkOption {
	return func(b *bulk) {
		b.flushBytes = flushBytes
	}
}

func WithFlushInterval(flushInterval time.Duration) BulkOption {
	return func(b *bulk) {
		b.flushInterval = flushInterval
	}
}

func WithOnError(onError func(context.Context, error)) BulkOption {
	return func(b *bulk) {
		b.onError = onError
	}
}

func WithOnFlushStart(onFlushStart func(context.Context) context.Context) BulkOption {
	return func(b *bulk) {
		b.onFlushStart = onFlushStart
	}
}

func WithOnFlushEnd(onFlushEnd func(context.Context)) BulkOption {
	return func(b *bulk) {
		b.onFlushEnd = onFlushEnd
	}
}

func WithTimeout(timeout time.Duration) BulkOption {
	return func(b *bulk) {
		b.timeout = timeout
	}
}
