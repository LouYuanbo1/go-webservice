package limiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TokenLimiterMiddleware(limiter *TokenLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c.Request.Context()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return // 必须 return，否则会继续执行本函数的后续代码（虽然这里没有，但以防以后添加）
		}
		c.Next() // 放行给后续中间件和 handler
	}
}

func TestTokenLimiterMiddleware(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	limiter := NewTokenLimiter(10, 10, rdb, "test")
	middleware := TokenLimiterMiddleware(limiter)
	r := gin.Default()
	r.Use(middleware)
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	statusCodes := make([]int, 0, 11)

	for range 11 {
		wg.Go(func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			r.ServeHTTP(w, req)
			mu.Lock()
			statusCodes = append(statusCodes, w.Code)
			mu.Unlock()
		})
	}
	wg.Wait()

	// 统计状态码数量
	okCount := 0
	tooManyRequestsCount := 0
	for _, code := range statusCodes {
		switch code {
		case http.StatusOK:
			okCount++
		case http.StatusTooManyRequests:
			tooManyRequestsCount++
		}
	}

	// 11个请求，10个被允许，1个被拒绝
	assert.Equal(t, 10, okCount, "Expected 10 OK responses")
	assert.Equal(t, 1, tooManyRequestsCount, "Expected 1 TooManyRequests response")
}

func TestNewTokenLimiter(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	limiter := NewTokenLimiter(10, 10, rdb, "test")
	assert.NotNil(t, limiter, "NewTokenLimiter() should return a non-nil pointer")
	assert.Equal(t, 10, limiter.rate, "NewTokenLimiter() should set rate to 10")
	assert.Equal(t, 10, limiter.burst, "NewTokenLimiter() should set burst to 10")
}

func TestTokenLimiter_Allow(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// 创建限流器：每秒10个token，burst为10
	limiter := NewTokenLimiter(10, 10, rdb, "test_allow")

	// 测试初始状态：应该允许请求
	for range 10 {
		assert.True(t, limiter.Allow(context.Background()), "Allow() should return true for first 10 iterations")
	}

	// burst用完后应该拒绝请求
	assert.False(t, limiter.Allow(context.Background()), "Allow() should return false after burst is exhausted")
}

func TestTokenLimiter_AllowN(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// 创建限流器：每秒5个token，burst为10
	limiter := NewTokenLimiter(5, 10, rdb, "test_allowN")

	// 请求超过burst的数量应该被拒绝
	assert.False(t, limiter.AllowN(context.Background(), time.Now(), 11), "AllowN(11) should return false when burst is 10")

	// 请求正好burst数量应该允许
	assert.True(t, limiter.AllowN(context.Background(), time.Now(), 10), "AllowN(10) should return true when burst is 10")

	// 再次请求应该被拒绝
	assert.False(t, limiter.AllowN(context.Background(), time.Now(), 1), "AllowN(1) should return false after burst is exhausted")
}

func TestTokenLimiter_RescueMode(t *testing.T) {
	s := miniredis.RunT(t)
	// 不启动redis，模拟redis不可用的情况
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	limiter := NewTokenLimiter(1, 1, rdb, "test_rescue")

	// 由于redis不可用，应该使用本地限流器
	assert.True(t, limiter.Allow(context.Background()), "Allow() should return true in rescue mode")

	// 本地限流器的burst用完后应该拒绝（使用rate=1, burst=1避免token恢复）
	assert.False(t, limiter.Allow(context.Background()), "Allow() should return false after rescue limiter burst is exhausted")
}

func TestTokenLimiter_Concurrent(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// 创建限流器：每秒10个token，burst为10
	limiter := NewTokenLimiter(10, 10, rdb, "test_concurrent")

	var wg sync.WaitGroup
	allowed := 0
	var mu sync.Mutex

	// 启动20个并发请求
	for range 20 {
		wg.Go(func() {
			if limiter.Allow(context.Background()) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	// 应该只有10个请求被允许（burst限制）
	assert.Equal(t, 10, allowed, "Only 10 requests should be allowed (burst limit)")
}

func TestTokenLimiter_RateLimit(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// 创建限流器：每秒2个token，burst为2
	limiter := NewTokenLimiter(2, 2, rdb, "test_rate")

	// 第一次burst
	for range 2 {
		assert.True(t, limiter.Allow(context.Background()), "Allow() should return true at first 2 iterations")
	}

	// burst用完，应该拒绝请求
	assert.False(t, limiter.Allow(context.Background()), "Allow() should return false after burst is exhausted")

	// 等待1.1秒，确保跨过整秒边界，恢复2个token（rate=2，1秒恢复2个）
	time.Sleep(time.Millisecond * 1100)

	// 此时应该有2个token，可以取2个
	assert.True(t, limiter.Allow(context.Background()), "Allow() should return true after 1.1 seconds")
	assert.True(t, limiter.Allow(context.Background()), "Allow() should return true (second token)")

	// 再次请求应该被拒绝
	assert.False(t, limiter.Allow(context.Background()), "Allow() should return false after burst is exhausted")

	// 再等待1.1秒，恢复2个token
	time.Sleep(time.Millisecond * 1100)

	// 应该可以再次获取token
	assert.True(t, limiter.Allow(context.Background()), "Allow() should return true after another 1.1 seconds")
}
