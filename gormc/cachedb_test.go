package gormc

import (
	"context"
	"fmt"
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
	assert.NoError(t, err)

	s := miniredis.RunT(t)
	port, err := strconv.Atoi(s.Port())
	assert.NoError(t, err)

	redisConfig := &redis.Config{
		Host:     s.Host(),
		Port:     port,
		Password: "",
		DB:       0,
	}
	redisClient, err := redis.InitRedisClient(redisConfig)
	assert.NoError(t, err)

	cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	assert.NoError(t, err)
	client := cache.NewClient(cacher)

	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return db, client, cleanup
}

func newCacheDB(t *testing.T, db *gorm.DB, client *cache.Client) *CacheDB {
	t.Helper()
	return NewCacheDB(gormx.NewDB(db), client, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
}

func setupTestDBWithRedis(t *testing.T) (*gorm.DB, *cache.Client, *miniredis.Miniredis, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	s := miniredis.RunT(t)
	port, err := strconv.Atoi(s.Port())
	assert.NoError(t, err)

	redisConfig := &redis.Config{
		Host:     s.Host(),
		Port:     port,
		Password: "",
		DB:       0,
	}
	redisClient, err := redis.InitRedisClient(redisConfig)
	assert.NoError(t, err)

	cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	assert.NoError(t, err)
	client := cache.NewClient(cacher)

	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return db, client, s, cleanup
}

func seedUser(t *testing.T, cdb *CacheDB, name string, age int) *User {
	t.Helper()
	user := &User{Name: name, Age: age, Email: name + "@test.com", Phone: "10000000000"}
	err := cdb.ExecNoCache(context.Background(), func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, user)
	})
	assert.NoError(t, err)
	return user
}

func seedUsers(t *testing.T, cdb *CacheDB, count int) []*User {
	t.Helper()
	users := make([]*User, 0, count)
	for i := 1; i <= count; i++ {
		users = append(users, &User{
			Name:   "user" + string(rune('A'+i-1)),
			Gender: i % 2,
			Age:    20 + i,
			Email:  "user" + string(rune('A'+i-1)) + "@test.com",
			Phone:  "1000000000" + string(rune('0'+i%10)),
		})
	}
	err := cdb.ExecNoCache(context.Background(), func(ctx context.Context, db *gormx.DB) error {
		return db.CreateInBatches(ctx, &users, 100)
	})
	assert.NoError(t, err)
	return users
}

func TestGetCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	user := seedUser(t, cdb, "getCache", 25)
	err := cdb.SetCache(ctx, "user:getCache", user)
	assert.NoError(t, err)

	var cached User
	err = cdb.GetCache(ctx, "user:getCache", &cached)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, cached.ID)
	assert.Equal(t, "getCache", cached.Name)
}

func TestGetCache_NotFound(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)

	var val User
	err := cdb.GetCache(context.Background(), "nonexistent_key", &val)
	assert.Error(t, err)
}

func TestSetCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	user := &User{ID: 1, Name: "setCache", Age: 30}
	err := cdb.SetCache(ctx, "user:setCache", user)
	assert.NoError(t, err)

	var cached User
	err = cdb.GetCache(ctx, "user:setCache", &cached)
	assert.NoError(t, err)
	assert.Equal(t, "setCache", cached.Name)
}

func TestSetCache_WithTTL(t *testing.T) {
	db, client, s, cleanup := setupTestDBWithRedis(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	user := &User{ID: 1, Name: "setCacheTTL", Age: 30}
	err := cdb.SetCache(ctx, "user:setCacheTTL", user, WithTTL(1*time.Second))
	assert.NoError(t, err)

	var cached User
	err = cdb.GetCache(ctx, "user:setCacheTTL", &cached)
	assert.NoError(t, err)
	assert.Equal(t, "setCacheTTL", cached.Name)

	s.FastForward(2 * time.Second)

	err = cdb.GetCache(ctx, "user:setCacheTTL", &cached)
	assert.Error(t, err)
}

func TestDelCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.SetCache(ctx, "key1", "value1")
	assert.NoError(t, err)
	err = cdb.SetCache(ctx, "key2", "value2")
	assert.NoError(t, err)

	err = cdb.DelCache(ctx, "key1", "key2")
	assert.NoError(t, err)

	var val string
	err = cdb.GetCache(ctx, "key1", &val)
	assert.Error(t, err)
	err = cdb.GetCache(ctx, "key2", &val)
	assert.Error(t, err)
}

func TestDelCache_MultipleKeys(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		err := cdb.SetCache(ctx, fmt.Sprintf("key%d", i), i)
		assert.NoError(t, err)
	}

	err := cdb.DelCache(ctx, "key1", "key2", "key3", "key4", "key5")
	assert.NoError(t, err)

	for i := 1; i <= 5; i++ {
		var val int
		err = cdb.GetCache(ctx, fmt.Sprintf("key%d", i), &val)
		assert.Error(t, err)
	}
}

func TestQuery_CacheMissThenHit(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUser(t, cdb, "queryTest", 25)

	queryCount := 0
	queryFn := func(ctx context.Context, db *gormx.DB, val *User) error {
		queryCount++
		return db.Where("name = ?", "queryTest").First(ctx, val)
	}

	var user1 User
	err := cdb.Query(ctx, "user:queryTest", &user1, queryFn)
	assert.NoError(t, err)
	assert.Equal(t, "queryTest", user1.Name)
	assert.Equal(t, 1, queryCount)

	var user2 User
	err = cdb.Query(ctx, "user:queryTest", &user2, queryFn)
	assert.NoError(t, err)
	assert.Equal(t, "queryTest", user2.Name)
	assert.Equal(t, 1, queryCount)
}

func TestQuery_WithTTL(t *testing.T) {
	db, client, s, cleanup := setupTestDBWithRedis(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUser(t, cdb, "queryTTL", 25)

	queryCount := 0
	queryFn := func(ctx context.Context, db *gormx.DB, val *User) error {
		queryCount++
		return db.Where("name = ?", "queryTTL").First(ctx, val)
	}

	var user1 User
	err := cdb.Query(ctx, "user:queryTTL", &user1, queryFn, WithTTL(1*time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 1, queryCount)

	s.FastForward(2 * time.Second)

	var user2 User
	err = cdb.Query(ctx, "user:queryTTL", &user2, queryFn)
	assert.NoError(t, err)
	assert.Equal(t, 2, queryCount)
}

func TestQuery_NotFound(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	var user User
	err := cdb.Query(ctx, "user:notFound", &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.First(ctx, val, 99999)
	})
	assert.Error(t, err)
}

func TestQueryIndex(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	user := seedUser(t, cdb, "indexUser", 30)

	indexQueryCount := 0
	primaryQueryCount := 0

	var found User
	err := cdb.QueryIndex(ctx,
		"idx:email:"+user.Email,
		&found,
		func(primary uint64) string {
			return fmt.Sprintf("user:%d", primary)
		},
		func(ctx context.Context, db *gormx.DB, val *User) (uint64, error) {
			indexQueryCount++
			err := db.Where("email = ?", user.Email).First(ctx, val)
			return val.ID, err
		},
		func(ctx context.Context, db *gormx.DB, val *User, primaryKey uint64) error {
			primaryQueryCount++
			return db.First(ctx, val, primaryKey)
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "indexUser", found.Name)
	assert.Equal(t, 1, indexQueryCount)
	assert.Equal(t, 0, primaryQueryCount)

	var found2 User
	err = cdb.QueryIndex(ctx,
		"idx:email:"+user.Email,
		&found2,
		func(primary uint64) string {
			return fmt.Sprintf("user:%d", primary)
		},
		func(ctx context.Context, db *gormx.DB, val *User) (uint64, error) {
			indexQueryCount++
			err := db.Where("email = ?", user.Email).First(ctx, val)
			return val.ID, err
		},
		func(ctx context.Context, db *gormx.DB, val *User, primaryKey uint64) error {
			primaryQueryCount++
			return db.First(ctx, val, primaryKey)
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found2.ID)
	assert.Equal(t, 1, indexQueryCount)
	assert.Equal(t, 0, primaryQueryCount)
}

func TestQueryIndex_PrimaryCacheHit(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	user := seedUser(t, cdb, "primaryCacheUser", 30)

	err := cdb.SetCache(ctx, fmt.Sprintf("user:%d", user.ID), user)
	assert.NoError(t, err)

	indexQueryCount := 0
	primaryQueryCount := 0

	var found User
	err = cdb.QueryIndex(ctx,
		"idx:email:"+user.Email,
		&found,
		func(primary uint64) string {
			return fmt.Sprintf("user:%d", primary)
		},
		func(ctx context.Context, db *gormx.DB, val *User) (uint64, error) {
			indexQueryCount++
			err := db.Where("email = ?", user.Email).First(ctx, val)
			return val.ID, err
		},
		func(ctx context.Context, db *gormx.DB, val *User, primaryKey uint64) error {
			primaryQueryCount++
			return db.First(ctx, val, primaryKey)
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, 1, indexQueryCount)
	assert.Equal(t, 0, primaryQueryCount)
}

func TestQueryIndex_NotFound(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	var found User
	err := cdb.QueryIndex(ctx,
		"idx:email:nonexistent@test.com",
		&found,
		func(primary uint64) string {
			return fmt.Sprintf("user:%d", primary)
		},
		func(ctx context.Context, db *gormx.DB, val *User) (uint64, error) {
			err := db.Where("email = ?", "nonexistent@test.com").First(ctx, val)
			return 0, err
		},
		func(ctx context.Context, db *gormx.DB, val *User, primaryKey uint64) error {
			return db.First(ctx, val, primaryKey)
		},
	)
	assert.Error(t, err)
}

func TestQueryNoCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	user := seedUser(t, cdb, "noCacheUser", 25)

	var found User
	err := cdb.QueryNoCache(ctx, &found, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.First(ctx, val, user.ID)
	})
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "noCacheUser", found.Name)

	var cached User
	err = cdb.GetCache(ctx, fmt.Sprintf("user:%d", user.ID), &cached)
	assert.Error(t, err)
}

func TestQueryNoCache_NotFound(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)

	var found User
	err := cdb.QueryNoCache(context.Background(), &found, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.First(ctx, val, 99999)
	})
	assert.Error(t, err)
}

func TestQueryRowsNoCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUsers(t, cdb, 5)

	var users []User
	err := cdb.QueryRowsNoCache(ctx, &users, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.Find(ctx, val)
	})
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestQueryRowsNoCache_WithWhere(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUsers(t, cdb, 5)

	var users []User
	err := cdb.QueryRowsNoCache(ctx, &users, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.Where("age > ?", 22).Find(ctx, val)
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestExec(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.SetCache(ctx, "key_to_del", "value")
	assert.NoError(t, err)

	err = cdb.Exec(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, &User{Name: "execUser", Age: 30, Email: "exec@test.com", Phone: "1111111111"})
	}, "key_to_del")
	assert.NoError(t, err)

	var val string
	err = cdb.GetCache(ctx, "key_to_del", &val)
	assert.Error(t, err)
}

func TestExec_MultipleKeys(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		err := cdb.SetCache(ctx, fmt.Sprintf("exec_key%d", i), i)
		assert.NoError(t, err)
	}

	err := cdb.Exec(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, &User{Name: "execMulti", Age: 30, Email: "execMulti@test.com", Phone: "2222222222"})
	}, "exec_key1", "exec_key2", "exec_key3")
	assert.NoError(t, err)

	for i := 1; i <= 3; i++ {
		var val int
		err = cdb.GetCache(ctx, fmt.Sprintf("exec_key%d", i), &val)
		assert.Error(t, err)
	}
}

func TestExec_NoKeys(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.Exec(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, &User{Name: "execNoKeys", Age: 30, Email: "execNoKeys@test.com", Phone: "3333333333"})
	})
	assert.NoError(t, err)

	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.Where("name = ?", "execNoKeys").First(ctx, val)
	})
	assert.NoError(t, err)
	assert.Equal(t, "execNoKeys", user.Name)
}

func TestExecNoCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.ExecNoCache(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, &User{Name: "execNoCache", Age: 30, Email: "execNoCache@test.com", Phone: "4444444444"})
	})
	assert.NoError(t, err)

	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.Where("name = ?", "execNoCache").First(ctx, val)
	})
	assert.NoError(t, err)
	assert.Equal(t, "execNoCache", user.Name)
}

func TestTransaction_Commit(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.Transaction(ctx, func(tx *gormx.Executor) error {
		return tx.Create(ctx, &User{Name: "txCommit", Age: 30, Email: "txCommit@test.com", Phone: "5555555555"})
	})
	assert.NoError(t, err)

	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.Where("name = ?", "txCommit").First(ctx, val)
	})
	assert.NoError(t, err)
	assert.Equal(t, "txCommit", user.Name)
}

func TestTransaction_Rollback(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	expectedErr := fmt.Errorf("rollback on purpose")
	err := cdb.Transaction(ctx, func(tx *gormx.Executor) error {
		createErr := tx.Create(ctx, &User{Name: "txRollback", Age: 30, Email: "txRollback@test.com", Phone: "6666666666"})
		if createErr != nil {
			return createErr
		}
		return expectedErr
	})
	assert.Error(t, err)

	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.Where("name = ?", "txRollback").First(ctx, val)
	})
	assert.Error(t, err)
}

func TestTransaction_UpdateInside(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUser(t, cdb, "txUpdate", 20)

	err := cdb.Transaction(ctx, func(tx *gormx.Executor) error {
		var u User
		if err := tx.First(ctx, &u, "name = ?", "txUpdate"); err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", u.ID).Update(ctx, "age", 100)
	})
	assert.NoError(t, err)

	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.Where("name = ?", "txUpdate").First(ctx, val)
	})
	assert.NoError(t, err)
	assert.Equal(t, 100, user.Age)
}

func TestGetXDB(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	tdb := gormx.NewDB(db)
	cdb := NewCacheDB(tdb, client, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})

	xdb := cdb.GetXDB()
	assert.Same(t, tdb, xdb)
}

func TestNewCacheDB(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	tdb := gormx.NewDB(db)
	cfg := &Config{
		TTL:                                10 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
	}
	cdb := NewCacheDB(tdb, client, cfg)

	assert.NotNil(t, cdb)
	assert.Same(t, tdb, cdb.db)
	assert.Same(t, client, cdb.cache)
	assert.Equal(t, cfg, cdb.cfg)
}

func TestCacheIntegration_ExecThenQuery(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.SetCache(ctx, "user:1", "stale_data")
	assert.NoError(t, err)

	err = cdb.Exec(ctx, func(ctx context.Context, db *gormx.DB) error {
		return db.Create(ctx, &User{ID: 1, Name: "freshData", Age: 30, Email: "fresh@test.com", Phone: "7777777777"})
	}, "user:1")
	assert.NoError(t, err)

	var cached string
	err = cdb.GetCache(ctx, "user:1", &cached)
	assert.Error(t, err)

	var user User
	err = cdb.QueryNoCache(ctx, &user, func(ctx context.Context, db *gormx.DB, val *User) error {
		return db.First(ctx, val, 1)
	})
	assert.NoError(t, err)
	assert.Equal(t, "freshData", user.Name)
}

func TestCache_ComplexTypes(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	type ComplexData struct {
		Numbers []int
		Mapping map[string]string
	}
	data := ComplexData{
		Numbers: []int{1, 2, 3},
		Mapping: map[string]string{"a": "apple", "b": "banana"},
	}

	err := cdb.SetCache(ctx, "complex", data)
	assert.NoError(t, err)

	var cached ComplexData
	err = cdb.GetCache(ctx, "complex", &cached)
	assert.NoError(t, err)
	assert.Equal(t, data.Numbers, cached.Numbers)
	assert.Equal(t, data.Mapping, cached.Mapping)
}

func TestGetCache_Struct(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	user := &User{ID: 100, Name: "structCache", Age: 30, Email: "struct@test.com", Phone: "8888888888"}
	err := cdb.SetCache(ctx, "user:struct", user)
	assert.NoError(t, err)

	var cached User
	err = cdb.GetCache(ctx, "user:struct", &cached)
	assert.NoError(t, err)
	assert.Equal(t, uint64(100), cached.ID)
	assert.Equal(t, "structCache", cached.Name)
	assert.Equal(t, 30, cached.Age)
	assert.Equal(t, "struct@test.com", cached.Email)
	assert.Equal(t, "8888888888", cached.Phone)
}

func TestQueryRowsNoCache_Empty(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	var users []User
	err := cdb.QueryRowsNoCache(ctx, &users, func(ctx context.Context, db *gormx.DB, val *[]User) error {
		return db.Where("name = ?", "nonexistent").Find(ctx, val)
	})
	assert.NoError(t, err)
	assert.Empty(t, users)
}

func TestQuery_ConsecutiveAccess(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUser(t, cdb, "consecutive", 25)

	queryCount := 0
	queryFn := func(ctx context.Context, db *gormx.DB, val *User) error {
		queryCount++
		return db.Where("name = ?", "consecutive").First(ctx, val)
	}

	for i := 0; i < 10; i++ {
		var user User
		err := cdb.Query(ctx, "user:consecutive", &user, queryFn)
		assert.NoError(t, err)
		assert.Equal(t, "consecutive", user.Name)
	}
	assert.Equal(t, 1, queryCount)
}
