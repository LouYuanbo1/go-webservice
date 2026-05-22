package local

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	ID   int
	Name string
}

func setupClient(t *testing.T) *cache.Client {
	var config *Config
	driver := NewDriver(config, singleflightx.NewSingleFlight())
	cacher, err := cache.Open(driver)
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
			value := testStruct{
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
	client := setupClient(t)
	ctx := context.Background()

	err := client.Set(ctx, "ttl_test", "ttl_value", 1)
	assert.NoError(t, err)

	var value string
	err = client.Get(ctx, "ttl_test", &value)
	assert.NoError(t, err)
	assert.Equal(t, "ttl_value", value)

	time.Sleep(2 * time.Second)

	var expiredValue string
	err = client.Get(ctx, "ttl_test", &expiredValue)
	assert.Error(t, err)
}

func TestTake(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	called := 0
	query := func(val any) error {
		called++
		s, ok := val.(*string)
		if ok {
			*s = "queried_value"
		}
		return nil
	}

	var value string
	err := client.Take(ctx, "take_test", &value, query, 10)
	assert.NoError(t, err)
	assert.Equal(t, "queried_value", value)
	assert.Equal(t, 1, called)

	var value2 string
	err = client.Take(ctx, "take_test", &value2, query, 10)
	assert.NoError(t, err)
	assert.Equal(t, "queried_value", value2)
	assert.Equal(t, 1, called)
}

func TestTakeWithError(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	expectedErr := errors.New("query error")
	query := func(val any) error {
		return expectedErr
	}

	var value string
	err := client.Take(ctx, "take_error_test", &value, query, 10)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTakeConcurrent(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	called := 0
	var mu sync.Mutex
	query := func(val any) error {
		mu.Lock()
		called++
		mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		s, ok := val.(*string)
		if ok {
			*s = "concurrent_value"
		}
		return nil
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	results := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var value string
			err := client.Take(ctx, "take_concurrent_test", &value, query, 10)
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

	for i := 0; i < 100; i++ {
		key := "overflow_" + strconv.Itoa(i)
		value := make([]byte, 100)
		err := client.Set(ctx, key, value, 100)
		assert.NoError(t, err)
	}

	var value []byte
	err := client.Get(ctx, "overflow_0", &value)
	assert.NoError(t, err)
}

func TestNilConfig(t *testing.T) {
	config := (*Config)(nil)
	driver := NewDriver(config, singleflightx.NewSingleFlight())
	assert.NotNil(t, driver)

	cacher, err := cache.Open(driver)
	assert.NoError(t, err)
	assert.NotNil(t, cacher)
}

func TestInvalidUnmarshal(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	err := client.Set(ctx, "invalid_json", "not a valid json", 10)
	assert.NoError(t, err)

	var value testStruct
	err = client.Get(ctx, "invalid_json", &value)
	assert.Error(t, err)
}

func TestGetRawCache(t *testing.T) {
	client := setupClient(t)
	rowcache := client.GetRawCache()
	localcache, ok := rowcache.(cache.LocalCache)
	assert.True(t, ok)
	localclient := localcache.GetLocalCache()
	assert.NotNil(t, localclient)
}

func TestDriverName(t *testing.T) {
	driver := NewDriver(nil, singleflightx.NewSingleFlight())
	assert.Equal(t, "local", driver.Name())
}
