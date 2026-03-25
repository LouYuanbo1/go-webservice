# Go WebService 封装库

这是一个功能丰富的 Go 语言封装库，为 Web 服务开发提供了多种实用工具组件。该库旨在简化常见的 Web 开发任务，提高开发效率和代码质量。

## 📁 组件结构

该项目包含以下核心组件：

| 组件 | 目录 | 功能描述 |
|------|------|----------|
| **cache** | `cache/` | 缓存组件，支持本地缓存和Redis缓存 |
| **cryptutil** | `cryptutil/` | 密码加密与验证工具 |
| **elasticsearchx** | `elasticsearchx/` | Elasticsearch操作封装 |
| **errorx** | `errorx/` | 错误处理工具 |
| **ginutil** | `ginutil/` | Gin 框架辅助工具，特别是 multipart 表单处理 |
| **gormc** | `gormc/` | GORM缓存连接 |
| **gormx** | `gormx/` | GORM ORM 框架的增强封装 |
| **hashutil** | `hashutil/` | 哈希工具 |
| **imgutil** | `imgutil/` | 图像处理工具 |
| **singleflightx** | `singleflightx/` | 单飞工具，避免缓存击穿 |

## 🚀 安装

使用 Go Modules 安装：

```bash
go get github.com/LouYuanbo1/go-webservice/cache
go get github.com/LouYuanbo1/go-webservice/cryptutil
go get github.com/LouYuanbo1/go-webservice/elasticsearchx
......
```

## 📦 组件详情

### 1. cache

**功能**：提供统一的缓存接口，支持本地缓存（基于ristretto）和Redis缓存。

**核心接口**：
- `Set(ctx context.Context, key string, val any, ttl time.Duration) error` - 设置缓存
- `Get(ctx context.Context, key string, val any) error` - 获取缓存
- `Take(ctx context.Context, val any, key string, query func(val any) error, ttl time.Duration) error` - 缓存穿透处理
- `Del(ctx context.Context, keys ...string) error` - 删除缓存

**特点**：
- 统一的缓存接口
- 支持多种缓存驱动
- 支持键前缀配置

**使用示例**：

```go
import (
    "context"
    "time"
    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cache/drivers/local"
    "github.com/LouYuanbo1/go-webservice/cache/drivers/redis"
)

// 创建本地缓存驱动
localConfig := &local.Config{
    NumCounters: 10000,
    MaxCost:     1000,
    BufferItems: 64,
}
localDriver := local.NewDriver(localConfig)

// 或创建Redis缓存驱动
redisConfig := &redis.Config{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    DB:       0,
}
redisDriver := redis.NewDriver(redisConfig)

// 打开缓存客户端
client, err := cache.Open(localDriver, cache.WithPrefix("app:"))

// 设置缓存
err = client.Set(context.Background(), "key", "value", time.Hour)

// 获取缓存
var value string
err = client.Get(context.Background(), "key", &value)

// 使用Take方法避免缓存穿透
var result string
err = client.Take(context.Background(), &result, "key", func(val any) error {
    // 从数据库获取数据
    *(val.(*string)) = "data from db"
    return nil
}, time.Hour)
```

### 2. cryptutil

**功能**：提供密码加密和验证功能，基于 bcrypt 算法。

**核心接口**：
- `Encrypt(secret string, opts ...CostOption) ([]byte, error)` - 加密密码
- `CheckSecret(secret string, hashedSecret []byte) error` - 验证密码

**配置选项**：
- `DefaultCost` - 默认加密成本

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/cryptutil"
)

// 创建加密工具
crypto := cryptutil.NewCryptUtil(cryptutil.Config{
    DefaultCost: 10,
})

// 加密密码
hashedPassword, err := crypto.Encrypt("your-password")

// 验证密码
err = crypto.CheckSecret("your-password", hashedPassword)
```

### 3. elasticsearchx

**功能**：Elasticsearch操作封装，提供文档CRUD、批量操作等功能。

**核心接口**：
- `CreateIndex(ctx context.Context, doc PT) error` - 创建索引
- `IndexDoc(ctx context.Context, doc PT) error` - 索引文档
- `BulkIndexDocs(ctx context.Context, docs []PT, opts ...BulkOption) error` - 批量索引文档
- `GetDoc(ctx context.Context, index string, id string) (PT, error)` - 获取文档
- `FindDocsByPages(ctx context.Context, index string, page, size int) ([]PT, error)` - 分页查询文档
- `UpdateDoc(ctx context.Context, doc PT) error` - 更新文档
- `DeleteDoc(ctx context.Context, index string, id string) error` - 删除文档
- `BulkDeleteDocs(ctx context.Context, index string, ids []string, opts ...BulkOption) error` - 批量删除文档

**特点**：
- 泛型支持
- 批量操作支持
- 索引管理
- 详细的错误处理

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/elasticsearchx"
    "github.com/elastic/go-elasticsearch/v9"
)

type Product struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
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

// 创建ElasticsearchX实例
es := elasticsearchx.NewElasticsearchX[Product, *Product](client, &elasticsearchx.Config{
    BulkIndexer: elasticsearchx.BulkIndexerConfig{
        Stats: true,
    },
})

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
es.BulkIndexDocs(context.Background(), docs)

// 查询文档
result, err := es.GetDoc(context.Background(), "products", "1")

// 分页查询
docs, err := es.FindDocsByPages(context.Background(), "products", 1, 10)
```

### 4. errorx

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

### 5. ginutil

**功能**：Gin 框架的辅助工具，特别是增强的 multipart 表单处理。

**核心函数**：
- `BindMultipart[T any](gctx *gin.Context, obj T) error` - 解析 multipart/form-data 请求，将文本字段和文件字段绑定到结构体

**特点**：
- 支持嵌套结构体
- 支持文件字段和普通字段混合
- 支持索引格式（bracket 或 dot 格式）
- 遵循 Go 嵌入规范
- 解决 Gin 框架在处理 multipart 表单时的一些限制

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/ginutil/multipart"
    "github.com/gin-gonic/gin"
    "mime/multipart"
)

type Form struct {
    Name  string                `form:"name"`
    Email string                `form:"email"`
    File  *multipart.FileHeader `form:"file"`
}

func uploadHandler(c *gin.Context) {
    var form Form
    if err := multipart.BindMultipart(c, &form); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // 处理表单数据
}
```

### 6. gormc

**功能**：提供GORM连接的缓存支持，实现Cache-Aside模式。

**核心接口**：
- `GetCache(ctx context.Context, key string, val any) error` - 获取缓存数据
- `SetCache(ctx context.Context, key string, val any, opts ...TTLOption) error` - 设置缓存数据
- `DelCache(ctx context.Context, key ...string) error` - 删除缓存数据
- `Exec(ctx context.Context, exec ExecFn, keys ...string) error` - 执行数据库操作并删除缓存
- `Query(ctx context.Context, val any, key string, query QueryFn, opts ...TTLOption) error` - 查询数据并缓存
- `QueryIndex(ctx context.Context, val any, key string, keyer func(primary any) string, indexQuery QueryFn, primaryQuery PrimaryQueryFn, opts ...TTLOption) error` - 索引查询并缓存
- `ExecNoCache(ctx context.Context, exec ExecFn) error` - 执行数据库操作不缓存
- `QueryNoCache(ctx context.Context, val any, query QueryFn) error` - 查询数据不缓存

**泛型版本**：
- `TypedCacheDB` - 泛型版本的缓存数据库
- `GetTypedCacheSession[T any, ID comparable, PT gormx.PointerModel[T, ID]](tcdb *TypedCacheDB) TypedCacheSession[T, ID, PT]` - 获取泛型缓存会话
- `TypedCacheSession` - 泛型缓存会话，支持查询和索引查询并缓存

- `GetCache(ctx context.Context, key string, val PT) error` - 获取缓存数据
- `SetCache(ctx context.Context, key string, val PT, opts ...TTLOption) error` - 设置缓存数据
- `DelCache(ctx context.Context, key ...string) error` - 删除缓存数据
- `Exec(ctx context.Context, exec TypedExecFn[T, ID, PT], keys ...string) error` - 执行数据库操作并删除缓存
- `Query(ctx context.Context,val PT,key string,query TypedQueryFn[T, ID, PT],opts ...TTLOption) error` - 查询数据并缓存
- `QueryIndex(ctx context.Context,val PT,key string,keyer func(primary ID) string,indexQuery TypedIndexQueryFn[T, ID, PT],primaryQuery TypedPrimaryQueryFn[T, ID, PT],opts ...TTLOption) error` - 索引查询并缓存
- `ExecNoCache(ctx context.Context, exec TypedExecFn[T, ID, PT]) error` - 执行数据库操作不缓存
- `QueryNoCache(ctx context.Context, val PT, query TypedQueryFn[T, ID, PT]) error` - 查询数据不缓存
- `QueryRowsNoCache(ctx context.Context, val *[]PT, query TypedQueryRowsFn[T, ID, PT]) error` - 查询多条数据不缓存

**特点**：
- 实现 Cache-Aside 模式
- 支持事务
- 缓存一致性管理
- 泛型支持，提供类型安全的 API

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/gormc"
    "github.com/LouYuanbo1/go-webservice/gormx"
    "gorm.io/gorm"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	db, err := gormx.InitGorm(cfg.DB)
	if err != nil {
		panic(err)
	}
	gormxDB := gormx.NewTypedDB(db)
	redisCache := redis.NewDriver(cfg.Redis)
	cache, err := cache.Open(redisCache)
	if err != nil {
		panic(err)
	}
	gormcDB := gormc.NewTypedCacheDB(gormxDB, cache, &gormc.Config{
		TTL:                                20 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
	})
	ctx := context.Background()

    // 执行数据库操作并删除缓存
    var user &User{
        ID: 1,
        Name: "Updated Name",
    }
    err := cachedDB.Exec(context.Background(), func(ctx context.Context, conn gormx.Conn) error {
        return conn.Update(ctx, &user)
    }, "user:1")

    // 查询数据并缓存
    var user User
    err := cachedConn.Query(context.Background(), &user, "user:1", func(ctx context.Context, conn gormx.Conn, val any) error {
        return conn.GetByID(ctx, val, 1)
    })
}
```

**泛型版本使用示例**：

```go
package main

import (
	"context"
	"fmt"
	"playground/config"
	"playground/model"
	"strconv"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/drivers/redis"
	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	db, err := gormx.InitGorm(cfg.DB)
	if err != nil {
		panic(err)
	}
	gormxDB := gormx.NewTypedDB(db)
	redisCache := redis.NewDriver(cfg.Redis)
	cache, err := cache.Open(redisCache)
	if err != nil {
		panic(err)
	}
	gormcDB := gormc.NewTypedCacheDB(gormxDB, cache, &gormc.Config{
		TTL:                                20 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
	})
	ctx := context.Background()
	userRepo := gormx.GetSession[model.User, uint64](gormxDB)
	userRepo.Create(ctx, &model.User{
		Name:  "test user",
		Email: "test user@test.com",
		Phone: fmt.Sprintf("%d", 13800000000),
	})
	userRepoCache := gormc.GetTypedCacheSession[model.User, uint64](gormcDB)
	execSliceFn := func(ctx context.Context, s gormx.TypedSession[model.User, uint64, *model.User]) error {
		users := make([]*model.User, 100)
		for i := range users {
			users[i] = &model.User{
				Name:  "test user" + strconv.Itoa(int(i)),
				Email: "test user" + strconv.Itoa(int(i)) + "@test.com",
				Phone: fmt.Sprintf("%d", 13800000000+i),
			}
		}
		s.CreateInBatches(ctx, users, 50)
		return nil
	}
	if err := userRepoCache.Exec(ctx, execSliceFn); err != nil {
		fmt.Println(err)
	}
	fmt.Println("done")
}
```

### 7. gormx

**功能**：GORM ORM 框架的增强封装，提供更便捷的数据库操作方法。

**核心接口**：
- `GetDBWithContext(ctx context.Context) *gorm.DB` - 获取上下文中的GORM数据库连接
- `Create(ctx context.Context, model any, opts ...ConflictOption) error` - 创建记录
- `CreateInBatches(ctx context.Context, models any, batchSize int, opts ...ConflictOption) error` - 批量创建记录
- `GetByID(ctx context.Context, dest any, id any) error` - 根据ID获取记录
- `GetByStructFilter(ctx context.Context, dest any, filter any) error` - 根据结构体过滤器获取记录
- `GetByMapFilter(ctx context.Context, dest any, filter map[string]any) error` - 根据map过滤器获取记录
- `FindByIDs(ctx context.Context, dest any, ids any, opts ...OrderOption) error` - 根据IDs查找记录
- `FindByStructFilter(ctx context.Context, dest any, filter any, opts ...OrderOption) error` - 根据结构体过滤器查找记录
- `FindByMapFilter(ctx context.Context, dest any, filter map[string]any, opts ...OrderOption) error` - 根据map过滤器查找记录
- `FindByPage(ctx context.Context, dest any, primaryKey string, page, pageSize int, opts ...OrderOption) error` - 分页查询
- `FindByCursor(ctx context.Context, dest any, primaryKey string, cursor any, limit int) error` - 根据游标分页查询
- `FindInBatches(ctx context.Context,dest any,batchSize int,callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,opts ...OrderOption) error` - 批量查询记录
- `FindInBatchesByStructFilter(ctx context.Context,dest any,filter any,batchSize int,callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,opts ...OrderOption) error` - 根据结构体过滤器批量查询记录
- `FindInBatchesByMapFilter(ctx context.Context,dest any,filter map[string]any,batchSize int,callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,opts ...OrderOption) error` - 根据map过滤器批量查询记录
- `Update context(ctx context.Context, updateData any) error` - 更新记录
- `UpdateByStructFilter(ctx context.Context, filter any, updateData any) error` - 根据结构体过滤器更新记录
- `UpdateByMapFilter(ctx context.Context, model any, filter map[string]any, updateData map[string]any) error` - 根据map过滤器更新记录
- `DeleteByID(ctx context.Context, model any, id any) error` - 根据ID删除记录
- `DeleteByIDs(ctx context.Context, model any, ids any) error` - 根据IDs删除记录
- `DeleteByStructFilter(ctx context.Context, model any, filter any) error` - 根据结构体过滤器删除记录
- `DeleteByMapFilter(ctx context.Context, model any, filter map[string]any) error` - 删除map过滤器删除记录


**泛型版本**：
- `TypedDB` - 泛型版本的数据库连接
- `GetSession[T any, ID comparable, PT PointerModel[T, ID]](tdb *TypedDB) TypedSession[T, ID, PT]` - 获取泛型会话
- `TypedSession[T any, ID comparable, PT PointerModel[T, ID]]` - 泛型会话接口

- `GetDBWithContext(ctx context.Context) *gorm.DB` - 获取上下文中的GORM数据库连接
- `Create(ctx context.Context, model PT, opts ...ConflictOption) error` - 创建记录
- `CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error` - 批量创建记录
- `GetByID(ctx context.Context, dest PT, id ID) error` - 根据ID获取记录
- `GetByStructFilter(ctx context.Context, dest PT, filter PT) error` - 根据结构体过滤器获取记录
- `GetByMapFilter(ctx context.Context, dest PT, filter map[string]any) error` - 根据map过滤器获取记录
- `FindByIDs(ctx context.Context, dest *[]PT, ids []ID, opts ...OrderOption) error` - 根据IDs查找记录
- `FindByStructFilter(ctx context.Context, dest *[]PT, filter PT, opts ...OrderOption) error` - 根据结构体过滤器查找记录
- `FindByMapFilter(ctx context.Context, dest *[]PT, filter map[string]any, opts ...OrderOption) error` - 根据映射过滤器查找记录
- `FindByPage(ctx context.Context, dest *[]PT, page, pageSize int, opts ...OrderOption) error` - 分页查询
- `FindByCursor(ctx context.Context, dest *[]PT, cursor ID, limit int) (newCursor ID, hasMore bool, err error)` - 根据游标分页查询
- `FindInBatches(ctx context.Context,batchSize int,callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...OrderOption) error` - 批量查询记录
- `FindInBatchesByStructFilter(ctx context.Context,filter PT,batchSize int,callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...OrderOption) error` - 根据结构体过滤器批量查询记录
- `FindInBatchesByMapFilter(ctx context.Context,filter map[string]any,batchSize int,callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...OrderOption) error` - 根据map过滤器批量查询记录
- `Update(ctx context.Context, updateData PT) error` - 更新记录
- `UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error` - 根据结构体过滤器更新记录
- `UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error` - 根据map过滤器更新记录
- `DeleteByID(ctx context.Context, id ID) error` - 根据ID删除记录
- `DeleteByIDs(ctx context.Context, ids ...ID) error` - 根据IDs删除记录
- `DeleteByStructFilter(ctx context.Context, filter PT) error` - 根据结构体过滤器删除记录
- `DeleteByMapFilter(ctx context.Context, filter map[string]any) error` - 根据map过滤器删除记录


**特点**：
- 丰富的查询方法
- 批量操作支持
- 事务支持
- 链式调用
- 泛型支持，提供类型安全的 API

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/gormx"
    "gorm.io/gorm"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// 创建GORM连接
db, _ := gorm.Open(...)
gormxDB := gormx.NewDB(db)

// 创建记录
user := &User{Name: "John", Age: 30}

// 根据ID获取记录
var user User
err := gormxDB.GetByID(context.Background(), &user, 1)

// 查找记录
var users []User
err := gormxDB.FindByStructFilter(context.Background(), &users, &User{Age: 30})

// 分页查询
err := gormxDB.FindByPage(context.Background(), &users, 1, 10)

// 事务操作
err := gormxDB.Transaction(context.Background(), func(ctx context.Context, s gormx.Session) error {
    // 执行数据库操作
    user := &User{Name: "John", Age: 30}
    if err := s.Create(ctx, user); err != nil {
        return err
    }
    user.Age = 31
    if err := s.Update(ctx, user); err != nil {
    // 其他操作...
    return nil
})
```

**泛型版本使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/gormx"
    "gorm.io/gorm"
)

// 定义模型
type User struct {
    ID   int64  `gorm:"primaryKey"`
    Name string
    Age  int
}

func (u *User) GetID() uint64 {
	return u.ID
}

func (u *User) PrimaryKey() string {
	return "id"
}

// 创建GORM连接
db, _ := gorm.Open(...)
typedDB := gormx.NewTypedDB(db)

// 获取泛型会话
userSession := gormx.GetSession[User, int64, *User](typedDB)

// 创建记录
user := &User{Name: "John", Age: 30}
err := userSession.Create(context.Background(), user)

// 根据ID获取记录
var user User
err := userSession.GetByID(context.Background(), &user, 1)

// 查找记录
var users []User
err := userSession.FindByStructFilter(context.Background(), &users, &User{Age: 30})

// 事务操作
err := typedDB.Transaction(context.Background(), func(ctx context.Context, txDB *gormx.TypedDB) error {
    // 在事务内获取会话
    txUserSession := gormx.GetSession[User, int64, *User](txDB)
    
    // 执行数据库操作
    user := &User{Name: "John", Age: 30}
    if err := txUserSession.Create(ctx, user); err != nil {
        return err
    }
    user.Age = 31
    if err := txUserSession.Update(ctx, user); err != nil {
        return err
    }
    // 其他操作...
    return nil
})
```

### 8. hashutil

**功能**：提供哈希工具，包括一致性哈希等。

**核心功能**：
- 一致性哈希
- 常用哈希函数

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/hashutil"
)

// 创建一致性哈希
ch := hashutil.NewConsistentHash(3, nil)

// 添加节点
ch.Add("node1")
ch.Add("node2")
ch.Add("node3")

// 查找节点
node := ch.Get("key")
```

### 9. imgutil

**功能**：图像处理工具，提供图像加载、缩略图生成、保存等功能。

**核心接口**：
- `Load(imgPath string) (image.Image, error)` - 加载图像
- `Thumbnail(img image.Image, opts ...options.TransformOption) image.Image` - 生成缩略图
- `Save(img image.Image, filename string, opts ...options.SaveOption) error` - 保存图像
- `Delete(imgPath string) error` - 删除图像

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/imgutil"
)

// 创建图像工具
imgUtil := imgutil.NewImgUtil(imgutil.Config{})

// 加载图像
img, err := imgUtil.Load("path/to/image.jpg")

// 生成缩略图
thumb := imgUtil.Thumbnail(img)

// 保存图像
err = imgUtil.Save(thumb, "path/to/thumbnail.jpg")

// 为图像路径添加时间戳
timestampPath := imgutil.WithUnixNanoTimestamp("path/to/image.jpg")
```

### 10. singleflightx

**功能**：提供单飞功能，避免缓存击穿。

**核心功能**：
- 并发请求合并
- 避免缓存击穿

**泛型版本**：
- `TypedSingleFlight[T any]` - 泛型版本的单飞接口
- `NewTypedSingleFlight[T any]() TypedSingleFlight[T]` - 创建泛型单飞实例

**特点**：
- 简单易用的API
- 有效防止缓存击穿
- 泛型支持，提供类型安全的操作

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)

// 创建单飞实例
sf := singleflightx.NewSingleFlight()

// 执行单飞操作
result, err := sf.Do(context.Background(), "key", func() (any, error) {
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

// 执行单飞操作
result, err := sf.Do(context.Background(), "user:1", func() (User, error) {
    // 执行耗时操作，如数据库查询
    return User{ID: 1, Name: "John", Age: 30}, nil
})

// 执行带额外信息的单飞操作
result, fresh, err := sf.DoEx(context.Background(), "user:1", func() (User, error) {
    // 执行耗时操作，如数据库查询
    return User{ID: 1, Name: "John", Age: 30}, nil
})
// fresh 表示结果是否是新计算的
```

## 🛠️ 技术栈

| 依赖 | 用途 |
|------|------|
| github.com/dgraph-io/ristretto/v2 | 本地缓存实现 |
| github.com/redis/go-redis/v9 | Redis 客户端 |
| golang.org/x/crypto | 密码加密 |
| github.com/elastic/go-elasticsearch/v9 | Elasticsearch客户端 |
| gorm.io/gorm | ORM框架 |
| github.com/gin-gonic/gin | Web框架 |
| github.com/go-playground/form/v4 | 表单解析 |
| github.com/disintegration/imaging | 图像处理 |

## ✨ 亮点特性

1. **泛型支持**：充分利用 Go 1.18+ 的泛型特性，提供类型安全的 API
2. **模块化设计**：各组件独立封装，可单独使用
3. **简洁易用**：提供直观的 API 接口，简化常见操作
4. **功能丰富**：涵盖 Web 开发中常见的多种工具需求
5. **配置灵活**：支持详细的配置选项，满足不同场景需求
6. **性能优化**：基于成熟的第三方库，提供高性能实现
7. **错误处理**：统一的错误处理机制，增强错误信息的可读性
8. **缓存策略**：实现 Cache-Aside 模式，提供缓存一致性管理
9. **批量操作**：支持批量处理，提高性能
10. **事务支持**：提供事务操作，确保数据一致性

## 📝 使用指南

### 安装依赖

```bash
go mod tidy
```

### 导入组件

根据需要导入相应的组件：

```go
import (
    "github.com/LouYuanbo1/go-webservice/cache"
    "github.com/LouYuanbo1/go-webservice/cryptutil"
    "github.com/LouYuanbo1/go-webservice/elasticsearchx"
    "github.com/LouYuanbo1/go-webservice/ginutil/multipart"
    "github.com/LouYuanbo1/go-webservice/gormc"
    "github.com/LouYuanbo1/go-webservice/gormx"
    "github.com/LouYuanbo1/go-webservice/hashutil"
    "github.com/LouYuanbo1/go-webservice/imgutil"
    "github.com/LouYuanbo1/go-webservice/singleflightx"
)
```

### 初始化组件

每个组件都有自己的初始化方法，通常需要提供配置选项：

```go
// 初始化 cache
localDriver := local.NewDriver(&local.Config{...})
cacheClient, _ := cache.Open(localDriver)

// 初始化 cryptutil
crypto := cryptutil.NewCryptUtil(cryptutil.Config{...})

// 初始化 gormx
db, _ := gorm.Open(...)
gormConn := gormx.NewConn(db)

// 初始化 gormc
cachedConn := gormc.NewConnWithCache(gormConn, cacheClient, &gormc.Config{...})
```

## 🤝 贡献指南

欢迎报告问题或提出建议！

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

**注意**：本库基于 Go 1.25 开发，建议使用 Go 1.25+ 版本以获得完整的功能支持，特别是泛型特性。