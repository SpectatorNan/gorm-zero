package pagex

import (
	"context"
	"fmt"
	"testing"

	"github.com/SpectatorNan/gorm-zero/gormc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestUserModel struct {
	Id       uint64 `gorm:"column:id;primary_key"`
	Age      int8   `gorm:"column:age"`
	Name     string `gorm:"column:name"`     // The username
	Nickname string `gorm:"column:nickname"` // The nickname
	Avatar   string `gorm:"column:avatar"`
	Email    string `gorm:"column:email"`
}

func (TestUserModel) TableName() string {
	return "user"
}

// createMockDB 创建内存SQLite数据库用于测试
func createMockDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&TestUserModel{})
	if err != nil {
		return nil, err
	}

	// 插入测试数据
	testUsers := []TestUserModel{
		{Id: 1, Age: 25, Name: "Alice", Nickname: "A", Email: "alice@test.com"},
		{Id: 2, Age: 30, Name: "Bob", Nickname: "B", Email: "bob@test.com"},
		{Id: 3, Age: 20, Name: "Charlie", Nickname: "C", Email: "charlie@test.com"},
		{Id: 4, Age: 35, Name: "David", Nickname: "D", Email: "david@test.com"},
		{Id: 5, Age: 28, Name: "Eva", Nickname: "E", Email: "eva@test.com"},
		{Id: 6, Age: 22, Name: "Frank", Nickname: "F", Email: "frank@test.com"},
		{Id: 7, Age: 32, Name: "Grace", Nickname: "G", Email: "grace@test.com"},
		{Id: 8, Age: 27, Name: "Henry", Nickname: "H", Email: "henry@test.com"},
	}

	for _, user := range testUsers {
		db.Create(&user)
	}

	return db, nil
}

// MockGormcCacheConn 实现 GormcCacheConn 接口用于测试
type MockGormcCacheConn struct {
	db *gorm.DB
}

func NewMockGormcCacheConn(db *gorm.DB) *MockGormcCacheConn {
	return &MockGormcCacheConn{db: db}
}

func (m *MockGormcCacheConn) QueryNoCacheCtx(_ context.Context, fn gormc.QueryCtxFn) error {
	return fn(m.db)
}

func (m *MockGormcCacheConn) ExecNoCacheCtx(_ context.Context, execCtx gormc.ExecCtxFn) error {
	return execCtx(m.db)
}

func TestFindPageList(t *testing.T) {
	db, err := createMockDB()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	// 使用 MockGormcCacheConn 而不是真实的 Redis 连接
	mockConn := NewMockGormcCacheConn(db)
	orderKeys := map[string]string{
		"age": "age",
		"id":  "id",
	}

	users, cnt, err := FindPageList[TestUserModel](
		context.Background(),
		mockConn,
		&ListReq{Page: 1, PageSize: 5},
		OrderBy{OrderKey: "age", Sort: "asc"},
		orderKeys,
		func(conn *gorm.DB) (*gorm.DB, *gorm.DB) {
			d := conn.Model(&TestUserModel{})
			return d, d
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageList Err,%v", err.Error())
	}

	if cnt != 8 {
		t.Errorf("Expected count 8, got %d", cnt)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}

	fmt.Printf("Users: %+v, Count: %d\n", users, cnt)
}

func TestFindPageListWithCount(t *testing.T) {
	db, err := createMockDB()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	orderKeys := map[string]string{
		"age": "age",
		"id":  "id",
	}

	users, cnt, err := FindPageListWithCount[TestUserModel](
		context.Background(),
		&ListReq{Page: 1, PageSize: 5},
		OrderBy{OrderKey: "age", Sort: "asc"},
		orderKeys,
		func() (*gorm.DB, *gorm.DB) {
			d := db.Model(&TestUserModel{})
			return d, nil
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageListWithCount Err,%v", err.Error())
	}

	if cnt != 8 {
		t.Errorf("Expected count 8, got %d", cnt)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}

	fmt.Printf("Users: %+v, Count: %d\n", users, cnt)
}

func TestFindPageListWithMock(t *testing.T) {
	db, err := createMockDB()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	mockConn := NewMockGormcCacheConn(db)
	orderKeys := map[string]string{
		"age":  "age",
		"id":   "id",
		"name": "name",
	}

	users, cnt, err := FindPageList[TestUserModel](
		context.Background(),
		mockConn,
		&ListReq{Page: 1, PageSize: 3},
		OrderBy{OrderKey: "age", Sort: Asc()},
		orderKeys,
		func(conn *gorm.DB) (*gorm.DB, *gorm.DB) {
			d := conn.Model(&TestUserModel{})
			return d, nil
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageListWithMock Err: %v", err)
	}

	if cnt != 8 {
		t.Errorf("Expected count 8, got %d", cnt)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// 验证按年龄升序排序
	if len(users) > 1 && users[0].Age > users[1].Age {
		t.Errorf("Users not sorted by age ascending")
	}

	fmt.Printf("Users: %+v, Count: %d\n", users, cnt)
}

func TestFindPageListMultiOrderByWithMock(t *testing.T) {
	db, err := createMockDB()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	mockConn := NewMockGormcCacheConn(db)
	orderKeys := map[string]string{
		"age":  "age",
		"id":   "id",
		"name": "name",
	}

	orderBys := []OrderBy{
		{OrderKey: "age", Sort: Desc()},
		{OrderKey: "id", Sort: Asc()},
	}

	users, cnt, err := FindPageListMultiOrderBy[TestUserModel](
		context.Background(),
		mockConn,
		&ListReq{Page: 1, PageSize: 5},
		orderBys,
		orderKeys,
		func(conn *gorm.DB) (*gorm.DB, *gorm.DB) {
			d := conn.Model(&TestUserModel{})
			return d, nil
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageListMultiOrderByWithMock Err: %v", err)
	}

	if cnt != 8 {
		t.Errorf("Expected count 8, got %d", cnt)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}

	fmt.Printf("Users: %+v, Count: %d\n", users, cnt)
}

func TestFindPageListWithCountMultiOrderByWithMock(t *testing.T) {
	db, err := createMockDB()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	orderKeys := map[string]string{
		"age":  "age",
		"id":   "id",
		"name": "name",
	}

	orderBys := []OrderBy{
		{OrderKey: "age", Sort: Desc()},
		{OrderKey: "id", Sort: Asc()},
	}

	users, cnt, err := FindPageListWithCountMultiOrderBy[TestUserModel](
		context.Background(),
		&ListReq{Page: 1, PageSize: 5},
		orderBys,
		orderKeys,
		func() (*gorm.DB, *gorm.DB) {
			d := db.Model(&TestUserModel{})
			return d, nil
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageListWithCountMultiOrderByWithMock Err: %v", err)
	}

	if cnt != 8 {
		t.Errorf("Expected count 8, got %d", cnt)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}

	fmt.Printf("Users: %+v, Count: %d\n", users, cnt)
}
