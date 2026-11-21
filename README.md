# Go WebService 封装库

这是一个功能丰富的 Go 语言封装库，为 Web 服务开发提供了多种实用工具组件。该库利用 Go 1.27 的泛型方法特性进行了深度重构，提供更加优雅和类型安全的 API。旨在简化常见的 Web 开发任务，提高开发效率和代码质量。

## 📁 组件结构

该项目包含以下核心组件：

| 组件 | 目录 | 功能描述 |
|------|------|----------|
| **antx** | `db/antx` | ant ORM 的熔断保护中间件（基于 ant 原生 Interceptor） |
| **antc** | `cachedb/antc` | ant ORM 整库缓存层（跨表操作 + 事务透传） |
| **breaker** | `breaker/` | Google SRE 自适应熔断器，支持降级函数 |
| **cache** | `cache/` | 统一缓存接口与泛型封装，支持本地缓存和 Redis 缓存 |
| **cachedb** | `cachedb/` | 缓存层公共配置（TTL、安全间隙、TTL 选项） |
| **elasticsearchx** | `elasticsearchx/` | Elasticsearch 操作封装 |
| **errorx** | `errorx/` | 错误处理工具 |
| **event** | `event/` | 泛型事件总线（本地/分布式驱动：Local、Redis、RabbitMQ） |
| **gormc** | `cachedb/gormc` | GORM 缓存连接（Cache-Aside 模式） |
| **gormx** | `db/gormx` | GORM ORM 框架的增强封装（熔断 + 泛型 + Trace） |
| **limiter** | `limiter/` | 基于 Redis 的令牌桶限流器，支持本地降级 |
| **monitor** | `monitor/` | Prometheus 指标监控中间件 |
| **rabbitmq** | `rabbitmq/` | RabbitMQ 生产者和消费者封装 |
| **singleflightx** | `singleflightx/` | 单飞工具，避免缓存击穿 |
| **trace** | `trace/` | OpenTelemetry Trace 链路追踪 |

## 🚀 安装

使用 Go Modules 安装：

```bash
go get github.com/LouYuanbo1/go-webservice
```

## 📦 组件详情

### 1. cachedb

**功能**：缓存层的公共配置封装，供 `cachedb/antc` 与 `cachedb/gormc` 共用。包括默认 TTL、二级缓存的"索引缓存与主键缓存安全间隙"以及 TTL 选项。

**核心类型**：

```go
type Config struct {
    TTL                                time.Duration // 缓存默认过期时间
    CacheSafeGapBetweenIndexAndPrimary time.Duration // 主键缓存相对索引缓存多保留的时间
}

type TTLOption func(*ttl)

func WithTTL(value time.Duration) TTLOption // 单次调用指定 TTL
func BuildTTL(cfg *Config, opts ...TTLOption) time.Duration // 从配置和选项计算 TTL
func SafeGap(cfg *Config) time.Duration // 计算索引缓存与主键缓存的安全间隙
```

**特点**：
- `WithTTL` 支持单次调用覆盖默认 TTL，其余调用走配置默认值
- `SafeGap` 保证主键缓存比索引缓存稍晚过期，避免索引未过期而主键先行过期造成的额外数据库访问

### 2. cache

**功能**：提供统一的缓存接口与泛型封装。底层驱动支持本地缓存（基于 freecache）和 Redis 缓存，上层 `Client` 使用 Go 1.27 泛型方法提供类型安全的 API。

**核心接口**：

```go
// Cache 底层缓存接口（驱动实现）
type Cache interface {
    Set(ctx context.Context, key string, val any, ttl time.Duration) error
    Get(ctx context.Context, key string, val any) error
    Take(ctx context.Context, key string, val any,
        query func(val any) error, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
}

// Client 泛型封装
type Client struct{ Cache }

func NewClient(cache Cache) *Client
```

**核心方法**（`Client` 泛型版本）：
- `Set[T any](ctx context.Context, key string, val T, ttl time.Duration) error` - 设置缓存
- `Get[T any](ctx context.Context, key string, val *T) error` - 获取缓存
- `Take[T any](ctx context.Context, key string, val *T, query func(val *T) error, ttl time.Duration) error` - 缓存穿透处理（内建 singleflight 合并回源）
- `Del(ctx context.Context, keys ...string) error` - 删除缓存
- `GetRawCache() Cache` - 获取底层缓存接口

**驱动**：
- `cache/driver/local`：`local.Config{CacheSize}` → `local.NewLocalCache(config, singleflightx.SingleFlight)`，暴露 `GetLocalCache() *freecache.Cache`
- `cache/driver/redis`：`redis.Config{Host, Port, Password, DB, Protocol, UnstableResp3, EnableBreaker}` → `redis.Init(config, hooks...)` 创建连接，`redis.NewRedisCache(client, singleflightx.SingleFlight)` 创建缓存，暴露 `GetRedisClient() *redis.Client`

**特点**：
- 统一的缓存接口，支持多种驱动（Local/Redis）
- 泛型 `Client` 提供类型安全的 API
- `Take` 集成 singleflightx 防止缓存击穿（并发请求合并，只有一个发起回源）
- Redis 驱动可内建 Google SRE 熔断器钩子（`EnableBreaker: true`）防止雪崩

**使用示例**：

```go
import (
    "context"
    "time"
    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cache/driver/local"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建本地缓存
localConfig := &local.Config{
    CacheSize: 100 * 1024 * 1024,
}
cacher, err := local.NewLocalCache(localConfig, singleflightx.NewSingleFlight())
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

**Redis 驱动示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cache/driver/redis"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建Redis配置
redisConfig := &redis.Config{
    Host:          "localhost",
    Port:          6379,
    Password:      "",
    DB:            0,
    Protocol:      2,               // RESP3 协议（向量搜索等功能必需）
    UnstableResp3: true,            // 启用 RESP3 支持
    EnableBreaker: true,            // 启用内建熔断器保护
}

// 创建Redis客户端（EnableBreaker 时自动挂载 BreakerHook）
redisClient, err := redis.Init(redisConfig)
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

### 3. antc

**功能**：ant ORM 的**整库缓存层**。包住用户自定义的 DB 结构体（内嵌 `ant.Context`），查询与写入闭包可直接访问任意表，实现跨表操作与统一缓存失效。事务通过 `AntContext().Tx` 透传给 ant。

**核心 API**：

```go
// 约束接口：要求用户 DB 结构体实现，暴露内嵌的 *ant.Context
type HasAntContext[T any] interface {
    *T
    AntContext() *ant.Context
}

func NewCacheDB[T any, DB HasAntContext[T]](db DB, c *cache.Client, cfg *cachedb.Config) *CacheDB[T, DB]
```

- `GetDB() DB` — 返回底层 `*DB`，执行未缓存的原生读写
- `GetCache[V any](ctx, key string, val *V) error` — 直接读取缓存
- `SetCache(ctx, key string, val any, opts ...cachedb.TTLOption) error` — 直接写入缓存
- `DelCache(ctx, key ...string) error` — 删除缓存
- `Exec(ctx, exec func(ctx, db DB) error, keys ...string) error` — 写入后自动删除关联缓存（可跨任意表）
- `ExecNoCache(ctx, exec func(ctx, db DB) error) error` — 写入但不清理缓存
- `Query[R any](ctx, key string, query func(ctx, db DB) (*R, error), opts ...cachedb.TTLOption) (*R, error)` — 缓存优先单行查询
- `QueryIndex[R any, ID comparable](ctx, key string, keyer func(primary ID) string, indexQuery func(ctx, db DB) (primaryKey ID, data *R, err error), primaryQuery func(ctx, db DB, primaryKey ID) (*R, error), opts ...cachedb.TTLOption) (*R, error)` — 二级缓存查询
- `QueryNoCache[R any](ctx, query func(ctx, db DB) (*R, error)) (*R, error)` — 直接查库（不缓存）
- `QueryRowsNoCache[R any](ctx, query func(ctx, db DB) ([]*R, error)) ([]*R, error)` — 直接查多行（不缓存）
- `Transaction(ctx, fn func(tx *T) error, opts ...ant.TxOption) error` — 事务内使用 `tx` 访问数据

**特点**：

| 特性 | 说明 |
|------|------|
| **跨表操作** | 查询/写入闭包持有整个 DB 结构体，可访问任意 `ant.Table` |
| **统一缓存失效** | `Exec` 一次写入可同时删除多个关联缓存 key |
| **二级缓存** | 索引缓存 + 主键缓存 + 安全间隙，防止缓存穿透与不一致 |
| **事务透传** | 事务经 `AntContext().Tx` 转发给 ant，提交/回滚由 ant 管理 |

**使用示例**：

```go
import (
    "context"
    "time"

    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cache/driver/redis"
    "github.com/LouYuanbo1/go-webservice/cachedb"
    "github.com/LouYuanbo1/go-webservice/cachedb/antc"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
    "github.com/skudra-go/ant"
    "github.com/skudra-go/ant/dialect/sqlite"
)

// 用户自定义 DB 结构体：必须实现 AntContext 方法
type AppDB struct {
    ant.Context
    Users ant.Table[User]
    Posts ant.Table[Post]
}

func (db *AppDB) AntContext() *ant.Context { return &db.Context }

// 打开 ant DB
db, _ := ant.Open[AppDB](ctx, sqlite.Dialect, ":memory:")

// 创建缓存客户端
redisClient, _ := redis.Init(&redis.Config{Host: "localhost"})
cacher, _ := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
cacheClient := cache.NewClient(cacher)

// 创建整库缓存层
cdb := antc.NewCacheDB[AppDB, *AppDB](db, cacheClient, &cachedb.Config{
    TTL:                                20 * time.Second,
    CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
})

// 缓存优先查询（闭包可访问任意表）
user, err := cdb.Query(ctx, "user:1", func(ctx context.Context, db *AppDB) (*User, error) {
    return db.Users.First(ctx, 1) // 注意：需转换为 *User
})

// 写入 + 自动删缓存
err = cdb.Exec(ctx, func(ctx context.Context, db *AppDB) error {
    _, err := db.Users.Insert(ctx, &User{Name: "John"})
    return err
}, "user:1")

// 跨表事务：事务内使用 tx 访问数据
err = cdb.Transaction(ctx, func(tx *AppDB) error {
    _, err := tx.Users.Insert(ctx, &User{Name: "John"})
    return err
})
```

### 4. gormc

**功能**：GORM 缓存层封装，实现 Cache-Aside 模式，在 `gormx.DB` 之上叠加 Redis 缓存能力。使用 Go 1.27 泛型方法重构，提供类型安全的缓存操作 API。

**核心方法**：
- `GetXDB() *gormx.DB` — 返回底层 `*gormx.DB`
- `GetCache[T any](ctx context.Context, key string, val *T) error` — 从缓存获取数据
- `SetCache(ctx context.Context, key string, val any, opts ...cachedb.TTLOption) error` — 写入缓存（支持自定义 TTL）
- `DelCache(ctx context.Context, keys ...string) error` — 删除缓存
- `Query[T any](ctx context.Context, key string, val *T, query func(ctx context.Context, db *gormx.DB, val *T) error, opts ...cachedb.TTLOption) error` — Cache-Through 查询：缓存命中直接返回，未命中则回源 DB 并写缓存
- `QueryIndex[T any, ID comparable](ctx context.Context, key string, val *T, keyer func(primary ID) string, indexQuery func(ctx context.Context, db *gormx.DB, val *T) (primaryKey ID, err error), primaryQuery func(ctx context.Context, db *gormx.DB, val *T, primaryKey ID) error, opts ...cachedb.TTLOption) error` — 二级缓存查询：索引缓存 → 主键缓存 → DB，解决缓存穿透
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
    "github.com/LouYuanbo1/go-webservice/cachedb"
    "github.com/LouYuanbo1/go-webservice/cachedb/gormc"
    "github.com/LouYuanbo1/go-webservice/db/gormx"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建缓存客户端
redisClient, _ := redis.Init(&redis.Config{Host: "localhost"})
cacher, _ := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
cacheClient := cache.NewClient(cacher)

// 创建缓存数据库（gormx.DB 包住 *gorm.DB）
cacheDB := gormc.NewCacheDB(gormx.NewDB(db), cacheClient, &cachedb.Config{
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

**功能**：GORM 的增强封装，在原生 GORM 链式 API 基础上叠加**熔断器保护**、**泛型类型安全**与 **OpenTelemetry Trace**。方法签名与 GORM 保持一致，学习成本极低。

**架构设计**：

```
DB {
    exec *Executor        // 直接持有，链式方法返回新 DB（Executor 本身支持链式）
    brk  breaker.Breaker  // 所有终结方法经熔断器保护
    acceptable func(error) bool
}
```

- **链式方法**（`Build`, `Model`, `Table`, `Raw`, `Clauses`, `Select`, `Where`, `StructFilter`, `MapFilter`, `Order`, `OrderByColumn`, `OrderBy`, `Joins`, `InnerJoins`, `Limit`, `Offset`, `Unscoped`, `Omit`, `Group`, `Having`）——返回新 `*DB`，行为与 GORM 一致。
- **终结方法**（`Exec`, `Create`, `CreateInBatches`, `First`, `Find`, `FindInBatches`, `Count`, `Update`, `Updates`, `UpdatesByStruct`, `UpdatesByMap`, `Delete`, `DeleteByStruct`, `DeleteByMap`, `Save`, `Pluck`, `Scan`）——执行实际 SQL，结果经熔断器 `DoWithAcceptable` 包裹。
- **事务** `Transaction(ctx, fn func(tx *Executor) error)` ——开启事务，回调接收 `*Executor`。

**三大亮点**：

| 特性 | 说明 |
|------|------|
| **熔断器** | 所有终结方法自动走 `DoWithAcceptable`，数据库故障时快速熔断，避免雪崩 |
| **泛型安全** | 利用 Go 1.27 泛型方法，`Create`、`Find` 等直接约束 `PointerModel[T]`，编译期杜绝类型错误 |
| **Trace 集成** | 每条 SQL 操作自动创建 OpenTelemetry span，`gorm.ErrRecordNotFound` 视为成功状态 |

**选项模式**：
- `WithBreaker(brk breaker.Breaker)` — 自定义熔断器
- `WithAcceptable(acc func(err error) bool)` — 自定义错误白名单（默认忽略 `gorm.ErrRecordNotFound` 和 `gorm.ErrInvalidTransaction`）

**使用示例**：

```go
import (
    "context"
    "time"

    "github.com/LouYuanbo1/go-webservice/breaker"
    "github.com/LouYuanbo1/go-webservice/db/gormx"
    "gorm.io/gorm"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// 1. 基础用法：无自定义熔断器（使用默认 breaker）
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

// 2. 接入自定义熔断器
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

### 6. antx

**功能**：ant ORM 的增强封装，基于 ant 原生 **Interceptor 中间件**为整个 Context 安装**熔断保护**。得益于 ant 的架构（所有语句统一从 `state.exec` 派发），无需包装 `ant.Context` 或 `ant.Table[T]`，即可拦截该 Context 派生的所有 Session 执行的每一条 SQL（含 Table 方法、Query 链、SaveChanges、Raw、事务）。

**核心 API**：
- `WithBreaker(opts ...breaker.Option) ant.Option` — 零配置安装熔断中间件，熔断器在 `ant.Open` 时创建一次
- `WithCustomBreaker(brk breaker.Breaker, acceptable func(error) bool) ant.Option` — 使用自建熔断器与自定义成功判定函数
- `DefaultAcceptable(err error) bool` — 默认成功判定（容忍 `nil`、`sql.ErrNoRows`、`sql.ErrTxDone`）

**特点**：

| 特性 | 说明 |
|------|------|
| **零侵入** | 不要求包装或继承任何 ant 类型，`ant.Context` + `ant.Table[T]` 结构体原样使用 |
| **Context 级共享** | 熔断器随 `ant.Open` 创建一次，被所有 Session 共享，失败统计跨请求持续累积 |
| **全覆盖** | 在 `Before` 触库前做准入判定，熔断时快速拒绝；`After` 按真实结果回写窗口 |

**使用示例**：

```go
import (
    "context"

    "github.com/LouYuanbo1/go-webservice/breaker"
    "github.com/LouYuanbo1/go-webservice/db/antx"
    "github.com/skudra-go/ant"
    "github.com/skudra-go/ant/dialect/sqlite"
)

// 业务结构体保持原样，无需任何升级仪式
type AppDB struct {
    ant.Context
    Users ant.Table[User]
    Posts ant.Table[Post]
}

// 1. 基础用法：零配置，使用默认熔断器
db, _ := ant.Open[AppDB](ctx, sqlite.Dialect, dsn, antx.WithBreaker())

// 2. 接入自建熔断器
brk := breaker.NewBreaker(
    breaker.WithName("ant-db-breaker"),
    breaker.WithProtection(100),
    breaker.WithK(2),
    breaker.WithWindow(10*time.Second),
)
db, _ = ant.Open[AppDB](ctx, sqlite.Dialect, dsn, antx.WithCustomBreaker(brk, nil))

// 此后该 Context 及派生 Session 的所有表操作均自动受熔断器保护
user, err := db.Users.First(ctx)
```

### 7. event

**功能**：泛型事件总线。提供类型安全的 `Subscribe`/`Publish` 泛型封装（`TypedBus`），底层支持进程内本地总线（`LocalBus`）与分布式驱动（Redis Pub/Sub、RabbitMQ topic 交换机），并内置 JSON 编解码器。

**架构设计**：

```
事件对象 T
   │  TypedBus 泛型包装（编解码 + 主题推导 + 过滤）
   ▼
Bus 接口（字节负载传输层）
   ├── event/driver/redis    （Redis Pub/Sub，发布即丢弃）
   ├── event/driver/rabbitmq （RabbitMQ topic 交换机，消息持久化）
   └── localbus.LocalBus     （进程内直派结构体，无序列化）
```

**核心接口**：

```go
type Bus interface {
    Subscribe(topic string, handler EventHandler) (Subscription, error)
    Publish(ctx context.Context, topic string, payload []byte) error
    Shutdown(ctx context.Context) error
}

type Subscription interface {
    Unsubscribe()
}
```

**泛型封装（TypedBus）**：
- `NewTypedBus(raw Bus, opts ...Option) *TypedBus` — 基于任意 `Bus` 驱动构建
- `Subscribe[T any](handler func(ctx context.Context, event T) error, opts ...Option) (Subscription, error)` — 订阅某类型事件
- `Publish[T any](ctx context.Context, event T, opts ...Option) error` — 发布某类型事件
- `Shutdown(ctx context.Context) error` — 关闭总线

**事件选项**：
- `WithTopic(topic string)` — 自定义主题（默认 `event.<类型名>`）
- `WithFilter[T any](fn func(event T) bool)` — 订阅过滤器，返回 false 的事件被忽略
- `WithCodec(c Codec)` — 自定义编解码器（默认 `JSONCodec`）

**本地总线（LocalBus）**：
- `NewLocalBus() *LocalBus` — 进程内事件总线，直接以结构体派发、不序列化
- `Subscribe[T any](handler, opts...) LocalSub` — 返回本地订阅，调用可取消

**分布式驱动**：
- `event/driver/redis.NewBus(&redis.Config{Host, Port, Password, DB})` — 基于 Redis Pub/Sub，消息发布即丢弃，支持多订阅者并发投递
- `event/driver/rabbitmq.NewBus(&rabbitmq.Config{Username, Password, Host, Port, VHost, Exchange})` — 基于 RabbitMQ topic 交换机，Exchange 默认 `event.bus`，每条订阅独占临时队列并自动删除

**使用示例**：

```go
import (
    "context"
    "log"

    "github.com/LouYuanbo1/go-webservice/event"
    "github.com/LouYuanbo1/go-webservice/event/driver/redis" // 或 rabbitmq
)

type OrderCreated struct {
    OrderID uint64
    UserID  uint64
}

// 创建分布式驱动（此处用 Redis）
bus, err := redis.NewBus(&redis.Config{Host: "localhost", Port: 6379})
if err != nil {
    panic(err)
}

// 泛型包装
tb := event.NewTypedBus(bus)

// 订阅：自动推导主题 event.OrderCreated
sub, err := tb.Subscribe(func(ctx context.Context, e OrderCreated) error {
    log.Printf("order %d created by user %d", e.OrderID, e.UserID)
    return nil
})
if err != nil {
    panic(err)
}
defer sub.Unsubscribe()

// 发布
err = tb.Publish(context.Background(), OrderCreated{OrderID: 1, UserID: 42})
```

**进程内本地总线示例**：

```go
bus := event.NewLocalBus() // 或 event.LocalBus 作为独立的进程内总线

unsub := bus.Subscribe(func(ctx context.Context, e OrderCreated) error {
    // 进程内直接接收结构体，无序列化
    return nil
})

err := bus.Publish(context.Background(), OrderCreated{OrderID: 2, UserID: 7})
```

### 8. trace

**功能**：OpenTelemetry Trace 链路追踪。提供 `TracerProvider` 初始化与从 Context 恢复 Tracer 的辅助方法，`gormx` 已内建自动埋点。

**核心 API**：
- `NewTracerProvider() (*sdkTrace.TracerProvider, error)` — 创建 stdout 导出的 TracerProvider（批量导出、1s 抽片、全部采样）并注册为全局 Provider
- `TracerFromContext(ctx context.Context) oteltrace.Tracer` — 从 Context 中的 Span 恢复对应 Provider 的 Tracer（无有效 Span 时回退全局 Tracer）
- `TraceName = "go-webservice"` — 库默认 Tracer 名称

**使用示例**：

```go
import (
    "context"

    "github.com/LouYuanbo1/go-webservice/trace"
)

// 初始化 Tracer Provider（通常在应用启动时）
tp, err := trace.NewTracerProvider()
if err != nil {
    panic(err)
}
defer tp.Shutdown(context.Background())

// 业务代码中获取 Tracer 并创建 Span
tracer := trace.TracerFromContext(ctx)
ctx, span := tracer.Start(ctx, "business.op")
defer span.End()

// 说明：db/gormx 的每条 SQL 操作也会自动创建 gormx.* 的 span
```

### 9. elasticsearchx

**功能**：Elasticsearch 操作封装，提供文档 CRUD、批量操作、索引管理等功能。使用 Go 1.27 泛型方法重构，支持 `PointerDocument` 接口约束。

**核心接口**（泛型版本）：
- `CreateIndex[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 创建索引
- `GetIndices(ctx context.Context) (map[string]string, error)` - 获取索引及文档数统计
- `DeleteIndex[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 删除索引
- `IndexDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 索引文档
- `BulkIndexDocs[T any, PT PointerDocument[T]](ctx context.Context, docs []PT, cfg esutil.BulkIndexerConfig, stats bool) error` - 批量索引文档
- `GetDoc[T any, PT PointerDocument[T]](ctx context.Context, id string) (PT, error)` - 获取文档
- `FindDocsByPages[T any, PT PointerDocument[T]](ctx context.Context, page, size int) (*[]T, error)` - 分页查询文档
- `CountDocs[T any, PT PointerDocument[T]](ctx context.Context) (int64, error)` - 统计文档数
- `UpdateDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 更新文档
- `DeleteDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error` - 删除文档
- `BulkDeleteDocs[T any, PT PointerDocument[T]](ctx context.Context, ids []string, cfg esutil.BulkIndexerConfig, stats bool) error` - 批量删除文档

**模型约束**：

```go
type Document interface {
    Index() string
    GetStringID() string
    GetTypeMapping() *types.TypeMapping
}

type PointerDocument[T any] interface {
    *T
    Document
}
```

**特点**：
- Go 1.27 泛型方法支持，类型安全的 API
- 批量操作支持，可配置统计信息
- 索引管理（创建/获取/删除）
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

func (p *Product) GetTypeMapping() *types.TypeMapping {
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

// 创建ElasticsearchX实例
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

### 10. errorx

**功能**：提供统一的错误处理机制，增强错误信息的可读性和可追踪性。

**核心功能**：
- 携带包/操作/详细信息/原始错误的上下文错误类型
- `New(err, pkg, op string, cause error) error` - 创建带上下文的错误
- `NewWithDetails(err, pkg, op, details string, cause error) error` - 创建带详细信息的错误
- `Is(err, target error) bool` - 错误链判定
- `In(err error, errs ...error) bool` - 错误是否属于给定集合之一
- 通过 `errors.Is` / `errors.As` 与标准库无缝衔接（支持 `Unwrap`）

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

### 11. singleflightx

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

### 12. breaker

**功能**：实现 Google SRE 自适应熔断器模式，用于保护服务免受级联故障影响。

**核心接口**：
- `Do(ctx context.Context, req func(ctx context.Context) error) error` - 执行请求
- `DoWithAcceptable(ctx context.Context, req func(ctx context.Context) error, acceptable func(err error) bool) error` - 执行请求，自定义成功判定
- `DoWithFallback(ctx context.Context, req func(ctx context.Context) error, fallback func(err error) error) error` - 执行请求，支持降级函数
- `Allow() error` - 准入判定（配合 `MarkSuccess`/`MarkFailure` 实现 Before/After 拦截器）
- `MarkSuccess()` / `MarkFailure()` - 手动记录请求结果
- `GetMetrics() (total, accepts int64, rate float64)` - 获取当前窗口指标
- `Reset()` - 手动重置熔断器

**配置选项**：
- `WithK(k float64)` - Google SRE 算法的 K 值（默认 1.5，越大越敏感）
- `WithProtection(p int)` - 保护请求数（默认 5，确保至少放行这么多请求）
- `WithWindow(d time.Duration)` - 滑动窗口时长（默认 10s）
- `WithBuckets(b int)` - 桶数量（默认 40）
- `WithName(name string)` - 熔断器名称（用于监控与日志）
- `WithRejectCallback(fn func())` - 拒绝回调（可用于打点/告警）

**特点**：
- 基于 Google SRE 自适应熔断算法
- 支持 Context 传播
- 支持自定义降级函数
- 支持 Before/After 式准入（`Allow` + `MarkSuccess`/`MarkFailure`），antx 即基于此实现
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
    breaker.WithRejectCallback(func() {
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

### 13. limiter

**功能**：基于 Redis 的令牌桶限流器，支持本地降级策略。

**核心接口**：
- `Allow(ctx context.Context) bool` - 允许单个请求
- `AllowN(ctx context.Context, now time.Time, n int) bool` - 允许 n 个请求

**特点**：
- 基于 Redis 实现分布式限流（Lua 脚本原子操作）
- Redis 故障时自动切换到本地限流（golang.org/x/time/rate 令牌桶）
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

### 14. monitor

**功能**：Prometheus 指标监控中间件，用于收集 HTTP 请求指标。

**核心接口**：
- `NewMetricsMiddleware(config MetricsConfig, reg prometheus.Registerer) (*MetricsMiddleware, error)` - 创建中间件（可传自定义 Registry）
- `Handler(next http.Handler) http.Handler` - HTTP 中间件
- `Record(path string, statusCode int, duration float64)` - 手动记录指标
- `AddCustomCallback(callback func(*RequestMetrics))` - 添加自定义回调

**配置**：
- `MetricsConfig{Namespace, Subsystem string, CustomCollectors []prometheus.Collector}` - 命名空间、子系统、自定义指标收集器
- `RequestMetrics{Path string, Status string, Duration float64, StatusCode int}` - 回调载荷

**收集的指标**：
- `http_requests_total` - 请求总数（按路径和状态码分组，labels: path, status）
- `http_request_duration_seconds` - 请求耗时直方图（labels: path）

**特点**：
- 支持自定义命名空间和子系统
- 支持自定义指标收集器、自定义 Registry
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
    log.Printf("Request: %s %s %.2fs", m.Path, m.Status, m.Duration)
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

reg := prometheus.NewRegistry()

// 创建监控中间件
mw, err := monitor.NewMetricsMiddleware(monitor.MetricsConfig{
		Namespace: "ginadapter",
	}, reg)

r := gin.Default()
r.GET("/", GinMetricsMiddleware(mw), func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
})
```

### 15. rabbitmq

**功能**：RabbitMQ 生产者和消费者封装，提供简洁的消息发布和订阅接口。适用于传统队列场景；如果你需要基于 RabbitMQ 的泛型事件总线，请使用 **[event](#7-event)** 组件的 `event/driver/rabbitmq` 驱动。

**核心接口（Producer）**：
- `Produce(exchange string, routeKey string, msg []byte) error` - 发送消息

**核心接口（Consumer）**：
- `Start()` - 启动消费者
- `Stop()` - 停止消费者

**配置选项**：
- `RabbitConfig` - RabbitMQ 连接配置（Username/Password/Host/Port/VHost）
- `RabbitProducerConfig` - 生产者配置（含 ContentType）
- `RabbitConsumerConfig` - 消费者配置（含 ListenerQueues、ConsumerConfig）
- 额外的 `QueueConfig`、`ExchangeConfig`、`BindConfig` 用于管理端声明队列/交换机/绑定（见 `rabbitmq/admin.go`）

**特点**：
- 支持自动确认和手动确认模式
- 支持多个监听队列（ListenerQueues）
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
| github.com/skudra-go/ant | ant ORM（配合 db/antx、cachedb/antc 使用） |
| github.com/coocood/freecache | 本地缓存实现 |
| github.com/redis/go-redis/v9 | Redis 客户端 |
| github.com/elastic/go-elasticsearch/v9 | Elasticsearch 客户端 |
| gorm.io/gorm | ORM 框架 |
| github.com/rabbitmq/amqp091-go | RabbitMQ 客户端 |
| go.opentelemetry.io/otel | OpenTelemetry 链路追踪（trace、gormx） |
| golang.org/x/time | 本地限流器实现（limiter 降级） |
| github.com/prometheus/client_golang | Prometheus 指标收集（monitor） |
| github.com/LouYuanbo1/go-burn | 日志库 |

## ✨ 亮点特性

1. **Go 1.27 泛型方法**：利用 Go 1.27 引入的泛型方法特性进行深度重构，提供更加优雅和类型安全的 API
2. **模块化设计**：各组件独立封装，可单独使用
3. **二级缓存策略**：索引缓存 + 主键缓存 + 安全间隙，兼顾性能与一致性
4. **全链路熔断保护**：从 GORM（gormx）、ant ORM（antx）到 Redis（cache 驱动）、缓存层，均内置 Google SRE 熔断器，防止雪崩
5. **泛型事件总线**：类型安全的发布/订阅，支持本地总线与 Redis/RabbitMQ 分布式驱动
6. **可观测性**：Prometheus 指标监控（monitor）与 OpenTelemetry 链路追踪（trace + gormx 自动埋点）
7. **简洁易用**：提供直观的 API 接口，简化常见操作
8. **配置灵活**：支持详细的配置选项，满足不同场景需求
9. **错误处理**：统一的错误处理机制，增强错误信息的可读性
10. **事务支持**：GORM 与 ant ORM 均提供事务透传支持，确保数据一致性

## 🤝 贡献指南

欢迎报告问题或提出建议！

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

**注意**：本库基于 Go 1.27 开发，使用了 Go 1.27 引入的泛型方法特性进行重构。您需要使用 Go 1.27 开发版本才能编译和使用本库。