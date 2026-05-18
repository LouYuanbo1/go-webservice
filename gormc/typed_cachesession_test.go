package gormc

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/stretchr/testify/assert"
)

// ---------- 测试用例 ----------

func TestTypedCreate(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	tdb := gormx.NewTypedDB(db)
	cdb := NewTypedCacheDB(tdb, cacheClient, &Config{
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
	err := cdb.ExecNoCache(ctx, func(ctx context.Context, tdb *gormx.TypedDB) error {
		session := gormx.GetSession[User, uint64](tdb)
		return session.Create(ctx, user)
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
	err = cdb.ExecNoCache(ctx, func(ctx context.Context, tdb *gormx.TypedDB) error {
		session := gormx.GetSession[User, uint64](tdb)
		return session.CreateInBatches(ctx, users, 100)
	})
	assert.NoError(t, err)
}

func TestTypedGet(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewTypedDB(db)
	cdb := NewTypedCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	session := GetSession[User, uint64](cdb)
	ctx := context.Background()

	// GetByID
	userByID := &User{}
	err := session.QueryNoCache(ctx, userByID, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *User) error {
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
	err = session.Query(ctx, "userByStruct", userByStruct, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *User) error {
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

func TestTypedFind(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewTypedDB(db)
	cdb := NewTypedCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	session := GetSession[User, uint64](cdb)
	ctx := context.Background()

	// FindByIDs
	usersByID := make([]*User, 0)
	err := session.QueryRowsNoCache(ctx, &usersByID, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *[]*User) error {
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
	usersByStruct := make([]*User, 0)
	err = session.QueryRowsNoCache(ctx, &usersByStruct, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *[]*User) error {
		return db.FindByStructFilter(ctx, val, &User{Age: 10})
	})
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 10, user.Age)
	}

	// FindByPage
	usersByPage := make([]*User, 0)
	err = session.QueryRowsNoCache(ctx, &usersByPage, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *[]*User) error {
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
	usersByCursor := make([]*User, 0)
	err = session.QueryRowsNoCache(ctx, &usersByCursor, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *[]*User) error {
		db.FindByCursor(ctx, val, 10, 10)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, usersByCursor, 10)
	for i, user := range usersByCursor {
		id := uint64(i + 11)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}
}

func TestTypedUpdate(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewTypedDB(db)
	cdb := NewTypedCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	session := GetSession[User, uint64](cdb)
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
	err := session.ExecNoCache(ctx, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User]) error {
		return db.Update(ctx, userUpdate)
	})
	assert.NoError(t, err)

	userByID := &User{}
	//不单独缓存,避免影响其他测试结果
	err = session.QueryNoCache(ctx, userByID, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *User) error {
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
	err = session.ExecNoCache(ctx, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User]) error {
		return db.UpdateByStructFilter(ctx, structFilter, structUpdate)
	})
	assert.NoError(t, err)
	usersByStruct := make([]*User, 0)
	err = session.QueryRowsNoCache(ctx, &usersByStruct, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User], val *[]*User) error {
		return db.FindByStructFilter(ctx, val, structFilter)
	})
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 11, user.Age)
		assert.Equal(t, "testUpdateByAge11@example.com", user.Email)
	}

}

func TestTypedDelete(t *testing.T) {
	db, cacheClient, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db, cacheClient)

	tdb := gormx.NewTypedDB(db)
	cdb := NewTypedCacheDB(tdb, cacheClient, &Config{
		TTL:                                3 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 2 * time.Second,
	})
	session := GetSession[User, uint64](cdb)
	ctx := context.Background()

	// 按主键删除
	err := session.ExecNoCache(ctx, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User]) error {
		return db.DeleteByID(ctx, 1)
	})
	assert.NoError(t, err)

	// 按多个主键删除（id=2,3 存在，应该成功）
	err = session.ExecNoCache(ctx, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User]) error {
		return db.DeleteByIDs(ctx, 2, 3)
	})
	assert.NoError(t, err)

	// 按结构体条件删除
	err = session.ExecNoCache(ctx, func(ctx context.Context, db gormx.TypedSession[User, uint64, *User]) error {
		return db.DeleteByStructFilter(ctx, &User{Age: 11})
	})
	assert.NoError(t, err)
}
