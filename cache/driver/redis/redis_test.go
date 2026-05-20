package redis

import (
	"context"
	"strconv"
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
	client, err := cache.Open(driver)
	assert.NoError(t, err)
	return client
}

func prepareSampleData(t *testing.T, client *cache.Client) {
	t.Helper()

	ctx := context.Background()
	// 准备 300 条数据
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
	prepareSampleData(t, client)
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
		assert.Equal(t, nil, value)
	}
}

func TestGetRowCache(t *testing.T) {
	client := setupClient(t)
	rowcache := client.GetRawCache()
	rediscache, ok := rowcache.(cache.RedisCache)
	assert.True(t, ok)
	redisclient := rediscache.GetRedisClient()
	assert.NotNil(t, redisclient)
}
