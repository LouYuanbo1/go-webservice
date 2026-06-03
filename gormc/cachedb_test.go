package gormc

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/driver/redis"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

func setupTestDB(t *testing.T) (*gorm.DB, *cache.Client, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err, "Failed to open SQLite database")

	s := miniredis.RunT(t)
	port, err := strconv.Atoi(s.Port())
	assert.NoError(t, err, "Failed to parse Redis port")

	redisConfig := &redis.Config{
		Host:     s.Host(),
		Port:     port,
		Password: "",
		DB:       0,
	}
	redisClient, err := redis.InitRedisClient(redisConfig)

	cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	assert.NoError(t, err, "Failed to open Redis cache")
	client := cache.NewClient(cacher)

	err = db.AutoMigrate(&User{})
	assert.NoError(t, err, "Failed to auto migrate database")

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return db, client, cleanup
}

func prepareSampleData(t *testing.T, db *gorm.DB, client *cache.Client) {
	t.Helper()

	xdb := gormx.NewDB(db)
	cdb := NewCacheDB(xdb, client, &Config{
		TTL:                                20 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
	})
	ctx := context.Background()

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

	execFn := func(ctx context.Context, db *gormx.DB) error {
		return db.CreateInBatches(ctx, users, 100)
	}
	err := cdb.ExecNoCache(ctx, execFn)
	assert.NoError(t, err, "Failed to create users in batches")
}

// ---------- 测试用例 ----------

func TestCreate(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	ctx := context.Background()

	// 单条创建
	user := &User{
		Name:   "testCreate1",
		Gender: 1,
		Age:    11,
		Email:  "testCreate1@example.com",
		Phone:  "10000000001",
	}
	err := cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, user)
	})
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
	err = cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.CreateInBatches(ctx, users, 100)
	})
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	ctx := context.Background()

	// GetByID
	userByID := &User{}
	err := cdb.QueryNoCache(ctx, userByID, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.GetByID(ctx, val, 1)
	})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), userByID.ID)
	assert.Equal(t, "testCreate1", userByID.Name)
	assert.Equal(t, 1, userByID.Gender)
	assert.Equal(t, 11, userByID.Age)
	assert.Equal(t, "testCreate1@example.com", userByID.Email)
	assert.Equal(t, "10000000001", userByID.Phone)

	// GetByStructFilter
	userByStruct := &User{}
	err = cdb.Query(ctx, "userByStruct", userByStruct, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.GetByStructFilter(ctx, val, &User{Name: "testCreate2"})
	})
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), userByStruct.ID)
	assert.Equal(t, "testCreate2", userByStruct.Name)
	assert.Equal(t, 0, userByStruct.Gender)
	assert.Equal(t, 12, userByStruct.Age)
	assert.Equal(t, "testCreate2@example.com", userByStruct.Email)
	assert.Equal(t, "10000000002", userByStruct.Phone)
}

func TestFind(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	ctx := context.Background()

	// FindByIDs
	usersByID := make([]User, 0)
	err := cdb.QueryRowsNoCache(ctx, &usersByID, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.FindByIDs(ctx, val, []uint64{1, 2, 3})
	})
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
	err = cdb.QueryRowsNoCache(ctx, &usersByStruct, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.FindByStructFilter(ctx, val, &User{Age: 10})
	})
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 10, user.Age)
	}

	// FindByPage
	usersByPage := make([]User, 0)
	err = cdb.QueryRowsNoCache(ctx, &usersByPage, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.FindByPage(ctx, val, 1, 10)
	})
	assert.NoError(t, err)
	assert.Len(t, usersByPage, 10)
	for i, user := range usersByPage {
		id := uint64(i + 1)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}

	// FindByCursor
	usersByCursor := make([]User, 0)
	err = cdb.QueryRowsNoCache(ctx, &usersByCursor, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.FindByCursor(ctx, val, 10, 10)
	})
	assert.NoError(t, err)
	assert.Len(t, usersByCursor, 10)
	for i, user := range usersByCursor {
		id := uint64(i + 11)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}
}

func TestUpdate(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
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
	err := cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Update(ctx, userUpdate)
	})
	assert.NoError(t, err)

	userByID := &User{}
	//不单独缓存,避免影响其他测试结果
	err = cdb.QueryNoCache(ctx, userByID, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.GetByID(ctx, val, 1)
	})
	assert.NoError(t, err)
	assert.Equal(t, "testUpdate1", userByID.Name)
	assert.Equal(t, 1, userByID.Gender)
	assert.Equal(t, 11, userByID.Age)
	assert.Equal(t, "testUpdate1@example.com", userByID.Email)
	assert.Equal(t, "10000000001", userByID.Phone)

	// 按结构体条件更新
	structFilter := &User{Age: 11}
	structUpdate := &User{Email: "testUpdateByAge11@example.com"}
	err = cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.UpdatesByStructFilter(ctx, structFilter, structUpdate)
	})
	assert.NoError(t, err)
	usersByStruct := make([]User, 0)
	err = cdb.QueryRowsNoCache(ctx, &usersByStruct, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.FindByStructFilter(ctx, val, structFilter)
	})
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 11, user.Age)
		assert.Equal(t, "testUpdateByAge11@example.com", user.Email)
	}

}

func TestDelete(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	ctx := context.Background()

	// 按主键删除
	err := cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.DeleteByID[User](ctx, 1)
	})
	assert.NoError(t, err)

	// 按多个主键删除（id=2,3 存在，应该成功）
	err = cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.DeleteByIDs[User](ctx, 2, 3)
	})
	assert.NoError(t, err)

	// 按结构体条件删除
	err = cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.DeleteByStructFilter(ctx, &User{Age: 11})
	})
	assert.NoError(t, err)
}

func TestTransaction(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	ctx := context.Background()

	err := cdb.Transaction(ctx, func(ctx context.Context, tx *gormx.DB) error {
		var user User
		tx.GetByID(ctx, &user, 1)
		user.Age = 100
		tx.Update(ctx, &user)
		return nil
	})
	assert.NoError(t, err)
	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.GetByID(ctx, val, 1)
	})
	assert.NoError(t, err)
	assert.Equal(t, 100, user.Age)
}

func TestGetXDB(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})

	// 测试 GetXDB 返回的是创建时传入的同一个 *gormx.DB 实例
	xdb := cdb.GetXDB()
	assert.Same(t, tdb, xdb, "GetXDB() should return the same gormx.DB instance used to create CacheDB")
}
