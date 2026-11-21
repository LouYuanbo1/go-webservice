package antc

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/driver/redis"
	"github.com/LouYuanbo1/go-webservice/cachedb"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/alicebob/miniredis/v2"
	"github.com/skudra-go/ant"
	"github.com/skudra-go/ant/dialect/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type User struct {
	ant.Base
	Name  string
	Email string `ant:"unique"`
	Age   int
	Phone string
}

type AppDB struct {
	ant.Context
	Users ant.Table[User]
}

func (db *AppDB) AntContext() *ant.Context {
	return &db.Context
}

var Users = ant.Fields[User]()

const userDDL = `
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	age INTEGER NOT NULL,
	phone VARCHAR(255) NOT NULL
);`

func setupTestDB(t *testing.T) (*AppDB, *cache.Client, func()) {
	t.Helper()

	db, err := ant.Open[AppDB](context.Background(), sqlite.Dialect, ":memory:", ant.WithTracking(false))
	require.NoError(t, err)

	// 限制单连接，保证 :memory: 库对后续所有语句可见
	db.DB().SetMaxOpenConns(1)
	_, err = db.DB().ExecContext(context.Background(), userDDL)
	require.NoError(t, err)

	s := miniredis.RunT(t)
	port, err := strconv.Atoi(s.Port())
	require.NoError(t, err)

	redisConfig := &redis.Config{
		Host:     s.Host(),
		Port:     port,
		Password: "",
		DB:       0,
	}
	redisClient, err := redis.Init(redisConfig)
	require.NoError(t, err)

	cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	require.NoError(t, err)
	client := cache.NewClient(cacher)

	cleanup := func() { db.Close() }
	return db, client, cleanup
}

func setupTestDBWithRedis(t *testing.T) (*AppDB, *cache.Client, *miniredis.Miniredis, func()) {
	t.Helper()

	db, err := ant.Open[AppDB](context.Background(), sqlite.Dialect, ":memory:", ant.WithTracking(false))
	require.NoError(t, err)

	db.DB().SetMaxOpenConns(1)
	_, err = db.DB().ExecContext(context.Background(), userDDL)
	require.NoError(t, err)

	s := miniredis.RunT(t)
	port, err := strconv.Atoi(s.Port())
	require.NoError(t, err)

	redisConfig := &redis.Config{
		Host:     s.Host(),
		Port:     port,
		Password: "",
		DB:       0,
	}
	redisClient, err := redis.Init(redisConfig)
	require.NoError(t, err)

	cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	require.NoError(t, err)
	client := cache.NewClient(cacher)

	cleanup := func() { db.Close() }
	return db, client, s, cleanup
}

func newCacheDB(t *testing.T, db *AppDB, client *cache.Client) *CacheDB[AppDB, *AppDB] {
	t.Helper()
	return NewCacheDB(db, client, &cachedb.Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
}

func seedUserWithDB(t *testing.T, cdb *CacheDB[AppDB, *AppDB], name string, age int) *User {
	t.Helper()
	user := &User{Name: name, Email: name + "@test.com", Age: age, Phone: "10000000000"}
	err := cdb.ExecNoCache(context.Background(), func(ctx context.Context, db *AppDB) error {
		return db.Users.Insert(ctx, user)
	})
	require.NoError(t, err)
	return user
}

func TestNewCacheDB(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &cachedb.Config{
		TTL:                                10 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
	}
	cdb := NewCacheDB(db, client, cfg)

	assert.NotNil(t, cdb)
	assert.Same(t, db, cdb.db)
	assert.Same(t, client, cdb.cache)
	assert.Equal(t, cfg, cdb.cfg)
}

func TestGetDB(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	assert.Same(t, db, cdb.GetDB())
}

func TestCacheDB_GetSetDelCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	user := seedUserWithDB(t, cdb, "cdbCache", 25)

	err := cdb.SetCache(ctx, "cdb:user", user)
	assert.NoError(t, err)

	var cached User
	err = cdb.GetCache(ctx, "cdb:user", &cached)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, cached.ID)
	assert.Equal(t, "cdbCache", cached.Name)

	err = cdb.DelCache(ctx, "cdb:user")
	assert.NoError(t, err)

	err = cdb.GetCache(ctx, "cdb:user", &cached)
	assert.Error(t, err)
}

func TestCacheDB_Exec(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUserWithDB(t, cdb, "cdbExec", 25)

	// 先让缓存命中（按 email 查询，exec 更新名字和年龄但不改 email）
	queryFn := func(ctx context.Context, db *AppDB) (*User, error) {
		return db.Users.Query().Where(ant.Eq(&Users.Email, "cdbExec@test.com")).First(ctx)
	}
	u, err := cdb.Query(ctx, "cdb:user:exec", queryFn)
	require.NoError(t, err)
	assert.Equal(t, "cdbExec", u.Name)

	// Exec 跨表（此处单表）写入并失效缓存
	err = cdb.Exec(ctx,
		func(ctx context.Context, db *AppDB) error {
			_, err := db.Users.UpdateMany(ctx,
				[]*User{{ID: u.ID, Name: "cdbExecUpdated", Age: 30}},
				ant.F(&Users.Name),
				ant.F(&Users.Age),
			)
			return err
		},
		"cdb:user:exec",
	)
	require.NoError(t, err)

	// 缓存已被清理，再次查询应回源数据库并拿到最新数据
	u2, err := cdb.Query(ctx, "cdb:user:exec", queryFn)
	require.NoError(t, err)
	assert.Equal(t, "cdbExecUpdated", u2.Name)
	assert.Equal(t, 30, u2.Age)
}

func TestCacheDB_Query(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUserWithDB(t, cdb, "cdbQuery", 25)

	queryCount := 0
	queryFn := func(ctx context.Context, db *AppDB) (*User, error) {
		queryCount++
		return db.Users.Query().Where(ant.Eq(&Users.Name, "cdbQuery")).First(ctx)
	}

	u1, err := cdb.Query(ctx, "cdb:user:query", queryFn)
	require.NoError(t, err)
	assert.Equal(t, "cdbQuery", u1.Name)
	assert.Equal(t, 1, queryCount)

	u2, err := cdb.Query(ctx, "cdb:user:query", queryFn)
	require.NoError(t, err)
	assert.Equal(t, "cdbQuery", u2.Name)
	assert.Equal(t, 1, queryCount)
}

func TestCacheDB_Query_WithTTL(t *testing.T) {
	db, client, s, cleanup := setupTestDBWithRedis(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUserWithDB(t, cdb, "cdbTTL", 25)

	queryCount := 0
	queryFn := func(ctx context.Context, db *AppDB) (*User, error) {
		queryCount++
		return db.Users.Query().Where(ant.Eq(&Users.Name, "cdbTTL")).First(ctx)
	}

	u1, err := cdb.Query(ctx, "cdb:user:ttl", queryFn, cachedb.WithTTL(1*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "cdbTTL", u1.Name)
	assert.Equal(t, 1, queryCount)

	s.FastForward(2 * time.Second)

	u2, err := cdb.Query(ctx, "cdb:user:ttl", queryFn)
	require.NoError(t, err)
	assert.Equal(t, "cdbTTL", u2.Name)
	assert.Equal(t, 2, queryCount)
}

func TestCacheDB_Query_NotFound(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	queryFn := func(ctx context.Context, db *AppDB) (*User, error) {
		return db.Users.Query().Where(ant.Eq(&Users.Name, "cdbNotFound")).First(ctx)
	}

	_, err := cdb.Query(ctx, "cdb:user:notfound", queryFn)
	require.Error(t, err)
}

func TestCacheDB_QueryIndex(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	u := seedUserWithDB(t, cdb, "cdbQueryIndex", 25)

	indexCallCount := 0
	primaryCallCount := 0
	indexQuery := func(ctx context.Context, db *AppDB) (int64, *User, error) {
		indexCallCount++
		row, err := db.Users.Query().Where(ant.Eq(&Users.Name, "cdbQueryIndex")).First(ctx)
		if err != nil {
			return 0, nil, err
		}
		return row.ID, row, nil
	}
	primaryQuery := func(ctx context.Context, db *AppDB, id int64) (*User, error) {
		primaryCallCount++
		return db.Users.Find(ctx, id)
	}
	keyer := func(id int64) string { return fmt.Sprintf("cdb:user:index:primary:%d", id) }

	// 第一次：索引与主键缓存均未命中，走数据库
	u1, err := cdb.QueryIndex(ctx, "cdb:user:index", keyer, indexQuery, primaryQuery)
	require.NoError(t, err)
	assert.Equal(t, u.ID, u1.ID)
	assert.Equal(t, 1, indexCallCount)
	assert.Equal(t, 0, primaryCallCount)

	// 第二次：命中索引缓存，直接读主键缓存，不再走数据库
	u2, err := cdb.QueryIndex(ctx, "cdb:user:index", keyer, indexQuery, primaryQuery)
	require.NoError(t, err)
	assert.Equal(t, u.ID, u2.ID)
	assert.Equal(t, 1, indexCallCount)
	assert.Equal(t, 0, primaryCallCount)
}

func TestCacheDB_QueryIndex_MissingPrimary(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	u := seedUserWithDB(t, cdb, "cdbQueryIndex2", 25)

	indexQuery := func(ctx context.Context, db *AppDB) (int64, *User, error) {
		row, err := db.Users.Query().Where(ant.Eq(&Users.Name, "cdbQueryIndex2")).First(ctx)
		if err != nil {
			return 0, nil, err
		}
		return row.ID, row, nil
	}
	primaryQuery := func(ctx context.Context, db *AppDB, id int64) (*User, error) {
		return db.Users.Find(ctx, id)
	}
	keyer := func(id int64) string { return fmt.Sprintf("cdb:user:index:primary:%d", id) }

	u1, err := cdb.QueryIndex(ctx, "cdb:user:index2", keyer, indexQuery, primaryQuery)
	require.NoError(t, err)
	assert.Equal(t, u.ID, u1.ID)

	// 主键缓存存在时直接命中
	u2, err := cdb.QueryIndex(ctx, "cdb:user:index2", keyer, indexQuery, primaryQuery)
	require.NoError(t, err)
	assert.Equal(t, u.ID, u2.ID)
}

func TestCacheDB_QueryNoCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUserWithDB(t, cdb, "cdbNoCache", 25)

	u, err := cdb.QueryNoCache(ctx, func(ctx context.Context, db *AppDB) (*User, error) {
		return db.Users.Query().Where(ant.Eq(&Users.Name, "cdbNoCache")).First(ctx)
	})
	require.NoError(t, err)
	assert.Equal(t, "cdbNoCache", u.Name)
}

func TestCacheDB_QueryRowsNoCache(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()
	seedUserWithDB(t, cdb, "cdbRows1", 25)
	seedUserWithDB(t, cdb, "cdbRows2", 26)

	users, err := cdb.QueryRowsNoCache(ctx, func(ctx context.Context, db *AppDB) ([]*User, error) {
		return db.Users.Query().Where(ant.Like(&Users.Name, "cdbRows%")).All(ctx)
	})
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestCacheDB_Transaction(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.Transaction(ctx, func(tx *AppDB) error {
		u := &User{Name: "cdbTx", Email: "cdbTx@test.com", Age: 25, Phone: "10000000000"}
		if err := tx.Users.Insert(ctx, u); err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	row, err := db.Users.Query().Where(ant.Eq(&Users.Name, "cdbTx")).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "cdbTx", row.Name)
}

func TestCacheDB_Transaction_Rollback(t *testing.T) {
	db, client, cleanup := setupTestDB(t)
	defer cleanup()

	cdb := newCacheDB(t, db, client)
	ctx := context.Background()

	err := cdb.Transaction(ctx, func(tx *AppDB) error {
		u := &User{Name: "cdbTxRollback", Email: "cdbTxRollback@test.com", Age: 25, Phone: "10000000000"}
		if err := tx.Users.Insert(ctx, u); err != nil {
			return err
		}
		return fmt.Errorf("force rollback")
	})
	require.Error(t, err)

	_, err = db.Users.Query().Where(ant.Eq(&Users.Name, "cdbTxRollback")).First(ctx)
	require.Error(t, err)
}
