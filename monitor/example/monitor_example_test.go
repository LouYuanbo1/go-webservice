package example

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-burn"
	"github.com/LouYuanbo1/go-webservice/limiter"
	"github.com/LouYuanbo1/go-webservice/monitor"
	"github.com/LouYuanbo1/go-webservice/monitor/parser"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var ginRouter *gin.Engine

func TokenLimiterMiddleware(limiter *limiter.TokenLimiter) gin.HandlerFunc {
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

func GinMetricsMiddleware(mw *monitor.MetricsMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		path := c.Request.URL.Path
		status := c.Writer.Status()

		mw.Record(path, status, duration)
	}
}

func TestMain(m *testing.M) {
	// ---- 全局 setup ----
	// 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	// 创建限流器
	limit := limiter.NewTokenLimiter(1, 2, redisClient, "rate-limiter")
	// 创建监控器
	res := prometheus.DefaultRegisterer

	monitor, err := monitor.NewMetricsMiddleware(monitor.MetricsConfig{
		Namespace: "test-breaker",
	}, res)
	if err != nil {
		panic(err)
	}
	// 创建 Gin 路由
	ginRouter = gin.Default()
	// 不经过限流器的接口
	ginRouter.GET("/api_no_limiter", GinMetricsMiddleware(monitor), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})
	ginRouter.GET("/api_with_limiter", TokenLimiterMiddleware(limit), GinMetricsMiddleware(monitor), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})
	ginRouter.GET("/metrics", gin.WrapH(promhttp.HandlerFor(res.(prometheus.Gatherer), promhttp.HandlerOpts{})))
	//创建分析器
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				ginRouter.ServeHTTP(w, req)
				families, err := parser.ParseFromReader(w.Body)
				if err != nil {
					panic(err)
				}
				for _, family := range families {
					fmt.Println(family.String())
				}
			}
		}
	}(ctx)

	// 运行所有测试
	code := m.Run()
	// ---- 全局 teardown ----
	if redisClient != nil {
		_ = redisClient.Close()
	}
	cancel()
	os.Exit(code)
}

// 抽离公共压测逻辑
func runBurnTest(t *testing.T, path string) {
	tester, ctx := burn.NewTester(context.Background())
	steps := []burn.StepConfig{
		{Concurrency: 2, Duration: 5 * time.Second, RateLimit: 2},
	}

	taskFn := func(ctx context.Context, reqID int) error {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		ginRouter.ServeHTTP(w, req)
		if w.Code < 200 || w.Code >= 300 {
			return fmt.Errorf("http status code: %d", w.Code)
		}
		return nil
	}

	if err := tester.BurnSteps(ctx, 5*time.Second, steps, taskFn); err != nil {
		t.Logf("压测异常退出: %v", err)
	}
	if err := tester.Wait(); err != nil && err != context.Canceled {
		t.Logf("压测等待异常: %v", err)
	}

	t.Log("限流压测报告：")
	tester.Stats().Report()
}

// 第1个压测场景:不经过限流器的接口
func TestNoLimiter(t *testing.T) {
	runBurnTest(t, "/api_no_limiter")
}

// 第2个压测场景:经过限流器的接口
func TestWithLimiter(t *testing.T) {
	runBurnTest(t, "/api_with_limiter")
}
