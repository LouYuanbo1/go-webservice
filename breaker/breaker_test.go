package breaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBreaker(t *testing.T) {
	b := NewBreaker()
	assert.NotNil(t, b)

	b2 := NewBreaker(
		WithName("test"),
		WithK(2.0),
		WithProtection(10),
		WithWindow(5*time.Second),
		WithBuckets(20),
	)
	assert.NotNil(t, b2)
}

func TestDo_Success(t *testing.T) {
	b := NewBreaker()

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)

	total, accepts, rate := b.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

func TestDo_Failure(t *testing.T) {
	b := NewBreaker()

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return errors.New("test error")
	})
	assert.Error(t, err)

	total, accepts, rate := b.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(0), accepts)
	assert.Equal(t, 0.0, rate)
}

func TestDo_WithTimeout(t *testing.T) {
	b := NewBreaker()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := b.Do(ctx, func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestDoWithAcceptable(t *testing.T) {
	b := NewBreaker()

	err := b.DoWithAcceptable(context.Background(),
		func(ctx context.Context) error {
			return errors.New("expected error")
		},
		func(err error) bool {
			return err != nil && err.Error() == "expected error"
		})
	assert.Error(t, err)

	total, accepts, rate := b.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

func TestDoWithFallback(t *testing.T) {
	rejectCount := 0
	b := NewBreaker(
		WithProtection(0),
		WithK(0),
		WithRejectCallback(func() { rejectCount++ }),
	)

	err := b.DoWithFallback(context.Background(),
		func(ctx context.Context) error {
			return errors.New("service error")
		},
		func(err error) error {
			return errors.New("fallback result")
		})
	assert.Error(t, err)
	assert.Equal(t, "fallback result", err.Error())
}

func TestDoWithFallback_NoFallback(t *testing.T) {
	b := NewBreaker(WithProtection(0), WithK(0))

	err := b.DoWithFallback(context.Background(),
		func(ctx context.Context) error {
			return errors.New("service error")
		},
		nil)
	assert.Error(t, err)
	assert.Equal(t, "service error", err.Error())
}

func TestBreaker_Trips(t *testing.T) {
	b := NewBreaker(WithProtection(1), WithK(0.1), WithWindow(time.Hour))

	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			b.Do(context.Background(), func(ctx context.Context) error {
				return errors.New("failure")
			})
		})
	}
	wg.Wait()

	total, accepts, _ := b.GetMetrics()
	assert.GreaterOrEqual(t, total, int64(1))
	assert.Equal(t, int64(0), accepts)
}

func TestReset(t *testing.T) {
	b := NewBreaker()

	b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})
	b.Do(context.Background(), func(ctx context.Context) error {
		return errors.New("error")
	})

	total, _, _ := b.GetMetrics()
	assert.Equal(t, int64(2), total)

	b.Reset()

	total, accepts, rate := b.GetMetrics()
	assert.Equal(t, int64(0), total)
	assert.Equal(t, int64(0), accepts)
	assert.Equal(t, 0.0, rate)
}

func TestRejectCallback(t *testing.T) {
	rejectCount := 0
	b := NewBreaker(
		WithProtection(0),
		WithK(0),
		WithRejectCallback(func() { rejectCount++ }),
	)

	for range 100 {
		b.Do(context.Background(), func(ctx context.Context) error {
			return errors.New("failure")
		})
	}

	assert.Greater(t, rejectCount, 0)
}

func TestConcurrency(t *testing.T) {
	b := NewBreaker(WithWindow(time.Hour), WithProtection(1000))

	var wg sync.WaitGroup
	for i := range 1000 {
		wg.Add(1)
		go func(success bool) {
			defer wg.Done()
			b.Do(context.Background(), func(ctx context.Context) error {
				if success {
					return nil
				}
				return errors.New("error")
			})
		}(i%2 == 0)
	}
	wg.Wait()

	total, accepts, _ := b.GetMetrics()
	assert.Equal(t, int64(1000), total)
	assert.Equal(t, int64(500), accepts)
}

func TestDo_Canceled(t *testing.T) {
	b := NewBreaker()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := b.Do(ctx, func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestDoWithAcceptable_CustomSuccess(t *testing.T) {
	b := NewBreaker()

	err := b.DoWithAcceptable(context.Background(),
		func(ctx context.Context) error {
			return errors.New("custom error")
		},
		func(err error) bool {
			return err != nil && err.Error() == "custom error"
		})
	assert.Error(t, err)

	total, accepts, rate := b.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

func TestDo_FallbackCalledOnReject(t *testing.T) {
	b := NewBreaker(WithProtection(0), WithK(0))

	fallbackCalled := false
	err := b.DoWithFallback(context.Background(),
		func(ctx context.Context) error {
			return errors.New("should not be called")
		},
		func(err error) error {
			fallbackCalled = true
			return nil
		})

	assert.NoError(t, err)
	assert.True(t, fallbackCalled)
}

func TestMetrics(t *testing.T) {
	b := NewBreaker()

	for i := range 10 {
		if i%2 == 0 {
			b.Do(context.Background(), func(ctx context.Context) error {
				return nil
			})
		} else {
			b.Do(context.Background(), func(ctx context.Context) error {
				return errors.New("error")
			})
		}
	}

	total, accepts, rate := b.GetMetrics()
	assert.Equal(t, int64(10), total)
	assert.Equal(t, int64(5), accepts)
	assert.Equal(t, 0.5, rate)
}
