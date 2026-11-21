package gormx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func setupExecutorTest(t *testing.T) (*gorm.DB, func()) {
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

func seedExecutorUsers(t *testing.T, exec *Executor, count int) []*User {
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
	err := exec.CreateInBatches(context.Background(), &users, 100)
	assert.NoError(t, err)
	return users
}

func TestExecutor_NewExecutor(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	assert.NotNil(t, exec)
	assert.NotNil(t, exec.db)
}

func TestExecutor_Create(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execCreate", Age: 25, Email: "execCreate@test.com", Phone: "1234567890"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "execCreate", found.Name)
}

func TestExecutor_Create_NilModel(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Create(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_CreateInBatches(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	users := seedExecutorUsers(t, exec, 5)

	var found []User
	err := exec.Find(context.Background(), &found)
	assert.NoError(t, err)
	assert.Len(t, found, 5)
	assert.Equal(t, users[0].ID, found[0].ID)
}

func TestExecutor_CreateInBatches_NilModels(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.CreateInBatches(context.Background(), (*[]User)(nil), 100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_First(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execFirst", Age: 30, Email: "execFirst@test.com", Phone: "1111111111"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "execFirst", found.Name)
}

func TestExecutor_First_WithConds(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var found User
	err := exec.First(context.Background(), &found, "name = ?", "userB")
	assert.NoError(t, err)
	assert.Equal(t, "userB", found.Name)
}

func TestExecutor_First_NilDest(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.First(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_First_NotFound(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	var found User
	err := exec.First(context.Background(), &found, 99999)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFirstFailed))
}

func TestExecutor_Scan(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.Model(&User{}).Scan(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestExecutor_Scan_NilDest(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Scan(context.Background(), (*[]User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_Pluck_Names(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var names []string
	err := exec.Model(&User{}).Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.Len(t, names, 5)
	assert.Contains(t, names, "userA")
	assert.Contains(t, names, "userE")
}

func TestExecutor_Pluck_IDs(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	users := seedExecutorUsers(t, exec, 5)

	var ids []uint64
	err := exec.Model(&User{}).Pluck(context.Background(), "id", &ids)
	assert.NoError(t, err)
	assert.Len(t, ids, 5)
	for _, u := range users {
		assert.Contains(t, ids, u.ID)
	}
}

func TestExecutor_Pluck_Ages(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var ages []int
	err := exec.Model(&User{}).Pluck(context.Background(), "age", &ages)
	assert.NoError(t, err)
	assert.Len(t, ages, 5)
	for _, age := range ages {
		assert.Greater(t, age, 0)
	}
}

func TestExecutor_Pluck_WithWhere(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var names []string
	err := exec.Model(&User{}).Where("age > ?", 22).Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.NotEmpty(t, names)
	assert.Less(t, len(names), 5)
}

func TestExecutor_Pluck_WithLimit(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var names []string
	err := exec.Model(&User{}).Limit(3).Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.Len(t, names, 3)
}

func TestExecutor_Pluck_WithOrder(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var ages []int
	err := exec.Model(&User{}).Order("age desc").Pluck(context.Background(), "age", &ages)
	assert.NoError(t, err)
	assert.Len(t, ages, 5)
	for i := 1; i < len(ages); i++ {
		assert.GreaterOrEqual(t, ages[i-1], ages[i])
	}
}

func TestExecutor_Pluck_WithLimitAndOffset(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var names []string
	err := exec.Model(&User{}).Order("id asc").Offset(2).Limit(2).Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.Len(t, names, 2)
}

func TestExecutor_Pluck_EmptyResult(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)

	var names []string
	err := exec.Model(&User{}).Where("age > ?", 999).Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.Empty(t, names)
}

func TestExecutor_Pluck_NilDest(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Model(&User{}).Pluck(context.Background(), "name", (*[]string)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_Pluck_InvalidColumn(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var result []string
	err := exec.Model(&User{}).Pluck(context.Background(), "non_existent_column", &result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPluckFailed))
}

func TestExecutor_Pluck_WithoutModel(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var names []string
	err := exec.Table("users").Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.Len(t, names, 3)
}

func TestExecutor_Pluck_ChainOperations(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var names []string
	err := exec.Model(&User{}).Where("age > ?", 22).Order("age desc").Limit(3).Pluck(context.Background(), "name", &names)
	assert.NoError(t, err)
	assert.Len(t, names, 3)
}

func TestExecutor_Find(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestExecutor_Find_WithWhere(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.Where("age > ?", 22).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestExecutor_Find_NilDest(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Find(context.Background(), (*[]User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_FindInBatches(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var allUsers []User
	var batchNums []int
	err := exec.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		batchNums = append(batchNums, batch)
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 10)
	assert.Equal(t, []int{1, 2, 3, 4}, batchNums)
}

func TestExecutor_FindInBatches_BatchSize(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var batchSizes []int
	err := exec.FindInBatches(context.Background(), 4, func(tx *Executor, batch int, dest *[]User) error {
		batchSizes = append(batchSizes, len(*dest))
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int{4, 4, 2}, batchSizes)
}

func TestExecutor_FindInBatches_EmptyResult(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)

	var allUsers []User
	invoked := false
	err := exec.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		invoked = true
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.False(t, invoked)
	assert.Len(t, allUsers, 0)
}

func TestExecutor_FindInBatches_WithWhere(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var allUsers []User
	err := exec.Where("age > ?", 23).FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, allUsers)
	for _, u := range allUsers {
		assert.Greater(t, u.Age, 23)
	}
}

func TestExecutor_FindInBatches_WithLimit(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var allUsers []User
	err := exec.Limit(5).FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
		allUsers = append(allUsers, *dest...)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 5)
}

func TestExecutor_FindInBatches_CallbackError(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	expectedErr := errors.New("callback error")
	var processedCount int
	err := exec.FindInBatches(context.Background(), 3, func(tx *Executor, batch int, dest *[]User) error {
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

func TestExecutor_Count(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var count int64
	err := exec.Model(&User{}).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestExecutor_Count_WithWhere(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var count int64
	err := exec.Model(&User{}).Where("age > ?", 22).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Greater(t, count, int64(0))
}

func TestExecutor_Count_NilCount(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Count(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_Update(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execUpdate", Age: 25, Email: "execUpdate@test.com", Phone: "2222222222"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	err = exec.Model(&User{}).Where("id = ?", user.ID).Update(context.Background(), "name", "updatedName")
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updatedName", found.Name)
}

func TestExecutor_Updates(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execUpdates", Age: 25, Email: "execUpdates@test.com", Phone: "3333333333"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	err = exec.Model(&User{}).Where("id = ?", user.ID).Updates(context.Background(), &User{Name: "updatedName", Age: 30})
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updatedName", found.Name)
	assert.Equal(t, 30, found.Age)
}

func TestExecutor_Updates_NilModel(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Updates(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_UpdatesByStruct(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execUpdStruct", Age: 25, Email: "execUpdStruct@test.com", Phone: "4444444444"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "execUpdStruct"}
	updateData := &User{Name: "updatedStruct", Age: 30}
	err = exec.UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updatedStruct", found.Name)
	assert.Equal(t, 30, found.Age)
}

func TestExecutor_UpdatesByStruct_NilUpdateData(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.UpdatesByStruct(context.Background(), &User{Name: "test"}, (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_UpdatesByStruct_NilFilter(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.UpdatesByStruct(context.Background(), (*User)(nil), &User{Name: "test"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestExecutor_UpdatesByStruct_IgnoresZeroValues(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execZeroVal", Age: 50, Email: "execZeroVal@test.com", Phone: "5555555555"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "execZeroVal"}
	updateData := &User{Name: "newName", Age: 0}
	err = exec.UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newName", found.Name)
	assert.Equal(t, 50, found.Age)
}

func TestExecutor_UpdatesByStruct_NoMatch(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execNoMatch", Age: 25, Email: "execNoMatch@test.com", Phone: "6666666666"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "nonExistent"}
	updateData := &User{Age: 100}
	err = exec.UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 25, found.Age)
}

func TestExecutor_UpdatesByStruct_WithModel(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execModelStruct", Age: 25, Email: "execModelStruct@test.com", Phone: "7777777777"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "execModelStruct"}
	updateData := &User{Age: 99}
	err = exec.Model(&User{}).UpdatesByStruct(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 99, found.Age)
}

func TestExecutor_UpdatesByMap(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execUpdMap", Age: 25, Email: "execUpdMap@test.com", Phone: "8888888888"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "execUpdMap"}
	updateData := &User{Name: "updatedMap", Age: 30}
	err = exec.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updatedMap", found.Name)
	assert.Equal(t, 30, found.Age)
}

func TestExecutor_UpdatesByMap_NilUpdateData(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.UpdatesByMap(context.Background(), map[string]any{"name": "test"}, (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_UpdatesByMap_EmptyFilter(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.UpdatesByMap(context.Background(), map[string]any{}, &User{Name: "test"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestExecutor_UpdatesByMap_ZeroValueUpdate(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execMapZero", Age: 50, Email: "execMapZero@test.com", Phone: "9999999999"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "execMapZero"}
	updateData := &User{Name: "newNameMap", Age: 0}
	err = exec.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newNameMap", found.Name)
	assert.Equal(t, 50, found.Age)
}

func TestExecutor_UpdatesByMap_MultipleFilterConditions(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user1 := &User{Name: "execMultiMap", Age: 25, Email: "multi1@test.com", Phone: "1010101010"}
	user2 := &User{Name: "execMultiMap", Age: 30, Email: "multi2@test.com", Phone: "1010101011"}
	assert.NoError(t, exec.Create(context.Background(), user1))
	assert.NoError(t, exec.Create(context.Background(), user2))

	filter := map[string]any{"name": "execMultiMap", "age": 25}
	updateData := &User{Age: 35}
	err := exec.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user1.ID)
	assert.NoError(t, err)
	assert.Equal(t, 35, found.Age)

	var found2 User
	err = exec.First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.Equal(t, 30, found2.Age)
}

func TestExecutor_UpdatesByMap_NoMatch(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execNoMatchMap", Age: 25, Email: "execNoMatchMap@test.com", Phone: "1212121212"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "nonExistent"}
	updateData := &User{Age: 100}
	err = exec.UpdatesByMap(context.Background(), filter, updateData)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 25, found.Age)
}

func TestExecutor_Delete(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execDelete", Age: 30, Email: "execDelete@test.com", Phone: "1313131313"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	err = exec.Delete(context.Background(), &User{}, user.ID)
	assert.NoError(t, err)

	var found User
	err = exec.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestExecutor_Delete_NilModel(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Delete(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidModel))
}

func TestExecutor_DeleteByStruct(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execDelStruct", Age: 25, Email: "execDelStruct@test.com", Phone: "1414141414"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "execDelStruct"}
	err = exec.DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = exec.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestExecutor_DeleteByStruct_NilFilter(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.DeleteByStruct(context.Background(), (*User)(nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestExecutor_DeleteByStruct_NoMatch(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execNoDelStruct", Age: 25, Email: "execNoDelStruct@test.com", Phone: "1515151515"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := &User{Name: "nonExistent"}
	err = exec.DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = exec.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.False(t, found.DeletedAt.Valid)
}

func TestExecutor_DeleteByStruct_MultipleMatches(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user1 := &User{Name: "execBatchDel", Age: 25, Email: "batch1@test.com", Phone: "1616161616"}
	user2 := &User{Name: "execBatchDel", Age: 30, Email: "batch2@test.com", Phone: "1616161617"}
	assert.NoError(t, exec.Create(context.Background(), user1))
	assert.NoError(t, exec.Create(context.Background(), user2))

	filter := &User{Name: "execBatchDel"}
	err := exec.DeleteByStruct(context.Background(), filter)
	assert.NoError(t, err)

	var found1 User
	err = exec.Unscoped().First(context.Background(), &found1, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found1.DeletedAt)

	var found2 User
	err = exec.Unscoped().First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found2.DeletedAt)
}

func TestExecutor_DeleteByMap(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execDelMap", Age: 25, Email: "execDelMap@test.com", Phone: "1717171717"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "execDelMap"}
	err = exec.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = exec.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestExecutor_DeleteByMap_EmptyFilter(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.DeleteByMap[User](context.Background(), map[string]any{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidFilter))
}

func TestExecutor_DeleteByMap_NoMatch(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execNoDelMap", Age: 25, Email: "execNoDelMap@test.com", Phone: "1818181818"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	filter := map[string]any{"name": "nonExistent"}
	err = exec.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found User
	err = exec.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.False(t, found.DeletedAt.Valid)
}

func TestExecutor_DeleteByMap_MultipleFilterConditions(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user1 := &User{Name: "execMultiDelMap", Age: 25, Email: "multi1@test.com", Phone: "1919191919"}
	user2 := &User{Name: "execMultiDelMap", Age: 30, Email: "multi2@test.com", Phone: "1919191920"}
	assert.NoError(t, exec.Create(context.Background(), user1))
	assert.NoError(t, exec.Create(context.Background(), user2))

	filter := map[string]any{"name": "execMultiDelMap", "age": 25}
	err := exec.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found1 User
	err = exec.Unscoped().First(context.Background(), &found1, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found1.DeletedAt)

	var found2 User
	err = exec.Unscoped().First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.False(t, found2.DeletedAt.Valid)
}

func TestExecutor_DeleteByMap_BatchDelete(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user1 := &User{Name: "execBatchDelMap", Age: 25, Email: "batch1@test.com", Phone: "2020202020"}
	user2 := &User{Name: "execBatchDelMap", Age: 30, Email: "batch2@test.com", Phone: "2020202021"}
	assert.NoError(t, exec.Create(context.Background(), user1))
	assert.NoError(t, exec.Create(context.Background(), user2))

	filter := map[string]any{"name": "execBatchDelMap"}
	err := exec.DeleteByMap[User](context.Background(), filter)
	assert.NoError(t, err)

	var found1 User
	err = exec.Unscoped().First(context.Background(), &found1, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found1.DeletedAt)

	var found2 User
	err = exec.Unscoped().First(context.Background(), &found2, user2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found2.DeletedAt)
}

func TestExecutor_Exec(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	err := exec.Exec(context.Background(), "INSERT INTO users (name, age, email, phone, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"execRaw", 30, "execRaw@test.com", "2121212121", time.Now(), time.Now())
	assert.NoError(t, err)

	var count int64
	err = exec.Model(&User{}).Where("name = ?", "execRaw").Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestExecutor_Transaction_Commit(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)

	err := exec.Transaction(context.Background(), func(tx *Executor) error {
		u := &User{Name: "execTxUser", Age: 20, Email: "execTx@test.com", Phone: "2222222223"}
		return tx.Create(context.Background(), u)
	})
	assert.NoError(t, err)

	var count int64
	err = exec.Model(&User{}).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestExecutor_Transaction_Rollback(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)

	expectedErr := errors.New("rollback on purpose")
	err := exec.Transaction(context.Background(), func(tx *Executor) error {
		u := &User{Name: "execTxRollback", Age: 20, Email: "execTxRollback@test.com", Phone: "2323232323"}
		createErr := tx.Create(context.Background(), u)
		if createErr != nil {
			return createErr
		}
		return expectedErr
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))

	var count int64
	err = exec.Model(&User{}).Count(context.Background(), &count)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestExecutor_Transaction_UpdateInside(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execTxUpdate", Age: 20, Email: "execTxUpdate@test.com", Phone: "2424242424"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	err = exec.Transaction(context.Background(), func(tx *Executor) error {
		var u User
		if err := tx.First(context.Background(), &u, user.ID); err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", u.ID).Update(context.Background(), "age", 100)
	})
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 100, found.Age)
}

func TestExecutor_Build(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execBuild", Age: 30, Email: "execBuild@test.com", Phone: "2525252525"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = exec.Build(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", user.ID)
	}).First(context.Background(), &found)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestExecutor_Model(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execModel", Age: 30, Email: "execModel@test.com", Phone: "2626262626"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = exec.Model(&User{}).First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestExecutor_Table(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execTable", Age: 30, Email: "execTable@test.com", Phone: "2727272727"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	type SimpleUser struct {
		ID   uint64
		Name string
	}
	var results []SimpleUser
	err = exec.Table("users").Select("id, name").Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestExecutor_Raw(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execRaw", Age: 30, Email: "execRaw@test.com", Phone: "2828282828"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	type SimpleUser struct {
		ID   uint64
		Name string
	}
	var results []SimpleUser
	err = exec.Raw("SELECT id, name FROM users WHERE id = ?", user.ID).Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, user.ID, results[0].ID)
}

func TestExecutor_Select(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execSelect", Age: 99, Email: "execSelect@test.com", Phone: "2929292929"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	var found User
	err = exec.Select("name", "age").First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "execSelect", found.Name)
	assert.Equal(t, 99, found.Age)
}

func TestExecutor_Where(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.Where("age > ?", 22).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestExecutor_StructFilter(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.StructFilter(&User{Age: 21}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Equal(t, 21, u.Age)
	}
}

func TestExecutor_MapFilter(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.MapFilter(map[string]any{"age": 22}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	for _, u := range users {
		assert.Equal(t, 22, u.Age)
	}
}

func TestExecutor_Order(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var users []User
	err := exec.Order("age desc").Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.GreaterOrEqual(t, users[0].Age, users[1].Age)
	assert.GreaterOrEqual(t, users[1].Age, users[2].Age)
}

func TestExecutor_OrderByColumn(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var users []User
	err := exec.OrderByColumn(clause.OrderByColumn{Column: clause.Column{Name: "age"}, Desc: true}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.GreaterOrEqual(t, users[0].Age, users[1].Age)
}

func TestExecutor_OrderBy(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	var users []User
	err := exec.OrderBy(clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "age"}, Desc: true}}}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.GreaterOrEqual(t, users[0].Age, users[1].Age)
}

func TestExecutor_Joins(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execJoin", Age: 30, Email: "execJoin@test.com", Phone: "3030303030"}
	err := exec.Create(context.Background(), user)
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
	err = exec.Model(&User{}).Select("users.id, users.name, orders.amount").
		Joins("JOIN orders ON orders.user_id = users.id").
		Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, user.Name, results[0].Name)
}

func TestExecutor_InnerJoins(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execInnerJoin", Age: 30, Email: "execInnerJoin@test.com", Phone: "3131313131"}
	err := exec.Create(context.Background(), user)
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
	err = exec.Model(&User{}).Select("users.id, users.name, orders.amount").
		Joins("INNER JOIN orders ON orders.user_id = users.id").
		Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestExecutor_Limit(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var users []User
	err := exec.Limit(3).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestExecutor_Offset(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	users := seedExecutorUsers(t, exec, 5)

	var result []User
	err := exec.Order("id asc").Offset(2).Limit(2).Find(context.Background(), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, users[2].ID, result[0].ID)
}

func TestExecutor_Unscoped(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execUnscoped", Age: 30, Email: "execUnscoped@test.com", Phone: "3232323232"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	err = exec.Delete(context.Background(), &User{}, user.ID)
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.Error(t, err)

	err = exec.Unscoped().First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestExecutor_Omit(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user := &User{Name: "execOmit", Age: 25, Email: "execOmit@test.com", Phone: "3333333334"}
	err := exec.Create(context.Background(), user)
	assert.NoError(t, err)

	err = exec.Model(&User{}).Where("id = ?", user.ID).Omit("name").Updates(context.Background(), &User{Name: "newName", Age: 99})
	assert.NoError(t, err)

	var found User
	err = exec.First(context.Background(), &found, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "execOmit", found.Name)
	assert.Equal(t, 99, found.Age)
}

func TestExecutor_Group(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 3)

	type AgeCount struct {
		Age   int
		Count int
	}
	var results []AgeCount
	err := exec.Model(&User{}).Select("age, count(*) as count").Group("age").Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestExecutor_Having(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	user1 := &User{Name: "execHavingA", Age: 25, Email: "a@test.com", Phone: "3434343434"}
	user2 := &User{Name: "execHavingB", Age: 25, Email: "b@test.com", Phone: "3434343435"}
	user3 := &User{Name: "execHavingC", Age: 30, Email: "c@test.com", Phone: "3434343436"}
	assert.NoError(t, exec.Create(context.Background(), user1))
	assert.NoError(t, exec.Create(context.Background(), user2))
	assert.NoError(t, exec.Create(context.Background(), user3))

	type AgeCount struct {
		Age   int
		Count int
	}
	var results []AgeCount
	err := exec.Model(&User{}).Select("age, count(*) as count").Group("age").Having("count > ?", 1).Find(context.Background(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 25, results[0].Age)
}

func TestExecutor_Clauses(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 5)

	var users []User
	err := exec.Clauses(clause.Locking{Strength: "UPDATE"}).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestExecutor_ChainOperations(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	seedExecutorUsers(t, exec, 10)

	var users []User
	err := exec.Where("age > ?", 22).Order("age desc").Limit(3).Offset(1).Find(context.Background(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	for _, u := range users {
		assert.Greater(t, u.Age, 22)
	}
}

func TestExecutor_ChainReturnsNewExecutor(t *testing.T) {
	db, cleanup := setupExecutorTest(t)
	defer cleanup()

	exec := NewExecutor(db)
	chained := exec.Where("age > ?", 22)
	assert.NotEqual(t, exec, chained)
	assert.NotNil(t, chained)
	assert.NotNil(t, chained.db)
}
