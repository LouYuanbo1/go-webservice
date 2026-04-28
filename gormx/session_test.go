package gormx

import (
	"context"
	"strconv"
	"testing"
	"time"

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

	// GetByMapFilter
	userByMap := &User{}
	err = xdb.GetByMapFilter(ctx, userByMap, map[string]any{"name": "testCreate3"})
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), userByMap.ID)
	assert.Equal(t, "testCreate3", userByMap.Name)
	assert.Equal(t, 1, userByMap.Gender)
	assert.Equal(t, 13, userByMap.Age)
	assert.Equal(t, "testCreate3@example.com", userByMap.Email)
	assert.Equal(t, "10000000003", userByMap.Phone)
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
	usersByStruct := make([]*User, 0)
	err = xdb.FindByStructFilter(ctx, &usersByStruct, &User{Age: 10})
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 10, user.Age)
	}

	// FindByMapFilter
	usersByMap := make([]*User, 0)
	err = xdb.FindByMapFilter(ctx, &usersByMap, map[string]any{"age": 11})
	assert.NoError(t, err)
	for _, user := range usersByMap {
		assert.Equal(t, 11, user.Age)
	}

	// FindByPage
	usersByPage := make([]*User, 0)
	err = xdb.FindByPage(ctx, &usersByPage, "id", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, usersByPage, 10)
	for i, user := range usersByPage {
		id := uint64(i + 1)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}

	// FindByCursor
	usersByCursor := make([]*User, 0)
	err = xdb.FindByCursor(ctx, &usersByCursor, "id", 10, 10)
	assert.NoError(t, err)
	assert.Len(t, usersByCursor, 10)
	for i, user := range usersByCursor {
		id := uint64(i + 11)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "testCreate"+strconv.Itoa(int(id)), user.Name)
	}
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

	usersByStruct := make([]*User, 0)
	err = xdb.FindByStructFilter(ctx, &usersByStruct, structFilter)
	assert.NoError(t, err)
	for _, user := range usersByStruct {
		assert.Equal(t, 11, user.Age)
		assert.Equal(t, "testUpdateByAge11@example.com", user.Email)
	}

	// 按 map 条件更新
	mapFilter := map[string]any{"age": 12}
	mapUpdate := map[string]any{"email": "testUpdateByAge12@example.com"}
	err = xdb.UpdatesByMapFilter(ctx, &User{}, mapFilter, mapUpdate)
	assert.NoError(t, err)

	usersByMap := make([]*User, 0)
	err = xdb.FindByMapFilter(ctx, &usersByMap, mapFilter)
	assert.NoError(t, err)
	for _, user := range usersByMap {
		assert.Equal(t, 12, user.Age)
		assert.Equal(t, "testUpdateByAge12@example.com", user.Email)
	}
}

func TestDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	prepareSampleData(t, db)

	xdb := NewDB(db)
	ctx := context.Background()

	// 按主键删除
	err := xdb.DeleteByID(ctx, &User{}, 1)
	assert.NoError(t, err)

	// 按多个主键删除（id=2,3 存在，应该成功）
	err = xdb.DeleteByIDs(ctx, &User{}, []uint64{2, 3})
	assert.NoError(t, err)

	// 按结构体条件删除
	err = xdb.DeleteByStructFilter(ctx, &User{}, &User{Age: 11})
	assert.NoError(t, err)

	// 按 map 条件删除
	err = xdb.DeleteByMapFilter(ctx, &User{}, map[string]any{"age": 12})
	assert.NoError(t, err)
}
