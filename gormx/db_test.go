package gormx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	ID        uint64         `gorm:"primaryKey" redis:"id"`
	Name      string         `gorm:"not null" redis:"name"`
	Gender    int            `gorm:"default:0"`
	Age       int            `gorm:"default:0"`
	Email     string         `gorm:"not null" redis:"email"`
	Phone     string         `gorm:"not null" redis:"phone"`
	CreatedAt time.Time      `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time      `gorm:"not null;default:current_timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) GetID() uint64      { return u.ID }
func (u *User) PrimaryKey() string { return "id" }
func (u *User) TableName() string  { return "users" }

type Order struct {
	ID     uint64  `gorm:"primaryKey"`
	UserID uint64  `gorm:"not null"`
	Amount float64 `gorm:"not null"`
}

func (o *Order) TableName() string { return "orders" }

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&User{}, &Order{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return db, cleanup
}

func seedUsers(t *testing.T, xdb *DB, count int) []*User {
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
	err := xdb.CreateInBatches(context.Background(), &users, 100)
	assert.NoError(t, err)
	return users
}

func TestCreate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "createTest", Age: 25, Email: "create@test.com", Phone: "1234567890"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "createTest", found.Name)
}

func TestCreate_NilModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Create(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestCreateInBatches(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	users := seedUsers(t, xdb, 5)

	var found []User
	err := xdb.Find(context.Background(), &found)
	assert.NoError(t, err)
	assert.Len(t, found, 5)
	assert.Equal(t, users[0].ID, found[0].ID)
}

func TestCreateInBatches_NilModels(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.CreateInBatches(context.Background(), (*[]User)(nil), 100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestFirst(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "firstTest", Age: 30, Email: "first@test.com", Phone: "1111111111"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "firstTest", found.Name)
}

func TestFirst_WithConds(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var found User
	err := xdb.First(context.Background(), &found, "name = ?", "userB")
	assert.NoError(t, err)
	assert.Equal(t, "userB", found.Name)
}

func TestFirst_NilDest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.First(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestFirst_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	var found User
	err := xdb.First(context.Background(), &found, 99999)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFirstFailed))
}

func TestScan(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.Model(&User{}).Scan(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestFind(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestFind_WithWhere(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.Where("age > ?", 22).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestFind_NilDest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Find(context.Background(), (*[]User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var count int64
	err := xdb.Model(&User{}).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestCount_WithWhere(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var count int64
	err := xdb.Model(&User{}).Where("age > ?", 22).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Greater(t, count, int64(0))
}

func TestCount_NilCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Count(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "beforeUpdate", Age: 25, Email: "before@test.com", Phone: "2222222222"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	err = xdb.Model(&User{}).Where("id = ?", user.ID).Update(context.Background(), "name", "afterUpdate")
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "afterUpdate", found.Name)
}

func TestUpdates(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "beforeUpdates", Age: 25, Email: "before@test.com", Phone: "3333333333"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	err = xdb.Model(&User{}).Where("id = ?", user.ID).Updates(context.Background(), &User{Name: "afterUpdates", Age: 30})
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "afterUpdates", found.Name)
	assert.Equal(t, 30, found.Age)
}

func TestUpdates_NilModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Updates(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "toDelete", Age: 30, Email: "delete@test.com", Phone: "4444444444"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	err = xdb.Delete(context.Background(), &User{}, user.ID)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.True(t, found.DeletedAt.Valid)
}

func TestDelete_NilModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Delete(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestTransaction_Commit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)

	err := xdb.Transaction(context.Background(), func(tx *Executor) error {
		u := &User{Name: "txUser", Age: 20, Email: "tx@test.com", Phone: "5555555555"}
		return tx.Create(context.Background(), u)
	})
	assert.NoError(t, err)

	var count int64
	err = xdb.Model(&User{}).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestTransaction_Rollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)

	expectedErr := errors.New("rollback on purpose")
	err := xdb.Transaction(context.Background(), func(tx *Executor) error {
		u := &User{Name: "txRollback", Age: 20, Email: "rollback@test.com", Phone: "6666666666"}
		createErr := tx.Create(context.Background(), u)
		if createErr != nil {
			return createErr
		}
		return expectedErr
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))

	var count int64
	err = xdb.Model(&User{}).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestTransaction_UpdateInside(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "txUpdate", Age: 20, Email: "txupdate@test.com", Phone: "7777777777"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	err = xdb.Transaction(context.Background(), func(tx *Executor) error {
		var u User
		if err := tx.First(context.Background(), &u, user.ID); err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", u.ID).Update(context.Background(), "age", 100)
	})
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 100, found.Age)
}

func TestWhere(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.Where("age > ?", 22).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestStructFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.StructFilter(&User{Age: 21}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Equal(t, 21, u.Age)
	}
}

func TestMapFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.MapFilter(map[string]any{"age": 22}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Equal(t, 22, u.Age)
	}
}

func TestSelect(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "selectTest", Age: 99, Email: "select@test.com", Phone: "8888888888"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = xdb.Select("name", "age").First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "selectTest", found.Name)
	assert.Equal(t, 99, found.Age)
}

func TestOrder(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var users []User
	err := xdb.Order("age desc").Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.GreaterOrEqual(t, users[0].Age, users[1].Age)
	assert.GreaterOrEqual(t, users[1].Age, users[2].Age)
}

func TestOrderByColumn(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var users []User
	err := xdb.OrderByColumn(clause.OrderByColumn{Column: clause.Column{Name: "age"}, Desc: true}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.GreaterOrEqual(t, users[0].Age, users[1].Age)
}

func TestOrderBy(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var users []User
	err := xdb.OrderBy(clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "age"}, Desc: true}}}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.GreaterOrEqual(t, users[0].Age, users[1].Age)
}

func TestLimit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	var users []User
	err := xdb.Limit(3).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestOffset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	users := seedUsers(t, xdb, 5)

	var result []User
	err := xdb.Order("id asc").Offset(2).Limit(2).Find(context.Background(), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, users[2].ID, result[0].ID)
}

func TestGroup(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	type AgeCount struct {
		Age   int
		Count int
	}
	var results []AgeCount
	err := xdb.Model(&User{}).Select("age, count(*) as count").Group("age").Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestHaving(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user1 := &User{Name: "havingA", Age: 25, Email: "a@test.com", Phone: "1111111111"}
	user2 := &User{Name: "havingB", Age: 25, Email: "b@test.com", Phone: "2222222222"}
	user3 := &User{Name: "havingC", Age: 30, Email: "c@test.com", Phone: "3333333333"}
	assert.NoError(t, xdb.Create(context.Background(), user1))
	assert.NoError(t, xdb.Create(context.Background(), user2))
	assert.NoError(t, xdb.Create(context.Background(), user3))

	type AgeCount struct {
		Age   int
		Count int
	}
	var results []AgeCount
	err := xdb.Model(&User{}).Select("age, count(*) as count").Group("age").Having("count > ?", 1).Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 25, results[0].Age)
}

func TestJoins(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "joinUser", Age: 30, Email: "join@test.com", Phone: "4444444444"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	order := &Order{UserID: user.ID, Amount: 100.5}
	err = db.Create(order).Error
	assert.NoError(t, err)

	type UserWithOrder struct {
		ID     uint64
		Name   string
		Amount float64
	}
	var results []UserWithOrder
	err = xdb.Model(&User{}).Select("users.id, users.name, orders.amount").
		Joins("JOIN orders ON orders.user_id = users.id").
		Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, user.Name, results[0].Name)
}

func TestInnerJoins(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "innerJoinUser", Age: 30, Email: "innerjoin@test.com", Phone: "5555555555"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	order := &Order{UserID: user.ID, Amount: 200.0}
	err = db.Create(order).Error
	assert.NoError(t, err)

	type UserWithOrder struct {
		ID     uint64
		Name   string
		Amount float64
	}
	var results []UserWithOrder
	err = xdb.Model(&User{}).Select("users.id, users.name, orders.amount").
		Joins("INNER JOIN orders ON orders.user_id = users.id").
		Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestTable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "tableTest", Age: 30, Email: "table@test.com", Phone: "6666666666"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	type SimpleUser struct {
		ID   uint64
		Name string
	}
	var results []SimpleUser
	err = xdb.Table("users").Select("id, name").Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "modelTest", Age: 30, Email: "model@test.com", Phone: "7777777777"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = xdb.Model(&User{}).First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUnscoped(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "unscopedTest", Age: 30, Email: "unscoped@test.com", Phone: "8888888888"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	err = xdb.Delete(context.Background(), &User{}, user.ID)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.Error(t, err)

	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestOmit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "omitTest", Age: 25, Email: "omit@test.com", Phone: "9999999999"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	err = xdb.Model(&User{}).Where("id = ?", user.ID).Omit("name").Updates(context.Background(), &User{Name: "newName", Age: 99})
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "omitTest", found.Name)
	assert.Equal(t, 99, found.Age)
}

func TestBuild(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "buildTest", Age: 30, Email: "build@test.com", Phone: "0000000000"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = xdb.Build(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", user.ID)
	}).First(context.Background(), &found)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestChainOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	var users []User
	err := xdb.Where("age > ?", 22).Order("age desc").Limit(3).Offset(1).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestChainPreservesBreaker(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	chained := xdb.Where("age > ?", 22).Order("age desc").Limit(3)
	assert.Equal(t, xdb.brk, chained.brk)
	assert.NotNil(t, chained.acceptable)
}

func TestNewDB_DefaultBreaker(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	xdb := NewDB(gdb)
	assert.NotNil(t, xdb)
	assert.NotNil(t, xdb.brk)
	assert.NotNil(t, xdb.exec)

	assert.True(t, xdb.acceptable(gorm.ErrRecordNotFound))
	assert.True(t, xdb.acceptable(gorm.ErrInvalidTransaction))
	assert.False(t, xdb.acceptable(errors.New("other error")))
}

func TestNewDB_WithCustomBreaker(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	customBrk := breaker.NewBreaker(
		breaker.WithName("custom-gormx-breaker"),
		breaker.WithK(2.0),
		breaker.WithProtection(100),
	)

	xdb := NewDB(gdb, WithBreaker(customBrk))
	assert.NotNil(t, xdb)
	assert.Equal(t, customBrk, xdb.brk)
}

func TestNewDB_WithCustomAcceptable(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	customErr := errors.New("custom acceptable error")
	customAcceptable := func(err error) bool {
		return err == customErr
	}

	xdb := NewDB(gdb, WithAcceptable(customAcceptable))
	assert.NotNil(t, xdb)
	assert.True(t, xdb.acceptable(customErr))
	assert.False(t, xdb.acceptable(gorm.ErrRecordNotFound))
}

func TestDB_BreakerMetrics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Create(context.Background(), &User{Name: "test", Email: "test@test.com", Phone: "1234567890"})
	assert.NoError(t, err)

	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

type FailedModel struct {
	Name string `gorm:"column:name"`
}

func TestDB_BreakerFailure(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Create(context.Background(), &FailedModel{Name: "test"})
	assert.Error(t, err)

	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(0), accepts)
	assert.Equal(t, 0.0, rate)
}

func TestDB_BreakerAcceptableError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	assert.NotNil(t, xdb)

	err := xdb.brk.DoWithAcceptable(context.Background(), func(ctx context.Context) error {
		return db.WithContext(ctx).First(&User{}).Error
	}, xdb.acceptable)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(1), accepts)
	assert.Equal(t, 1.0, rate)
}

func TestDB_BreakerReset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.Create(context.Background(), &FailedModel{Name: "test"})
	assert.Error(t, err)

	total, accepts, _ := xdb.brk.GetMetrics()
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(0), accepts)

	xdb.brk.Reset()

	total, accepts, rate := xdb.brk.GetMetrics()
	assert.Equal(t, int64(0), total)
	assert.Equal(t, int64(0), accepts)
	assert.Equal(t, float64(0), rate)
}

func TestMapFilter_EmptyMap(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var users []User
	err := xdb.MapFilter(map[string]any{}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestStructFilter_ZeroStruct(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var users []User
	err := xdb.StructFilter(&User{}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestRaw(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "rawTest", Age: 30, Email: "raw@test.com", Phone: "1111111112"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	type SimpleUser struct {
		ID   uint64
		Name string
	}
	var results []SimpleUser
	err = xdb.Raw("SELECT id, name FROM users WHERE id = ?", user.ID).Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, user.ID, results[0].ID)
}

func TestClauses(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var users []User
	err := xdb.Clauses(clause.Locking{Strength: "UPDATE"}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestFindInBatches(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	var allUsers []User
	var batchNums []int
	err := xdb.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		batchNums = append(batchNums, batch)
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 10)
	assert.Equal(t, []int{1, 2, 3, 4}, batchNums)
}

func TestFindInBatches_BatchSize(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	var batchSizes []int
	err := xdb.FindInBatches(context.Background(), 4, func(tx *Executor, batch int, dest *[]User) error {
		batchSizes = append(batchSizes, len(*dest))
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int{4, 4, 2}, batchSizes)
}

func TestFindInBatches_ExactMultiple(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 9)

	var batchSizes []int
	err := xdb.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		batchSizes = append(batchSizes, len(*dest))
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int{3, 3, 3}, batchSizes)
}

func TestFindInBatches_SingleBatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 3)

	var batchSizes []int
	err := xdb.FindInBatches(context.Background(), 10, func(tx *Executor, batch int, dest *[]User) error {
		batchSizes = append(batchSizes, len(*dest))
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int{3}, batchSizes)
}

func TestFindInBatches_EmptyResult(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)

	var allUsers []User
	invoked := false
	err := xdb.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		invoked = true
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.False(t, invoked)
	assert.Len(t, allUsers, 0)
}

func TestFindInBatches_WithWhere(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	var allUsers []User
	err := xdb.Where("age > ?", 23).FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, allUsers)
	for _, u := range allUsers {
		assert.Greater(t, u.Age, 23)
	}
}

func TestFindInBatches_WithLimit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	var allUsers []User
	err := xdb.Limit(5).FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 5)
}

func TestFindInBatches_WithOffset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	users := seedUsers(t, xdb, 10)

	var allUsers []User
	err := xdb.Order("id asc").Offset(3).FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 7)
	assert.Equal(t, users[3].ID, allUsers[0].ID)
}

func TestFindInBatches_CallbackError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	expectedErr := errors.New("callback error")
	var processedCount int
	err := xdb.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		processedCount++
		if batch == 2 {
			return expectedErr
		}
		return nil
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFindFailed))
	assert.Equal(t, 2, processedCount)
}

func TestFindInBatches_CallbackCanUseTx(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 10)

	err := xdb.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		for _, u := range *dest {
			updateErr := tx.Model(&User{}).Where("id = ?", u.ID).Update(context.Background(), "age", 100)
			if updateErr != nil {
				return updateErr
			}
		}
		return nil
	})
	assert.NoError(t, err)

	var count int64
	err = xdb.Model(&User{}).Where("age = ?", 100).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestFindInBatches_BatchNumbersSequential(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 15)

	var batchNums []int
	err := xdb.FindInBatches(context.Background(), 4, func(tx *Executor, batch int, dest *[]User) error {
		batchNums = append(batchNums, batch)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4}, batchNums)
}

func TestFindInBatches_BatchSizeOne(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var allUsers []User
	var batchNums []int
	err := xdb.FindInBatches(context.Background(), 1, func(tx *Executor, batch int, dest *[]User) error {
		batchNums = append(batchNums, batch)
		assert.Len(t, *dest, 1)
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 5)
	assert.Len(t, batchNums, 5)
}

func TestFindInBatches_ModelSpecific(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var allUsers []User
	err := xdb.Model(&User{}).FindInBatches(context.Background(), 2, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 5)
}

func TestFindInBatches_SelectSpecificColumns(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	seedUsers(t, xdb, 5)

	var allUsers []User
	err := xdb.Select("id", "name").FindInBatches(context.Background(), 2, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 5)
	for _, u := range allUsers {
		assert.NotZero(t, u.ID)
		assert.NotEmpty(t, u.Name)
		assert.Zero(t, u.Age)
	}
}

func TestUpdatesByStruct(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "updByStruct", Age: 25, Email: "updstruct@test.com", Phone: "1111111111"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "updByStruct"}
	updateData := &User{Name: "updatedStruct", Age: 30}
	err = xdb.UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updatedStruct", found.Name)
	assert.Equal(t, 30, found.Age)
}

func TestUpdatesByStruct_NilUpdateData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.UpdatesByStruct(context.Background(), &User{Name: "test"}, (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestUpdatesByStruct_NilFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.UpdatesByStruct(context.Background(), (*User)(nil), &User{Name: "test"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestUpdatesByStruct_IgnoresZeroValues(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "zeroValStruct", Age: 50, Email: "zeroval@test.com", Phone: "2222222222"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "zeroValStruct"}
	updateData := &User{Name: "newNameForZero", Age: 0}
	err = xdb.UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newNameForZero", found.Name)
	assert.Equal(t, 50, found.Age)
}

func TestUpdatesByStruct_WithModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "modelStruct", Age: 25, Email: "modelstruct@test.com", Phone: "3333333333"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "modelStruct"}
	updateData := &User{Age: 99}
	err = xdb.Model(&User{}).UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 99, found.Age)
}

func TestUpdatesByStruct_NoMatchingFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "noMatchStruct", Age: 25, Email: "nomatch@test.com", Phone: "4444444444"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "nonExistentName"}
	updateData := &User{Age: 100}
	err = xdb.UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 25, found.Age)
}

func TestUpdatesByMap(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "updByMap", Age: 25, Email: "updmap@test.com", Phone: "5555555555"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "updByMap"}
	updateData := &User{Name: "updatedMap", Age: 30}
	err = xdb.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updatedMap", found.Name)
	assert.Equal(t, 30, found.Age)
}

func TestUpdatesByMap_NilUpdateData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.UpdatesByMap(context.Background(), map[string]any{"name": "test"}, (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestUpdatesByMap_EmptyFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.UpdatesByMap(context.Background(), map[string]any{}, &User{Name: "test"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestUpdatesByMap_ZeroValueUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "mapZeroVal", Age: 50, Email: "mapzero@test.com", Phone: "6666666666"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "mapZeroVal"}
	updateData := &User{Name: "newNameMapZero", Age: 0}
	err = xdb.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newNameMapZero", found.Name)
	assert.Equal(t, 50, found.Age)
}

func TestUpdatesByMap_MultipleFilterConditions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user1 := &User{Name: "multiMap", Age: 25, Email: "multi1@test.com", Phone: "7777777771"}
	user2 := &User{Name: "multiMap", Age: 30, Email: "multi2@test.com", Phone: "7777777772"}
	assert.NoError(t, xdb.Create(context.Background(), user1))
	assert.NoError(t, xdb.Create(context.Background(), user2))

	filter := map[string]any{"name": "multiMap", "age": 25}
	updateData := &User{Age: 35}
	err := xdb.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user1.ID)
	assert.NoError(t, err)
	assert.Equal(t, 35, found.Age)

	var found2 User
	err = xdb.First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.Equal(t, 30, found2.Age)
}

func TestUpdatesByMap_WithModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "modelMap", Age: 25, Email: "modelmap@test.com", Phone: "8888888888"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "modelMap"}
	updateData := &User{Age: 99}
	err = xdb.Model(&User{}).UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 99, found.Age)
}

func TestUpdatesByMap_NoMatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "noMatchMap", Age: 25, Email: "nomatchmap@test.com", Phone: "9999999999"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "nonExistent"}
	updateData := &User{Age: 100}
	err = xdb.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = xdb.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 25, found.Age)
}

func TestDeleteByStruct(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "delByStruct", Age: 25, Email: "delstruct@test.com", Phone: "0000000001"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "delByStruct"}
	err = xdb.DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestDeleteByStruct_NilFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.DeleteByStruct(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestDeleteByStruct_NoMatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "noDelStruct", Age: 25, Email: "nodelstruct@test.com", Phone: "0000000002"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "nonExistentName"}
	err = xdb.DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.False(t, found.DeletedAt.Valid)
}

func TestDeleteByStruct_WithModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "modelDelStruct", Age: 25, Email: "modeldelstruct@test.com", Phone: "0000000003"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "modelDelStruct"}
	err = xdb.Model(&User{}).DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestDeleteByStruct_MultipleMatches(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user1 := &User{Name: "batchDelStruct", Age: 25, Email: "batch1@test.com", Phone: "0000000004"}
	user2 := &User{Name: "batchDelStruct", Age: 30, Email: "batch2@test.com", Phone: "0000000005"}
	assert.NoError(t, xdb.Create(context.Background(), user1))
	assert.NoError(t, xdb.Create(context.Background(), user2))

	filter := &User{Name: "batchDelStruct"}
	err := xdb.DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found1 User
	err = xdb.Unscoped().First(context.Background(), &found1, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found1.DeletedAt)

	var found2 User
	err = xdb.Unscoped().First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found2.DeletedAt)
}

func TestDeleteByMap(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "delByMap", Age: 25, Email: "delmap@test.com", Phone: "0000000006"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "delByMap"}
	err = xdb.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestDeleteByMap_EmptyFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	err := xdb.DeleteByMap[User](context.Background(), map[string]any{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestDeleteByMap_NoMatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "noDelMap", Age: 25, Email: "nodelmap@test.com", Phone: "0000000007"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "nonExistent"}
	err = xdb.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.False(t, found.DeletedAt.Valid)
}

func TestDeleteByMap_MultipleFilterConditions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user1 := &User{Name: "multiDelMap", Age: 25, Email: "multi1@test.com", Phone: "0000000008"}
	user2 := &User{Name: "multiDelMap", Age: 30, Email: "multi2@test.com", Phone: "0000000009"}
	assert.NoError(t, xdb.Create(context.Background(), user1))
	assert.NoError(t, xdb.Create(context.Background(), user2))

	filter := map[string]any{"name": "multiDelMap", "age": 25}
	err := xdb.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found1 User
	err = xdb.Unscoped().First(context.Background(), &found1, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found1.DeletedAt)

	var found2 User
	err = xdb.Unscoped().First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.False(t, found2.DeletedAt.Valid)
}

func TestDeleteByMap_WithModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user := &User{Name: "modelDelMap", Age: 25, Email: "modeldelmap@test.com", Phone: "0000000010"}
	err := xdb.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "modelDelMap"}
	err = xdb.Model(&User{}).DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = xdb.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestDeleteByMap_BatchDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	xdb := NewDB(db)
	user1 := &User{Name: "batchDelMap", Age: 25, Email: "batch1@test.com", Phone: "0000000011"}
	user2 := &User{Name: "batchDelMap", Age: 30, Email: "batch2@test.com", Phone: "0000000012"}
	assert.NoError(t, xdb.Create(context.Background(), user1))
	assert.NoError(t, xdb.Create(context.Background(), user2))

	filter := map[string]any{"name": "batchDelMap"}
	err := xdb.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found1 User
	err = xdb.Unscoped().First(context.Background(), &found1, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found1.DeletedAt)

	var found2 User
	err = xdb.Unscoped().First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found2.DeletedAt)
}
