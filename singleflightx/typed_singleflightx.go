package singleflightx

import "sync"

type TypedSingleFlight[T any] interface {
	Do(key string, fn func() (T, error)) (T, error)
	DoEx(key string, fn func() (T, error)) (val T, fresh bool, err error)
}

type typedCall[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

type typedFlightGroup[T any] struct {
	calls map[string]*typedCall[T]
	lock  sync.Mutex
}

// NewSingleFlight returns a SingleFlight.
func NewTypedSingleFlight[T any]() TypedSingleFlight[T] {
	return &typedFlightGroup[T]{
		calls: make(map[string]*typedCall[T]),
	}
}

func (tfg *typedFlightGroup[T]) Do(key string, fn func() (T, error)) (T, error) {
	c, done := tfg.createCall(key)
	if done {
		return c.val, c.err
	}

	tfg.makeCall(c, key, fn)
	return c.val, c.err
}

func (tfg *typedFlightGroup[T]) DoEx(key string, fn func() (T, error)) (val T, fresh bool, err error) {
	c, done := tfg.createCall(key)
	if done {
		return c.val, false, c.err
	}

	tfg.makeCall(c, key, fn)
	return c.val, true, c.err
}

func (tfg *typedFlightGroup[T]) createCall(key string) (c *typedCall[T], done bool) {
	tfg.lock.Lock()
	if c, ok := tfg.calls[key]; ok {
		tfg.lock.Unlock()
		c.wg.Wait()
		return c, true
	}

	c = new(typedCall[T])
	c.wg.Add(1)
	tfg.calls[key] = c
	tfg.lock.Unlock()

	return c, false
}

func (tfg *typedFlightGroup[T]) makeCall(c *typedCall[T], key string, fn func() (T, error)) {
	defer func() {
		tfg.lock.Lock()
		delete(tfg.calls, key)
		tfg.lock.Unlock()
		c.wg.Done()
	}()

	c.val, c.err = fn()
}
