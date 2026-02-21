# Go WebService 封装库

这是一个功能丰富的 Go 语言封装库，为 Web 服务开发提供了多种实用工具组件。该库旨在简化常见的 Web 开发任务，提高开发效率和代码质量。

## 📁 组件结构

该项目包含以下核心组件：

| 组件 | 目录 | 功能描述 |
|------|------|----------|
| **cryptutil** | `cryptutil/` | 密码加密与验证工具 |
| **excelutil** | `excelutil/` | Excel 文件处理工具 |
| **ginutil** | `ginutil/` | Gin 框架辅助工具，特别是 multipart 表单处理 |
| **gormx** | `gormx/` | GORM ORM 框架的增强封装 |
| **imgutil** | `imgutil/` | 图像处理工具 |
| **localcache** | `localcache/` | 本地缓存工具 |
| **redisx** | `redisx/` | Redis 缓存工具的封装 |

## 🚀 安装

使用 Go Modules 安装：

```bash
go get github.com/LouYuanbo1/go-webservice
```

## 📦 组件详情

### 1. cryptutil

**功能**：提供密码加密和验证功能，基于 bcrypt 算法。

**核心接口**：
- `Encrypt(secret string, opts ...options.CostOption) ([]byte, error)` - 加密密码
- `CheckSecret(secret string, hashedSecret []byte) error` - 验证密码

**配置选项**：
- `DefaultCost` - 默认加密成本

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/cryptutil"
    "github.com/LouYuanbo1/go-webservice/cryptutil/config"
)

// 创建加密工具
cfg := config.CryptUtilConfig{
    DefaultCost: 10,
}
crypto := cryptutil.NewCryptUtil(cfg)

// 加密密码
hashedPassword, err := crypto.Encrypt("your-password")

// 验证密码
err = crypto.CheckSecret("your-password", hashedPassword)
```

### 2. excelutil

**excelize链接**：[excelize/v2](https://github.com/xuri/excelize/v2)

**功能**：提供 Excel 文件处理功能，基于 excelize 库。

**核心函数**：
- `OpenReader(fileHeader multipart.FileHeader) (*excelize.File, error)` - 从文件头打开 Excel 文件

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/excelutil/internal"
)

// 处理上传的 Excel 文件
file, err := internal.OpenReader(fileHeader)
if err != nil {
    // 处理错误
}
// 使用 file 进行 Excel 操作
```

### 3. ginutil

**gin链接**：[gin](https://github.com/gin-gonic/gin)

**功能**：Gin 框架的辅助工具，特别是增强的 multipart 表单处理。

**核心函数**：
- `BindMultipart[T any](gctx *gin.Context, obj T) error` - 解析 multipart/form-data 请求，将文本字段和文件字段绑定到结构体,支持嵌套结构体,支持文件字段和普通字段混合,支持索引格式（bracket 或 dot 格式）,遵循 Go 嵌入规范。试图解决gin框架在处理 multipart 表单时的一些限制,例如无法绑定数组类型的表单到slice上

**特点**：
- 支持嵌套结构体
- 支持文件字段和普通字段混合
- 支持索引格式（bracket 或 dot 格式）
- 遵循 Go 嵌入规范

**使用示例**：

```go
import (
    "github.com/LouYuanbo1/go-webservice/ginutil/multipart"
    "github.com/gin-gonic/gin"
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

### 4. gormx

**gorm链接**：[gorm](https://github.com/go-gorm/gorm)

**功能**：GORM ORM 框架的增强封装，提供更便捷的数据库操作方法。

**核心接口**：
- `Create(ctx context.Context, model PT, opts ...options.ConflictOption) error` - 创建记录
- `GetByID(ctx context.Context, id ID) (PT, error)` - 根据 ID 获取记录
- `FindByStructFilter(ctx context.Context, filter PT, opts ...options.OrderOption) ([]PT, error)` - 根据结构体过滤器查找记录
- `Update(ctx context.Context, updateData PT) error` - 更新记录
- `DeleteByID(ctx context.Context, id ID) error` - 根据 ID 删除记录

**特点**：
- 泛型支持，类型安全
- 丰富的查询选项
- 事务支持
- 批量操作支持

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

// 创建 GormX 实例
db, _ := gorm.Open(...) // 初始化 GORM 数据库连接
gormX := gormx.NewGormX[User, uint, *User](db)

// 创建记录
user := &User{Name: "John", Age: 30}
err := gormX.Create(context.Background(), user)

// 根据 ID 获取记录
user, err := gormX.GetByID(context.Background(), 1)

// 查找记录
users, err := gormX.FindByStructFilter(context.Background(), &User{Age: 30})
```

### 5. imgutil

**imaging链接**：[imaging](https://github.com/disintegration/imaging)

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
    "github.com/LouYuanbo1/go-webservice/imgutil/config"
)

// 创建图像工具
cfg := config.ImgUtilConfig{}
imgUtil := imgutil.NewImgUtil(cfg)

// 加载图像
img, err := imgUtil.Load("path/to/image.jpg")

// 生成缩略图
thumb := imgUtil.Thumbnail(img)

// 保存图像
err = imgUtil.Save(thumb, "path/to/thumbnail.jpg")
```

### 6. localcache

**ristretto链接**：[ristretto/v2](https://github.com/dgraph-io/ristretto/v2)

**功能**：本地缓存工具，基于 ristretto 缓存库。

**核心接口**：
- `SetWithTTL(ctx context.Context, key string, value T, opts ...options.TTLOption) bool` - 设置缓存值
- `Get(ctx context.Context, key string) (T, bool)` - 获取缓存值
- `GetPointer(ctx context.Context, key string) (*T, bool)` - 获取缓存值指针
- `Del(ctx context.Context, key string)` - 删除缓存值

**特点**：
- 泛型支持
- TTL 支持
- 类型安全

**使用示例**：

```go
import (
    "context"
    "github.com/LouYuanbo1/go-webservice/localcache"
    "github.com/LouYuanbo1/go-webservice/localcache/config"
)

// 创建本地缓存
cfg := &config.LocalConfig{}
cache, err := localcache.NewLocalCache[string](cfg)

// 设置缓存
cache.SetWithTTL(context.Background(), "key", "value")

// 获取缓存
value, exists := cache.Get(context.Background(), "key")
if exists {
    // 使用 value
}
```

### 7. redisx

**redis链接**：[redis/v9](https://github.com/redis/go-redis/v9)

**功能**：Redis 缓存工具的封装，提供更便捷的 Redis 操作方法。

**核心接口**：
- `SetWithTTL(ctx context.Context, key string, value T, opts ...options.TTLOption) error` - 设置缓存值
- `Get(ctx context.Context, key string) (T, error)` - 获取缓存值
- `HGet(ctx context.Context, key string, field string) (string, error)` - 获取哈希字段值
- `Del(ctx context.Context, key string) error` - 删除缓存值
- `Acquire(ctx context.Context, key string, expire time.Duration) (string, bool, error)` - 获取分布式锁
- `Release(ctx context.Context, key, lockID string) error` - 释放分布式锁

**特点**：
- 泛型支持
- 自动序列化和反序列化
- TTL 支持
- 分布式锁支持

**使用示例**：

```go
import (
    "context"
    "time"
    "github.com/LouYuanbo1/go-webservice/redisx"
    "github.com/redis/go-redis/v9"
)

// 创建 Redis 客户端
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建 RedisX 实例
redisX := redisx.NewRedisX[string](client, time.Hour)

// 设置缓存
redisX.SetWithTTL(context.Background(), "key", "value")

// 获取缓存
value, err := redisX.Get(context.Background(), "key")

// 获取分布式锁
lockID, acquired, err := redisX.Acquire(context.Background(), "lock-key", time.Minute)
if acquired {
    defer redisX.Release(context.Background(), "lock-key", lockID)
    // 执行需要锁定的操作
}
```

## 🛠️ 技术栈

| 依赖 | 版本 | 用途 |
|------|------|------|
| github.com/dgraph-io/ristretto/v2 | v2.4.0 | 本地缓存实现 |
| github.com/disintegration/imaging | v1.6.2 | 图像处理 |
| github.com/gin-gonic/gin | v1.11.0 | Web 框架 |
| github.com/go-playground/form/v4 | v4.3.0 | 表单解析 |
| github.com/google/uuid | v1.6.0 | UUID 生成 |
| github.com/redis/go-redis/v9 | v9.17.3 | Redis 客户端 |
| github.com/xuri/excelize/v2 | v2.10.0 | Excel 文件处理 |
| golang.org/x/crypto | v0.47.0 | 密码加密 |
| gorm.io/gorm | v1.31.1 | ORM 框架 |

## ✨ 亮点特性

1. **泛型支持**：充分利用 Go 1.18+ 的泛型特性，提供类型安全的 API
2. **模块化设计**：各组件独立封装，可单独使用
3. **简洁易用**：提供直观的 API 接口，简化常见操作
4. **功能丰富**：涵盖 Web 开发中常见的多种工具需求
5. **配置灵活**：支持详细的配置选项，满足不同场景需求
6. **性能优化**：基于成熟的第三方库，提供高性能实现

## 📝 使用指南

### 安装依赖

```bash
go mod tidy
```

### 导入组件

根据需要导入相应的组件：

```go
import (
    "github.com/LouYuanbo1/go-webservice/cryptutil"
    "github.com/LouYuanbo1/go-webservice/ginutil/multipart"
    "github.com/LouYuanbo1/go-webservice/gormx"
    // 其他组件...
)
```

### 初始化组件

每个组件都有自己的初始化方法，通常需要提供配置选项：

```go
// 初始化 cryptutil
crypto := cryptutil.NewCryptUtil(config)

// 初始化 gormx
gormX := gormx.NewGormX[Model, ID, *Model](db)

// 初始化 redisx
redisX := redisx.NewRedisX[T](client, defaultTTL)
```

## 🤝 贡献指南

欢迎报告问题或提出建议！

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

**注意**：本库基于 Go 1.25.4 开发，建议使用 Go 1.18+ 版本以获得完整的泛型支持。# go-webservice
对Go语言常用Web组件的简单封装
