package redis

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	ID   int
	Name string
}

func setupClient(t *testing.T) *cache.Client {
	t.Helper()

	// 启动 miniredis 模拟服务器
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	// 在测试结束时关闭 miniredis
	t.Cleanup(func() {
		mr.Close()
	})

	// 将端口字符串转换为整数
	port, err := strconv.Atoi(mr.Port())
	assert.NoError(t, err)

	// 使用 miniredis 的地址创建配置
	config := &Config{
		Host: mr.Host(),
		Port: port,
	}

	client, err := InitRedisClient(config)
	assert.NoError(t, err)

	cacher, err := NewRedisCache(client, singleflightx.NewSingleFlight())
	assert.NoError(t, err)
	return cache.NewClient(cacher)
}

func prepareSampleData(t *testing.T, client *cache.Client) {
	t.Helper()

	ctx := context.Background()
	for i := range 30 {
		switch i % 3 {
		case 0:
			key := "test" + strconv.Itoa(i)
			value := i
			err := client.Set(ctx, key, value, 10*time.Second)
			assert.NoError(t, err)
		case 1:
			key := "test" + strconv.Itoa(i)
			value := "testString" + strconv.Itoa(i)
			err := client.Set(ctx, key, value, 10*time.Second)
			assert.NoError(t, err)
		case 2:
			key := "test" + strconv.Itoa(i)
			value := &testStruct{
				ID:   i,
				Name: "testStruct" + strconv.Itoa(i),
			}
			err := client.Set(ctx, key, value, 10*time.Second)
			assert.NoError(t, err)
		}
	}
}

func TestSet(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	testCases := []struct {
		name  string
		key   string
		value any
		ttl   time.Duration
	}{
		{"int_value", "set_test_int", 42, 10 * time.Second},
		{"string_value", "set_test_string", "hello world", 10 * time.Second},
		{"struct_value", "set_test_struct", testStruct{ID: 1, Name: "test"}, 10 * time.Second},
		{"nil_value", "set_test_nil", nil, 10 * time.Second},
		{"empty_string", "set_test_empty", "", 10 * time.Second},
		{"special_chars", "test@#$%^&*()", "special", 10 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.Set(ctx, tc.key, tc.value, tc.ttl)
			assert.NoError(t, err)

			var result any
			if tc.value != nil {
				err := client.Get(ctx, tc.key, &result)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	client := setupClient(t)
	prepareSampleData(t, client)
	ctx := context.Background()

	for i := range 30 {
		key := "test" + strconv.Itoa(i)
		switch i % 3 {
		case 0:
			var value int
			err := client.Get(ctx, key, &value)
			assert.NoError(t, err)
			assert.Equal(t, i, value)
		case 1:
			var value string
			err := client.Get(ctx, key, &value)
			assert.NoError(t, err)
			assert.Equal(t, "testString"+strconv.Itoa(i), value)
		case 2:
			var value testStruct
			err := client.Get(ctx, key, &value)
			assert.NoError(t, err)
			assert.Equal(t, testStruct{
				ID:   i,
				Name: "testStruct" + strconv.Itoa(i),
			}, value)
		}
	}
}

func TestGetNotFound(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	var value any
	err := client.Get(ctx, "non_existent_key", &value)
	assert.Error(t, err)
	assert.Nil(t, value)
}

func TestGetEmptyKey(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	var value any
	err := client.Get(ctx, "", &value)
	assert.Error(t, err)
}

func TestDel(t *testing.T) {
	client := setupClient(t)
	prepareSampleData(t, client)
	ctx := context.Background()

	for i := range 30 {
		key := "test" + strconv.Itoa(i)
		err := client.Del(ctx, key)
		assert.NoError(t, err)
	}

	for i := range 30 {
		key := "test" + strconv.Itoa(i)
		var value any
		err := client.Get(ctx, key, &value)
		assert.Error(t, err)
		assert.Nil(t, value)
	}
}

func TestDelMultipleKeys(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	keys := []string{"del_multi_1", "del_multi_2", "del_multi_3"}
	for _, key := range keys {
		err := client.Set(ctx, key, "value", 10)
		assert.NoError(t, err)
	}

	err := client.Del(ctx, keys...)
	assert.NoError(t, err)

	for _, key := range keys {
		var value any
		err := client.Get(ctx, key, &value)
		assert.Error(t, err)
	}
}

func TestTTL(t *testing.T) {
	// 使用 miniredis 的时间控制功能
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	t.Cleanup(func() { mr.Close() })

	// 将端口字符串转换为整数
	port, err := strconv.Atoi(mr.Port())
	assert.NoError(t, err)

	config := &Config{
		Host: mr.Host(),
		Port: port,
	}

	redisClient, err := InitRedisClient(config)
	assert.NoError(t, err)

	cacher, err := NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	assert.NoError(t, err)
	client := cache.NewClient(cacher)

	ctx := context.Background()

	err = client.Set(ctx, "ttl_test", "ttl_value", 1*time.Second)
	assert.NoError(t, err)

	var value string
	err = client.Get(ctx, "ttl_test", &value)
	assert.NoError(t, err)
	assert.Equal(t, "ttl_value", value)

	// 使用 miniredis 的时间推进功能
	mr.FastForward(2 * time.Second)

	var expiredValue string
	err = client.Get(ctx, "ttl_test", &expiredValue)
	assert.Error(t, err)
}

func TestTake(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	called := 0
	query := func(val *string) error {
		called++
		*val = "queried_value"
		return nil
	}

	var value string
	err := client.Take(ctx, "take_test", &value, query, 10*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, "queried_value", value)
	assert.Equal(t, 1, called)

	var value2 string
	err = client.Take(ctx, "take_test", &value2, query, 10*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, "queried_value", value2)
	assert.Equal(t, 1, called)
}

func TestTakeWithError(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	expectedErr := errors.New("query error")
	query := func(val *string) error {
		return expectedErr
	}

	var value string
	err := client.Take(ctx, "take_error_test", &value, query, 10*time.Second)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTakeConcurrent(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	called := 0
	var mu sync.Mutex
	query := func(val *string) error {
		mu.Lock()
		called++
		mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		*val = "concurrent_value"
		return nil
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	results := make([]string, numGoroutines)

	for i := range numGoroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var value string
			err := client.Take(ctx, "take_concurrent_test", &value, query, 10*time.Second)
			assert.NoError(t, err)
			results[idx] = value
		}(i)
	}

	wg.Wait()

	for _, result := range results {
		assert.Equal(t, "concurrent_value", result)
	}

	mu.Lock()
	assert.Equal(t, 1, called, "query should only be called once due to singleflight")
	mu.Unlock()
}

func TestCacheOverflow(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	for i := range 100 {
		key := "overflow_" + strconv.Itoa(i)
		value := make([]byte, 100)
		err := client.Set(ctx, key, value, 100*time.Second)
		assert.NoError(t, err)
	}

	var value []byte
	err := client.Get(ctx, "overflow_0", &value)
	assert.NoError(t, err)
}

func TestNilConfig(t *testing.T) {
	config := (*Config)(nil)
	client, err := InitRedisClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestInvalidUnmarshal(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	err := client.Set(ctx, "invalid_json", "not a valid json", 10*time.Second)
	assert.NoError(t, err)

	var value testStruct
	err = client.Get(ctx, "invalid_json", &value)
	assert.Error(t, err)
}

func TestGetRawCache(t *testing.T) {
	client := setupClient(t)
	rowcache := client.GetRawCache()
	rediscache, ok := rowcache.(cache.RedisCache)
	assert.True(t, ok)
	redisClient := rediscache.GetRedisClient()
	assert.NotNil(t, redisClient)
}

func TestMiniredisDirect(t *testing.T) {
	// 直接测试 miniredis 功能
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	// 测试一些 Redis 命令
	mr.Set("foo", "bar")
	val, err := mr.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", val)

	mr.HSet("myhash", "field1", "value1")
	hval := mr.HGet("myhash", "field1")
	assert.Equal(t, "value1", hval)
}

func TestAcceptable(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil_error", nil, true},
		{"redis_nil", redis.Nil, true},
		{"context_canceled", context.Canceled, true},
		{"other_error", errors.New("other error"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := acceptable(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNewBreakerHook(t *testing.T) {
	brk := breaker.NewBreaker()
	hook := NewBreakerHook(brk)
	assert.NotNil(t, hook)
}

func TestDialHook(t *testing.T) {
	brk := breaker.NewBreaker()
	hook := NewBreakerHook(brk)

	next := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, nil
	}

	result := hook.DialHook(next)
	assert.NotNil(t, result)
}

func TestProcessHook(t *testing.T) {
	brk := breaker.NewBreaker()
	hook := NewBreakerHook(brk)

	t.Run("normal_command", func(t *testing.T) {
		nextCalled := false
		next := func(ctx context.Context, cmd redis.Cmder) error {
			nextCalled = true
			return nil
		}

		result := hook.ProcessHook(next)
		assert.NotNil(t, result)

		ctx := context.Background()
		cmd := redis.NewStringCmd(ctx, "GET", "test")
		err := result(ctx, cmd)
		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("ignored_command_blpop", func(t *testing.T) {
		nextCalled := false
		next := func(ctx context.Context, cmd redis.Cmder) error {
			nextCalled = true
			return nil
		}

		result := hook.ProcessHook(next)
		assert.NotNil(t, result)

		ctx := context.Background()
		cmd := redis.NewSliceCmd(ctx, "BLPOP", "test", "1")
		err := result(ctx, cmd)
		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("command_with_error", func(t *testing.T) {
		expectedErr := errors.New("redis error")
		next := func(ctx context.Context, cmd redis.Cmder) error {
			return expectedErr
		}

		result := hook.ProcessHook(next)
		assert.NotNil(t, result)

		ctx := context.Background()
		cmd := redis.NewStringCmd(ctx, "GET", "test")
		err := result(ctx, cmd)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestProcessPipelineHook(t *testing.T) {
	brk := breaker.NewBreaker()
	hook := NewBreakerHook(brk)

	t.Run("pipeline_success", func(t *testing.T) {
		nextCalled := false
		next := func(ctx context.Context, cmds []redis.Cmder) error {
			nextCalled = true
			return nil
		}

		result := hook.ProcessPipelineHook(next)
		assert.NotNil(t, result)

		ctx := context.Background()
		cmds := []redis.Cmder{
			redis.NewStringCmd(ctx, "GET", "test1"),
			redis.NewStringCmd(ctx, "GET", "test2"),
		}
		err := result(ctx, cmds)
		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("pipeline_with_error", func(t *testing.T) {
		expectedErr := errors.New("pipeline error")
		next := func(ctx context.Context, cmds []redis.Cmder) error {
			return expectedErr
		}

		result := hook.ProcessPipelineHook(next)
		assert.NotNil(t, result)

		ctx := context.Background()
		cmds := []redis.Cmder{
			redis.NewStringCmd(ctx, "GET", "test1"),
		}
		err := result(ctx, cmds)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
