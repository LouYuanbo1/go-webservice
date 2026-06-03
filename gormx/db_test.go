package gormx

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------- 模型定义 ----------
type User struct {
	ID        uint64    `gorm:"primaryKey" redis:"id"`
	Name      string    `gorm:"not null" redis:"name"`
	Gender    int       `gorm:"default:0"`
	Age       int       `gorm:"default:0"`
	Email     string    `gorm:"not null" redis:"email"`
	Phone     string    `gorm:"not null" redis:"phone"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
}

func (u *User) GetID() uint64      { return u.ID }
func (u *User) PrimaryKey() string { return "id" }
func (u *User) TableName() string  { return "users" }

// ---------- 测试辅助函数 ----------

// setupTestDB 为每个测试创建一个全新的 SQLite 内存数据库，并完成自动迁移。
// 返回数据库实例和清理函数。
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return db, cleanup
}

// prepareSampleData 准备 500 条固定样本数据，传入当前测试的数据库
func prepareSampleData(t *testing.T, db *gorm.DB) {
	t.Helper()

	xdb := NewDB(db)

	users := make([]*User, 0, 500)
	for i := 1; i <= 500; i++ {
		users = append(users, &User{
			Name:   "testCreate" + strconv.Itoa(i),
			Gender: i % 2,
			Age:    10 + i%50,
			Email:  "testCreate" + strconv.Itoa(i) + "@example.com",
			Phone:  strconv.FormatUint(uint64(i)+10000000000, 10),
		})
	}
	err := xdb.CreateInBatches(context.Background(), users, 100)
	assert.NoError(t, err)
}

// ---------- 测试用例 ----------

func TestCreate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	ctx := context.Background()

	// 单条创建
	user := &User{
		Name:   "testCreate1",
		Gender: 1,
		Age:    11,
		Email:  "testCreate1@example.com",
		Phone:  "10000000001",
	}
	err := xdb.Create(ctx, user)
	assert.NoError(t, err)

	// 批量创建
	users := make([]*User, 0, 499)
	for i := 2; i < 501; i++ {
		users = append(users, &User{
			Name:   "testCreate" + strconv.Itoa(i),
			Gender: i % 2,
			Age:    10 + i%50,
			Email:  "testCreate" + strconv.Itoa(i) + "@example.com",
			Phone:  strconv.FormatUint(uint64(i)+10000000000, 10),
		})
	}
	err = xdb.CreateInBatches(ctx, users, 100)
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db) // 准备 500 条数据

	xdb := NewDB(db)
	ctx := context.Background()

	// GetByID
	userByID := &User{}
	err := xdb.GetByID(ctx, userByID, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), userByID.ID)
	assert.Equal(t, "testCreate1", userByID.Name)
	assert.Equal(t, 1, userByID.Gender)
	assert.Equal(t, 11, userByID.Age)
	assert.Equal(t, "testCreate1@example.com", userByID.Email)
	assert.Equal(t, "10000000001", userByID.Phone)

	// GetByStructFilter
	userByStruct := &User{}
	err = xdb.GetByStructFilter(ctx, userByStruct, &User{Name: "testCreate2"})
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), userByStruct.ID)
	assert.Equal(t, "testCreate2", userByStruct.Name)
	assert.Equal(t, 0, userByStruct.Gender)
	assert.Equal(t, 12, userByStruct.Age)
	assert.Equal(t, "testCreate2@example.com", userByStruct.Email)
	assert.Equal(t, "10000000002", userByStruct.Phone)

}

func TestFind(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db)

	xdb := NewDB(db)
	ctx := context.Background()

	// FindByIDs
	usersByID := make([]*User, 0)
	err := xdb.FindByIDs(ctx, &usersByID, []uint64{1, 2, 3})
	assert.NoError(t, err)
	assert.Len(t, usersByID, 3)
	for i, user := range usersByID {
		id := uint64(i + 1)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
		assert.Equal(t, int(id%2), user.Gender)
		assert.Equal(t, int(10+id%50), user.Age)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id))+"@example.com", user.Email)
		assert.Equal(t, strconv.FormatUint(id+10000000000, 10), user.Phone)
	}

	// FindByStructFilter
	usersByStruct := make([]User, 0)
	err = xdb.FindByStructFilter(ctx, &usersByStruct, &User{Age: 10})
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 10, user.Age)
	}

	// FindByPage
	usersByPage := make([]*User, 0)
	err = xdb.FindByPage(ctx, &usersByPage, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, usersByPage, 10)
	for i, user := range usersByPage {
		id := uint64(i + 1)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}

	// FindByCursor
	usersByCursor := make([]*User, 0)
	err = xdb.FindByCursor(ctx, &usersByCursor, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, usersByCursor, 10)
	for i, user := range usersByCursor {
		id := uint64(i + 11)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}

	// FindInBatches
	var batchCountFind int
	var totalUsersFind int
	err = xdb.FindInBatches(ctx, 100, func(ctx context.Context, tx *DB, batch int, users *[]User) error {
		batchCountFind++
		totalUsersFind += len(*users)
		// 验证每批次数据量不超过 batchSize
		assert.LessOrEqual(t, len(*users), 100)
		// 验证用户ID按顺序递增
		for i, user := range *users {
			expectedID := uint64((batch-1)*100 + i + 1)
			assert.Equal(t, expectedID, user.ID)
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, batchCountFind) // 500条数据，每批100条，共5批
	assert.Equal(t, 500, totalUsersFind)
}

func TestUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db)

	xdb := NewDB(db)
	ctx := context.Background()

	// 更新单条记录（按主键）
	userUpdate := &User{
		ID:     1,
		Name:   "testUpdate1",
		Gender: 1,
		Age:    11,
		Email:  "testUpdate1@example.com",
		Phone:  "10000000001",
	}
	err := xdb.Update(ctx, userUpdate)
	assert.NoError(t, err)

	userByID := &User{}
	err = xdb.GetByID(ctx, userByID, 1)
	assert.NoError(t, err)
	assert.Equal(t, "testUpdate1", userByID.Name)
	assert.Equal(t, 1, userByID.Gender)
	assert.Equal(t, 11, userByID.Age)
	assert.Equal(t, "testUpdate1@example.com", userByID.Email)
	assert.Equal(t, "10000000001", userByID.Phone)

	// 按结构体条件更新
	structFilter := &User{Age: 11}
	structUpdate := &User{Email: "testUpdateByAge11@example.com"}
	err = xdb.UpdatesByStructFilter(ctx, structFilter, structUpdate)
	assert.NoError(t, err)

	usersByStruct := make([]User, 0)
	err = xdb.FindByStructFilter(ctx, &usersByStruct, structFilter)
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 11, user.Age)
		assert.Equal(t, "testUpdateByAge11@example.com", user.Email)
	}

}

func TestDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db)

	xdb := NewDB(db)
	ctx := context.Background()

	// 按主键删除
	err := xdb.DeleteByID[User](ctx, 1)
	assert.NoError(t, err)

	// 按多个主键删除（id=2,3 存在，应该成功）
	err = xdb.DeleteByIDs[User](ctx, 2, 3)
	assert.NoError(t, err)

	// 按结构体条件删除
	err = xdb.DeleteByStructFilter(ctx, &User{Age: 11})
	assert.NoError(t, err)

}

func TestTransaction(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db)

	xdb := NewDB(db)
	ctx := context.Background()

	err := xdb.Transaction(ctx, func(ctx context.Context, tx *DB) error {
		var user User
		tx.GetByID(ctx, &user, 1)
		user.Age = 100
		tx.Update(ctx, &user)
		return nil
	})
	assert.NoError(t, err)
	var user User
	err = xdb.GetByID(ctx, &user, 1)
	assert.NoError(t, err)
	assert.Equal(t, 100, user.Age)
}

// TestNewDB_DefaultBreaker 测试默认熔断器创建
func TestNewDB_DefaultBreaker(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	db := NewDB(gdb)
	assert.NotNil(t, db)
	assert.NotNil(t, db.brk)

	// 验证默认 acceptable 函数
	assert.True(t, db.acceptable(gorm.ErrRecordNotFound))
	assert.True(t, db.acceptable(gorm.ErrInvalidTransaction))
	assert.False(t, db.acceptable(errors.New("other error")))
}

// TestNewDB_WithCustomBreaker 测试自定义熔断器注入
func TestNewDB_WithCustomBreaker(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	customBrk := breaker.NewBreaker(
		breaker.WithName("custom-gormx-breaker"),
		breaker.WithK(2.0),
		breaker.WithProtection(100),
	)

	db := NewDB(gdb, WithBreaker(customBrk))
	assert.NotNil(t, db)
	assert.Equal(t, customBrk, db.brk)
}

// TestNewDB_WithCustomAcceptable 测试自定义 acceptable 函数
func TestNewDB_WithCustomAcceptable(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	customErr := errors.New("custom acceptable error")
	customAcceptable := func(err error) bool {
		return err == customErr
	}

	db := NewDB(gdb, WithAcceptable(customAcceptable))
	assert.NotNil(t, db)
	assert.True(t, db.acceptable(customErr))
	assert.False(t, db.acceptable(gorm.ErrRecordNotFound))
}

// TestDB_BreakerMetrics 测试熔断器指标收集
func TestDB_BreakerMetrics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 执行成功的操作
	xdb := NewDB(db)
	err := xdb.Create(context.Background(), &User{Name: "test"})
	assert.NoError(t, err)

	// 检查熔断器指标
	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

type FailedModel struct {
	Name string `gorm:"column:name"`
}

// TestDB_BreakerFailure 测试熔断器记录失败
func TestDB_BreakerFailure(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	// 执行失败的操作（表不存在）
	err := xdb.Create(context.Background(), &FailedModel{Name: "test"})
	assert.Error(t, err)

	// 检查熔断器指标
	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(0), accepts)
	assert.Equal(t, 0.0, rate)
}

// TestDB_BreakerAcceptableError 测试可接受错误不触发熔断计数
func TestDB_BreakerAcceptableError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	assert.NotNil(t, db)

	// 使用 Find 方法触发 gorm.ErrRecordNotFound
	var user User
	err := xdb.brk.DoWithAcceptable(context.Background(), func(ctx context.Context) error {
		return db.WithContext(ctx).First(&user).Error
	}, xdb.acceptable)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	// 检查熔断器指标 - 可接受错误应被视为成功
	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

// TestDB_BreakerReset 测试熔断器重置
func TestDB_BreakerReset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	assert.NotNil(t, xdb)

	// 执行失败操作
	err := xdb.Create(context.Background(), &FailedModel{Name: "test"})
	assert.Error(t, err)

	// 检查熔断器指标
	total, accepts, _ := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(0), accepts)

	// 重置熔断器
	xdb.brk.Reset()

	// 检查指标是否被重置
	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(0), total)
	assert.Equal(t, int64(0), accepts)
	assert.Equal(t, float64(0), rate)
}
