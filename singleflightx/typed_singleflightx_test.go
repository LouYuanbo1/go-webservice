package singleflightx

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTypedSingleFlight_Do(t *testing.T) {
	sf := NewTypedSingleFlight[string]()

	val, err := sf.Do("test", func() (string, error) {
		return "result", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", val)
}

func TestTypedSingleFlight_DoWithError(t *testing.T) {
	sf := NewTypedSingleFlight[string]()
	expectedErr := errors.New("test error")

	val, err := sf.Do("test", func() (string, error) {
		return "", expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", val)
}

func TestTypedSingleFlight_DoEx(t *testing.T) {
	sf := NewTypedSingleFlight[string]()

	val, fresh, err := sf.DoEx("test", func() (string, error) {
		return "result", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", val)
	assert.True(t, fresh)
}

func TestTypedSingleFlight_DoExWithError(t *testing.T) {
	sf := NewTypedSingleFlight[string]()
	expectedErr := errors.New("test error")

	val, fresh, err := sf.DoEx("test", func() (string, error) {
		return "", expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", val)
	assert.True(t, fresh)
}

func TestTypedSingleFlight_ConcurrentCalls(t *testing.T) {
	sf := NewTypedSingleFlight[string]()
	callCount := 0
	var mu sync.Mutex

	var wg sync.WaitGroup
	startCh := make(chan struct{})
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startCh
			val, err := sf.Do("concurrent_key", func() (string, error) {
				time.Sleep(10 * time.Millisecond)
				mu.Lock()
				callCount++
				mu.Unlock()
				return "concurrent_result", nil
			})
			assert.NoError(t, err)
			assert.Equal(t, "concurrent_result", val)
		}()
	}

	close(startCh)
	wg.Wait()

	assert.Equal(t, 1, callCount, "function should only be called once")
}

func TestTypedSingleFlight_ConcurrentCallsDoEx(t *testing.T) {
	sf := NewTypedSingleFlight[string]()
	callCount := 0
	freshCount := 0
	var mu sync.Mutex

	var wg sync.WaitGroup
	startCh := make(chan struct{})
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startCh
			val, fresh, err := sf.DoEx("concurrent_key_ex", func() (string, error) {
				time.Sleep(10 * time.Millisecond)
				mu.Lock()
				callCount++
				mu.Unlock()
				return "concurrent_result", nil
			})
			assert.NoError(t, err)
			assert.Equal(t, "concurrent_result", val)
			if fresh {
				mu.Lock()
				freshCount++
				mu.Unlock()
			}
		}()
	}

	close(startCh)
	wg.Wait()

	assert.Equal(t, 1, callCount, "function should only be called once")
	assert.Equal(t, 1, freshCount, "only one call should be fresh")
}

func TestTypedSingleFlight_DifferentKeys(t *testing.T) {
	sf := NewTypedSingleFlight[string]()
	callCount := 0
	var mu sync.Mutex

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		val, err := sf.Do("key1", func() (string, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			return "result1", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "result1", val)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		val, err := sf.Do("key2", func() (string, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			return "result2", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "result2", val)
	}()

	wg.Wait()

	assert.Equal(t, 2, callCount, "each key should trigger its own function call")
}

func TestTypedSingleFlight_IntType(t *testing.T) {
	sf := NewTypedSingleFlight[int]()

	val, err := sf.Do("int_test", func() (int, error) {
		return 42, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 42, val)
}
