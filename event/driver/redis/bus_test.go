package redis

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func setupBus(t *testing.T) (*Bus, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	assert.NoError(t, err)
	t.Cleanup(mr.Close)

	port, err := strconv.Atoi(mr.Port())
	assert.NoError(t, err)

	bus, err := NewBus(&Config{Host: mr.Host(), Port: port})
	assert.NoError(t, err)
	t.Cleanup(func() { _ = bus.Shutdown(context.Background()) })

	return bus, mr
}

func TestNewBus_NilConfig(t *testing.T) {
	bus, err := NewBus(nil)
	assert.Nil(t, bus)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrConnectFailed.Error())
}

func TestNewBus_InvalidConfig(t *testing.T) {
	bus, err := NewBus(&Config{Host: "127.0.0.1", Port: 1})
	assert.Nil(t, bus)
	assert.Error(t, err)
}

func TestPublishSubscribe(t *testing.T) {
	bus, _ := setupBus(t)
	ctx := context.Background()

	got := make(chan string, 1)
	sub, err := bus.Subscribe("order.created", func(_ context.Context, payload []byte) error {
		got <- string(payload)
		return nil
	})
	assert.NoError(t, err)
	defer sub.Unsubscribe()

	assert.NoError(t, bus.Publish(ctx, "order.created", []byte(`{"id":1}`)))

	select {
	case payload := <-got:
		assert.Equal(t, `{"id":1}`, payload)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestUnsubscribe(t *testing.T) {
	bus, _ := setupBus(t)
	ctx := context.Background()

	got := make(chan string, 1)
	sub, err := bus.Subscribe("order.cancelled", func(_ context.Context, payload []byte) error {
		got <- string(payload)
		return nil
	})
	assert.NoError(t, err)

	sub.Unsubscribe()
	// 等待取消生效
	time.Sleep(50 * time.Millisecond)

	assert.NoError(t, bus.Publish(ctx, "order.cancelled", []byte("x")))

	select {
	case <-got:
		t.Fatal("unexpected message after unsubscribe")
	case <-time.After(300 * time.Millisecond):
	}
}

func TestSubscribeDuplicateTopic(t *testing.T) {
	bus, _ := setupBus(t)

	handler := func(_ context.Context, _ []byte) error { return nil }
	_, err := bus.Subscribe("order.updated", handler)
	assert.NoError(t, err)

	_, err = bus.Subscribe("order.updated", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrSubscribeFailed.Error())
}

func TestPublishAfterShutdown(t *testing.T) {
	bus, _ := setupBus(t)
	ctx := context.Background()

	assert.NoError(t, bus.Shutdown(ctx))
	assert.Error(t, bus.Publish(ctx, "order.created", []byte("x")))
}
