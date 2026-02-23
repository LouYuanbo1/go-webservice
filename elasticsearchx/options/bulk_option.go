package options

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v9/esutil"
)

type Bulk struct {
	numWorkers    int
	flushBytes    int
	flushInterval time.Duration
	onError       func(context.Context, error)          // Called for indexer errors.
	onFlushStart  func(context.Context) context.Context // Called when the flush starts.
	onFlushEnd    func(context.Context)                 // Called when the flush ends.
	timeout       time.Duration
}

func NewBulk() *Bulk {
	return &Bulk{}
}

func (b *Bulk) WithNumWorkers(numWorkers int) *Bulk {
	b.numWorkers = numWorkers
	return b
}

func (b *Bulk) WithFlushBytes(flushBytes int) *Bulk {
	b.flushBytes = flushBytes
	return b
}

func (b *Bulk) WithFlushInterval(flushInterval time.Duration) *Bulk {
	b.flushInterval = flushInterval
	return b
}

func (b *Bulk) WithOnError(onError func(context.Context, error)) *Bulk {
	b.onError = onError
	return b
}

func (b *Bulk) WithOnFlushStart(onFlushStart func(context.Context) context.Context) *Bulk {
	b.onFlushStart = onFlushStart
	return b
}

func (b *Bulk) WithOnFlushEnd(onFlushEnd func(context.Context)) *Bulk {
	b.onFlushEnd = onFlushEnd
	return b
}

func (b *Bulk) WithTimeout(timeout time.Duration) *Bulk {
	b.timeout = timeout
	return b
}

func (b *Bulk) Build() *esutil.BulkIndexerConfig {
	bulkIndexerConfig := &esutil.BulkIndexerConfig{
		NumWorkers:    b.numWorkers,
		FlushBytes:    b.flushBytes,
		FlushInterval: b.flushInterval,
		OnError:       b.onError,
		OnFlushStart:  b.onFlushStart,
		OnFlushEnd:    b.onFlushEnd,
		Timeout:       b.timeout,
	}
	return bulkIndexerConfig
}

type BulkOption func(*Bulk)

func WithNumWorkers(numWorkers int) BulkOption {
	return func(b *Bulk) {
		b.numWorkers = numWorkers
	}
}

func WithFlushBytes(flushBytes int) BulkOption {
	return func(b *Bulk) {
		b.flushBytes = flushBytes
	}
}

func WithFlushInterval(flushInterval time.Duration) BulkOption {
	return func(b *Bulk) {
		b.flushInterval = flushInterval
	}
}

func WithOnError(onError func(context.Context, error)) BulkOption {
	return func(b *Bulk) {
		b.onError = onError
	}
}

func WithOnFlushStart(onFlushStart func(context.Context) context.Context) BulkOption {
	return func(b *Bulk) {
		b.onFlushStart = onFlushStart
	}
}

func WithOnFlushEnd(onFlushEnd func(context.Context)) BulkOption {
	return func(b *Bulk) {
		b.onFlushEnd = onFlushEnd
	}
}

func WithTimeout(timeout time.Duration) BulkOption {
	return func(b *Bulk) {
		b.timeout = timeout
	}
}

func NewBulkWithOptions(opts ...BulkOption) *Bulk {
	b := NewBulk()
	for _, opt := range opts {
		opt(b)
	}
	return b
}
