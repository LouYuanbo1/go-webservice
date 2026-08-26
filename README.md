# Go WebService 封装库

这是一个功能丰富的 Go 语言封装库，为 Web 服务开发提供了多种实用工具组件。该库利用 Go 1.27 的泛型方法特性进行了深度重构，提供更加优雅和类型安全的 API。旨在简化常见的 Web 开发任务，提高开发效率和代码质量。

## 📁 组件结构

该项目包含以下核心组件：

| 组件 | 目录 | 功能描述 |
|------|------|----------|
| **breaker** | `breaker/` | Google SRE 自适应熔断器，支持降级函数 |
| **cache** | `cache/` | 缓存组件，支持本地缓存和Redis缓存 |
| **elasticsearchx** | `elasticsearchx/` | Elasticsearch操作封装 |
| **errorx** | `errorx/` | 错误处理工具 |
| **gormc** | `gormc/` | GORM缓存连接 |
| **gormx** | `gormx/` | GORM ORM 框架的增强封装 |
| **limiter** | `limiter/` | 基于 Redis 的令牌桶限流器，支持本地降级 |
| **monitor** | `monitor/` | Prometheus 指标监控中间件 |
| **rabbitmq** | `rabbitmq/` | RabbitMQ 生产者和消费者封装 |
| **singleflightx** | `singleflightx/` | 单飞工具，避免缓存击穿 |

## 🚀 安装

使用 Go Modules 安装：

```bash
go get github.com/LouYuanbo1/go-webservice
......
```

## 📦 组件详情

### 1. cache

**功能**：提供统一的缓存接口，支持本地缓存（基于 freecache）和 Redis 缓存。使用 Go 1.27 泛型方法重构，提供类型安全的 API。

**核心接口**：
- `Cache` - 基础缓存接口，定义 Set、Get、Take、Del 方法
- `RedisCache` - Redis 缓存接口，继承 Cache 并提供 `GetRedisClient()` 方法
- `LocalCache` - 本地缓存接口，继承 Cache 并提供 `GetLocalCache()` 方法

**核心方法**（泛型版本）：
- `Set[T any](ctx context.Context, key string, val T, ttl time.Duration) error` - 设置缓存
- `Get[T any](ctx context.Context, key string, val *T) error` - 获取缓存
- `Take[T any](ctx context.Context, key string, val *T, query func(val *T) error, ttl time.Duration) error` - 缓存穿透处理
- `Del(ctx context.Context, keys ...string) error` - 删除缓存
- `GetRawCache() Cache` - 获取底层缓存接口

**特点**：
- 统一的缓存接口，支持多种驱动（Local/Redis）
- 泛型方法提供类型安全的 API
- 支持 `PointerModel` 接口约束
- 集成 singleflightx 防止缓存击穿
- 支持获取底层原生缓存客户端

**使用示例**：

```go
import (
    "context"
    "time"
    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cache/driver/local"
    "github.com/LouYuanbo1/go-webservice/cache/driver/redis"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建本地缓存驱动
localConfig := &local.Config{
    CacheSize: 100 * 1024 * 1024,
}
driver := local.NewDriver(localConfig, singleflightx.NewSingleFlight())

// 或创建Redis缓存驱动（不带熔断器）
redisConfig := &redis.Config{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    DB:       0,
}
driver := redis.NewDriver(redisConfig, singleflightx.NewSingleFlight())

// 打开缓存客户端
cacher, err := cache.Open(driver)
if err != nil {
    panic(err)
}
client := cache.NewClient(cacher)

// 设置缓存（泛型方法）
err = client.Set(context.Background(), "user:1", User{ID: 1, Name: "John"}, time.Hour)

// 获取缓存（泛型方法）
var user User
err = client.Get(context.Background(), "user:1", &user)

// 使用Take方法避免缓存穿透（泛型方法）
var result User
err = client.Take(context.Background(), "user:1", &result, func(val *User) error {
    // 从数据库获取数据
    *val = User{ID: 1, Name: "John"}
    return nil
}, time.Hour)
```

**集成熔断器示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/breaker"
    "github.com/LouYuanbo1/go-webservice/cache/driver/redis"
)

// 创建带熔断器的Redis缓存驱动
redisConfig := &redis.Config{
    Host:          "localhost",
    Port:          6379,
    Password:      "",
    DB:            0,
    Protocol:      2,                    // RESP3协议
    UnstableResp3: true,                 // 启用RESP3支持
    EnableBreaker: true,                 // 启用熔断器保护
}

// 自定义熔断器（可选）
customBreaker := breaker.NewBreaker(
    breaker.WithName("redis-cache-breaker"),
    breaker.WithProtection(100),
    breaker.WithK(2),
    breaker.WithWindow(10*time.Second),
)

// 创建Redis客户端（带熔断器钩子）
redisClient, err := redis.InitRedisClient(redisConfig)
if err != nil {
    panic(err)
}

// 创建Redis缓存
cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
if err != nil {
    panic(err)
}

client := cache.NewClient(cacher)
```

### 2. elasticsearchx

**功能**：Elasticsearch操作封装，提供文档CRUD、批量操作等功能。使用 Go 1.27 泛型方法重构，支持 `PointerDocument` 接口约束。

**核心接口**（泛型版本）：
- `CreateIndex[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 创建索引
- `IndexDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 索引文档
- `BulkIndexDocs[T any, PT PointerDocument[T]](ctx context.Context, docs []PT, cfg esutil.BulkIndexerConfig, stats bool) error` - 批量索引文档
- `GetDoc[T any, PT PointerDocument[T]](ctx context.Context, id string) (PT, error)` - 获取文档
- `FindDocsByPages[T any, PT PointerDocument[T]](ctx context.Context, page, size int) (*[]T, error)` - 分页查询文档
- `UpdateDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 更新文档
- `DeleteDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 删除文档
- `BulkDeleteDocs[T any, PT PointerDocument[T]](ctx context.Context, ids []string, cfg esutil.BulkIndexerConfig, stats bool) error` - 批量删除文档

**特点**：
- Go 1.27 泛型方法支持，类型安全的 API
- 批量操作支持，可配置统计信息
- 索引管理（创建/删除/获取索引列表）
- 详细的错误处理和日志记录
- 支持 `PointerDocument` 接口约束

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/elasticsearchx"
    "github.com/elastic/go-elasticsearch/v9"
    "github.com/elastic/go-elasticsearch/v9/esutil"
)

type Product struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

func (p *Product) Index() string {
    return "products"
}

func (p *Product) GetStringID() string {
    return p.ID
}

func (p *Product) GetTypeMapping() any {
    return map[string]any{
        "properties": map[string]any{
            "name": map[string]any{
                "type": "text",
            },
            "price": map[string]any{
                "type": "float",
            },
        },
    }
}

// 创建Elasticsearch客户端
client, _ := elasticsearch.NewTypedClient(elasticsearch.Config{
    Addresses: []string{"http://localhost:9200"},
})

// 创建ElasticsearchX实例（简化版）
es := elasticsearchx.NewElasticsearchX(client)

// 创建索引
doc := &Product{ID: "1", Name: "Product 1", Price: 100.0}
es.CreateIndex(context.Background(), doc)

// 索引文档
es.IndexDoc(context.Background(), doc)

// 批量索引文档
docs := []*Product{
    {ID: "2", Name: "Product 2", Price: 200.0},
    {ID: "3", Name: "Product 3", Price: 300.0},
}
es.BulkIndexDocs(context.Background(), docs, esutil.BulkIndexerConfig{}, true)

// 查询文档
result, err := es.GetDoc[Product](context.Background(), "1")

// 分页查询
docs, err := es.FindDocsByPages[Product](context.Background(), 1, 10)
```

### 3. errorx

**功能**：提供统一的错误处理机制，增强错误信息的可读性和可追踪性。

**核心功能**：
- 错误分类
- 详细的错误信息
- 错误链追踪

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/errorx"
)

// 创建错误
err := errorx.New(
    errorx.ErrInternal,
    "service",
    "Method",
    nil,
)

// 创建带详细信息的错误
err = errorx.NewWithDetails(
    errorx.ErrInternal,
    "service",
    "Method",
    "详细错误信息",
    nil,
)
```

### 4. gormc

**功能**：GORM 缓存层封装，实现 Cache-Aside 模式，在 `gormx.DB` 之上叠加 Redis 缓存能力。使用 Go 1.27 泛型方法重构，提供类型安全的缓存操作 API。

**核心方法**：
- `GetCache[T any](ctx context.Context, key string, val *T) error` — 从缓存获取数据
- `SetCache(ctx context.Context, key string, val any, opts ...TTLOption) error` — 写入缓存（支持自定义 TTL）
- `DelCache(ctx context.Context, keys ...string) error` — 删除缓存
- `Query[T any](ctx context.Context, key string, val *T, query func(ctx context.Context, db *gormx.DB, val *T) error, opts ...TTLOption) error` — Cache-Through 查询：缓存命中直接返回，未命中则回源 DB 并写缓存
- `QueryIndex[T any, ID comparable](ctx context.Context, key string, val *T, keyer func(primary ID) string, indexQuery func(ctx context.Context, db *gormx.DB, val *T) (primaryKey ID, err error), primaryQuery func(ctx context.Context, db *gormx.DB, val *T, primaryKey ID) error, opts ...TTLOption) error` — 二级缓存查询：索引缓存 → 主键缓存 → DB，解决缓存穿透
- `Exec(ctx context.Context, exec func(ctx context.Context, db *gormx.DB) error, keys ...string) error` — 执行 DB 写操作并自动删除关联缓存 key
- `ExecNoCache(ctx context.Context, exec func(ctx context.Context, db *gormx.DB) error) error` — 执行 DB 操作，不涉及缓存
- `QueryNoCache[T any](ctx context.Context, val *T, query func(ctx context.Context, db *gormx.DB, val *T) error) error` — 绕过缓存直接查 DB（单条）
- `QueryRowsNoCache[T any](ctx context.Context, val *[]T, query func(ctx context.Context, db *gormx.DB, val *[]T) error) error` — 绕过缓存直接查 DB（多条）
- `Transaction(ctx context.Context, fn func(tx *gormx.Executor) error) error` — 事务操作，回调接收 `*gormx.Executor`

**特点**：
- 实现 Cache-Aside 模式，自动管理缓存生命周期
- 支持二级缓存策略（索引缓存 + 主键缓存），缓存安全间隙防止数据不一致
- Go 1.27 泛型方法提供类型安全 API
- 事务操作直接透传底层 `gormx.DB.Transaction`

**使用示例**：

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cache/driver/redis"
    "github.com/LouYuanbo1/go-webservice/gormc"
    "github.com/LouYuanbo1/go-webservice/gormx"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建缓存客户端
redisDriver := redis.NewDriver(&redis.Config{Host: "localhost"}, singleflightx.NewSingleFlight())
cacher, _ := cache.Open(redisDriver)
cacheClient := cache.NewClient(cacher)

// 创建缓存数据库
cacheDB := gormc.NewCacheDB(gormx.NewDB(db), cacheClient, &gormc.Config{
    TTL:                                20 * time.Second,
    CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
})

// Cache-Through 查询：首次 Miss 回源 DB 并缓存，后续 Hit 直接返回
var user User
err := cacheDB.Query(context.Background(), "user:1", &user,
    func(ctx context.Context, db *gormx.DB, val *User) error {
        return db.First(ctx, val, 1)
    },
)

// 写操作 + 自动删缓存
err = cacheDB.Exec(context.Background(),
    func(ctx context.Context, db *gormx.DB) error {
        return db.Model(&User{}).Where("id = ?", 1).Update(ctx, "name", "Updated")
    },
    "user:1",
)

// 二级缓存查询（索引 → 主键 → DB）
var product Product
err = cacheDB.QueryIndex(context.Background(),
    "product:sku:abc123",
    &product,
    func(primary uint64) string { return fmt.Sprintf("product:%d", primary) },
    func(ctx context.Context, db *gormx.DB, val *Product) (uint64, error) {
        return getProductIDBySKU(ctx, db, "abc123")
    },
    func(ctx context.Context, db *gormx.DB, val *Product, primaryKey uint64) error {
        return db.First(ctx, val, primaryKey)
    },
)

// 事务
err = cacheDB.Transaction(context.Background(), func(tx *gormx.Executor) error {
    var u User
    tx.First(ctx, &u, 1)
    return tx.Model(&User{}).Where("id = ?", u.ID).Update(ctx, "age", 100)
})
```

### 5. gormx

**功能**：GORM 的增强封装，在原生 GORM 链式 API 基础上叠加**熔断器保护**和**泛型类型安全**。方法签名与 GORM 保持一致，学习成本极低。

**架构设计**：

```
DB {
    exec *Executor        // 直接持有，链式方法返回新 DB（Executor 本身支持链式）
    brk  breaker.Breaker  // 所有终结方法经熔断器保护
    acceptable func(error) bool
}
```

- **链式方法**（`Build`, `Model`, `Table`, `Raw`, `Clauses`, `Select`, `Where`, `StructFilter`, `MapFilter`, `Order`, `OrderByColumn`, `OrderBy`, `Joins`, `InnerJoins`, `Limit`, `Offset`, `Unscoped`, `Omit`, `Group`, `Having`）——返回新 `*DB`，行为与 GORM 一致。
- **终结方法**（`Create`, `CreateInBatches`, `First`, `Find`, `Count`, `Update`, `Updates`, `Delete`, `Save`, `Pluck`, `Scan`）——执行实际 SQL，结果经熔断器 `DoWithAcceptable` 包裹。
- **事务** `Transaction(ctx, fn func(tx *Executor) error)` ——开启事务，回调接收 `*Executor`。

**三大亮点**：

| 特性 | 说明 |
|------|------|
| **熔断器** | 所有终结方法自动走 `DoWithAcceptable`，数据库故障时快速熔断，避免雪崩 |
| **泛型安全** | 利用 Go 1.27 泛型方法，`Create`、`Find` 等直接约束 `PointerModel[T]`，编译期杜绝类型错误 |
| **API 一致** | 方法与 GORM 同名同参，无需学习新 DSL，上手即用 |

**选项模式**：
- `WithBreaker(brk breaker.Breaker)` — 自定义熔断器
- `WithAcceptable(acc func(err error) bool)` — 自定义错误白名单（默认忽略 `gorm.ErrRecordNotFound` 和 `gorm.ErrInvalidTransaction`）

**使用示例**：

```go
import (
    "context"
    "time"

    "github.com/LouYuanbo1/go-webservice/breaker"
    "github.com/LouYuanbo1/go-webservice/gormx"
    "gorm.io/gorm"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// 1. 基础用法：无熔断器（使用默认 breaker）
db, _ := gorm.Open(...)
xdb := gormx.NewDB(db)

// 链式查询
var users []User
err := xdb.Where("age > ?", 20).Order("id DESC").Limit(10).Find(context.Background(), &users)

// 创建
user := &User{Name: "John", Age: 30}
err = xdb.Create(context.Background(), user)

// 批量创建
batch := []*User{{Name: "A", Age: 20}, {Name: "B", Age: 21}}
err = xdb.CreateInBatches(context.Background(), &batch, 100)

// 聚合
var count int64
err = xdb.Model(&User{}).Where("age > ?", 18).Count(context.Background(), &count)

// 事务
err = xdb.Transaction(context.Background(), func(tx *gormx.Executor) error {
    var u User
    tx.First(ctx, &u, 1)
    return tx.Model(&User{}).Where("id = ?", u.ID).Update(ctx, "age", 100)
})

// 2. 接入熔断器
customBreaker := breaker.NewBreaker(
    breaker.WithName("db-breaker"),
    breaker.WithProtection(100),
    breaker.WithK(2),
    breaker.WithWindow(10*time.Second),
)

xdb = gormx.NewDB(db,
    gormx.WithBreaker(customBreaker),
    gormx.WithAcceptable(func(err error) bool {
        return errorx.In(err, gorm.ErrRecordNotFound, gorm.ErrInvalidTransaction)
    }),
)

// 此后所有 Create/Find/Update/Delete 等操作均自动受熔断器保护
err = xdb.Create(context.Background(), user)
```

### 6. singleflightx

**功能**：提供单飞功能，避免缓存击穿。支持泛型版本，提供类型安全的并发请求合并。

**核心接口**：
- `Do(key string, fn func() (any, error)) (any, error)` - 执行单飞操作
- `DoEx(key string, fn func() (any, error)) (val any, fresh bool, err error)` - 执行单飞操作，返回是否为新计算结果

**泛型版本**：
- `TypedSingleFlight[T any]` - 泛型版本的单飞接口
- `NewTypedSingleFlight[T any]() TypedSingleFlight[T]` - 创建泛型单飞实例
- `Do(key string, fn func() (T, error)) (T, error)` - 执行泛型单飞操作
- `DoEx(key string, fn func() (T, error)) (val T, fresh bool, err error)` - 执行泛型单飞操作，返回是否为新计算结果

**特点**：
- 并发请求合并，有效防止缓存击穿
- 泛型版本提供类型安全的操作
- 支持判断结果是否为新计算（fresh）
- 简单易用的 API

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建单飞实例
sf := singleflightx.NewSingleFlight()

// 执行单飞操作
result, err := sf.Do("key", func() (any, error) {
    // 执行耗时操作，如数据库查询
    return "result", nil
})
```

**泛型版本使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 定义结果类型
type User struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// 创建泛型单飞实例
sf := singleflightx.NewTypedSingleFlight[User]()

// 执行单飞操作（泛型版本）
result, err := sf.Do("user:1", func() (User, error) {
    // 执行耗时操作，如数据库查询
    return User{ID: 1, Name: "John", Age: 30}, nil
})

// 执行带额外信息的单飞操作（泛型版本）
result, fresh, err := sf.DoEx("user:1", func() (User, error) {
    // 执行耗时操作，如数据库查询
    return User{ID: 1, Name: "John", Age: 30}, nil
})
// fresh 表示结果是否是新计算的
```

### 7. breaker

**功能**：实现 Google SRE 自适应熔断器模式，用于保护服务免受级联故障影响。

**核心接口**：
- `Do(ctx context.Context, req func(ctx context.Context) error) error` - 执行请求
- `DoWithAcceptable(ctx context.Context, req func(ctx context.Context) error, acceptable func(err error) bool) error` - 执行请求，自定义成功判定
- `DoWithFallback(ctx context.Context, req func(ctx context.Context) error, fallback func(err error) error) error` - 执行请求，支持降级函数
- `GetMetrics() (total, accepts int64, rate float64)` - 获取当前窗口指标
- `Reset()` - 手动重置熔断器

**特点**：
- 基于 Google SRE 自适应熔断算法
- 支持 Context 传播
- 支持自定义降级函数
- 内置 Prometheus 指标导出支持
- 支持手动重置（运维干预）

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/breaker"
)

// 创建熔断器
b := breaker.NewBreaker(
    breaker.WithName("service-breaker"),
    breaker.WithProtection(100),
    breaker.WithK(2),
    breaker.WithWindow(10*time.Second),
    breaker.WithOnReject(func() {
        // 熔断触发时的回调
        log.Println("circuit breaker tripped")
    }),
)

// 执行请求
err := b.Do(context.Background(), func(ctx context.Context) error {
    // 调用外部服务
    return callExternalService(ctx)
})

// 执行请求并支持降级
err := b.DoWithFallback(context.Background(), func(ctx context.Context) error {
    return callExternalService(ctx)
}, func(err error) error {
    // 降级逻辑
    return getFallbackData()
})
```

### 8. limiter

**功能**：基于 Redis 的令牌桶限流器，支持本地降级策略。

**核心接口**：
- `Allow(ctx context.Context) bool` - 允许单个请求
- `AllowN(ctx context.Context, now time.Time, n int) bool` - 允许 n 个请求

**特点**：
- 基于 Redis 实现分布式限流
- Redis 故障时自动切换到本地限流
- 自动检测 Redis 恢复并切回
- 支持突发流量（burst）

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/limiter"
    "github.com/redis/go-redis/v9"
)

// 创建 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建限流器（每秒 100 个请求，突发 200 个）
lim := limiter.NewTokenLimiter(100, 200, redisClient, "api:rate:limiter")

// 检查是否允许请求
if lim.Allow(context.Background()) {
    // 处理请求
    handleRequest()
} else {
    // 返回限流响应
    http.Error(w, "too many requests", http.StatusTooManyRequests)
}
```

**可以作为中间件使用，限制请求频率**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/limiter"
    "github.com/redis/go-redis/v9"
    "github.com/gin-gonic/gin"
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

// 创建 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建限流器（每秒 100 个请求，突发 200 个）
lim := limiter.NewTokenLimiter(100, 200, redisClient, "api:rate:limiter")
r := gin.Default()
r.GET("/", TokenLimiterMiddleware(lim), func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
})
```


### 9. monitor

**功能**：Prometheus 指标监控中间件，用于收集 HTTP 请求指标。

**核心接口**：
- `Handler(next http.Handler) http.Handler` - HTTP 中间件
- `Record(path string, statusCode int, duration float64)` - 手动记录指标
- `AddCustomCallback(callback func(*RequestMetrics))` - 添加自定义回调

**收集的指标**：
- `http_requests_total` - 请求总数（按路径和状态码分组）
- `http_request_duration_seconds` - 请求耗时直方图

**特点**：
- 支持自定义命名空间和子系统
- 支持自定义指标收集器
- 支持自定义回调处理
- 支持手动记录指标

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/monitor"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

// 创建监控中间件
mw, err := monitor.NewMetricsMiddleware(monitor.MetricsConfig{
    Namespace: "myapp",
    Subsystem: "api",
}, nil)

// 添加自定义回调
mw.AddCustomCallback(func(m *monitor.RequestMetrics) {
    // 自定义处理逻辑
    log.Printf("Request: %s %d %.2fs", m.Path, m.StatusCode, m.Duration)
})

// 注册 Prometheus 端点
http.Handle("/metrics", promhttp.Handler())

// 使用中间件
http.Handle("/api", mw.Handler(http.HandlerFunc(apiHandler)))
```


**可以作为中间件使用，记录请求指标**：

```go
import (
    "context"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/gin-gonic/gin"
)

func GinMetricsMiddleware(mw *MetricsMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		path := c.Request.URL.Path
		status := c.Writer.Status()

		mw.Record(path, status, duration)
	}
}

reg := prometheus.NewRegistry()

// 创建监控中间件
mw, err := NewMetricsMiddleware(MetricsConfig{
		Namespace: "ginadapter",
	}, reg)

r := gin.Default()
r.GET("/", GinMetricsMiddleware(mw), func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
})
```

### 10. rabbitmq

**功能**：RabbitMQ 生产者和消费者封装，提供简洁的消息发布和订阅接口。

**核心接口（Producer）**：
- `Produce(exchange string, routeKey string, msg []byte) error` - 发送消息

**核心接口（Consumer）**：
- `Start()` - 启动消费者
- `Stop()` - 停止消费者

**配置选项**：
- `RabbitConfig` - RabbitMQ 连接配置
- `RabbitProducerConfig` - 生产者配置
- `RabbitConsumerConfig` - 消费者配置

**特点**：
- 支持自动确认和手动确认模式
- 支持多个监听队列
- 完善的错误处理和日志记录
- 支持恐慌恢复和消息重入队列

**使用示例（生产者）**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/rabbitmq"
)

// 创建生产者
producer, err := rabbitmq.NewProducer(rabbitmq.RabbitProducerConfig{
    RabbitConfig: rabbitmq.RabbitConfig{
        Username: "guest",
        Password: "guest",
        Host:     "localhost",
        Port:     5672,
        VHost:    "/",
    },
    ContentType: "application/json",
})

// 发送消息
err = producer.Produce("exchange-name", "route-key", []byte(`{"message": "hello"}`))
```

**使用示例（消费者）**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/rabbitmq"
)

// 消息处理函数
handler := func(msg []byte) error {
    // 处理消息
    log.Printf("Received: %s", string(msg))
    return nil
}

// 创建消费者
consumer, err := rabbitmq.NewConsumer(rabbitmq.RabbitConsumerConfig{
    RabbitConfig: rabbitmq.RabbitConfig{
        Username: "guest",
        Password: "guest",
        Host:     "localhost",
        Port:     5672,
        VHost:    "/",
    },
    ListenerQueues: []rabbitmq.ConsumerConfig{
        {
            Name:    "queue-name",
            AutoAck: false, // 手动确认
        },
    },
}, handler)

// 启动消费者
consumer.Start()

// 停止消费者（通常在应用关闭时调用）
// consumer.Stop()
```

## 🛠️ 技术栈

| 依赖 | 用途 |
|------|------|
| github.com/coocood/freecache | 本地缓存实现 |
| github.com/redis/go-redis/v9 | Redis 客户端 |
| github.com/elastic/go-elasticsearch/v9 | Elasticsearch 客户端 |
| gorm.io/gorm | ORM 框架 |
| golang.org/x/time/rate | 本地限流器实现 |
| github.com/prometheus/client_golang | Prometheus 指标收集 |
| github.com/rabbitmq/amqp091-go | RabbitMQ 客户端 |
| github.com/LouYuanbo1/go-burn | 日志库 |

## ✨ 亮点特性

1. **Go 1.27 泛型方法**：利用 Go 1.27 引入的泛型方法特性进行深度重构，提供更加优雅和类型安全的 API
2. **模块化设计**：各组件独立封装，可单独使用
3. **简洁易用**：提供直观的 API 接口，简化常见操作
4. **功能丰富**：涵盖 Web 开发中常见的多种工具需求
5. **配置灵活**：支持详细的配置选项，满足不同场景需求
6. **性能优化**：基于成熟的第三方库，提供高性能实现
7. **错误处理**：统一的错误处理机制，增强错误信息的可读性
8. **缓存策略**：实现 Cache-Aside 模式，提供缓存一致性管理
9. **批量操作**：支持批量处理，提高性能
10. **事务支持**：提供事务操作，确保数据一致性

## 🤝 贡献指南

欢迎报告问题或提出建议！

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

**注意**：本库基于 Go 1.27 开发，使用了 Go 1.27 引入的泛型方法特性进行重构。由于 Go 1.27 目前仍处于开发分支阶段，正式版本尚未公布，您需要使用 Go 1.27 开发版本才能编译和使用本库。