package conn

import (
	"context"
	"fmt"
	"gorm.io/gen/field"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestUser 测试用的用户模型
type TestUser struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:100;not null"`
	Age  int    `gorm:"not null"`
}

func (TestUser) TableName() string {
	return "test_users"
}

type TestUserColumns struct {
	ID   field.Int
	Name field.String
	Age  field.Int
}

func getTestUserColumns() TestUserColumns {
	return TestUserColumns{
		ID:   field.NewInt("test_users", "id"),
		Name: field.NewString("test_users", "name"),
		Age:  field.NewInt("test_users", "age"),
	}
}

// ConnRaceTestSuite 竞争条件测试套件
type ConnRaceTestSuite struct {
	suite.Suite
	db       *gorm.DB
	testConn Conn[TestUser]
}

// SetupSuite 在整个测试套件开始前执行一次
func (suite *ConnRaceTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 关闭日志避免干扰测试
	})
	suite.Require().NoError(err)

	// 配置连接池以确保连接复用
	sqlDB, err := db.DB()
	suite.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1) // 强制使用单个连接
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	// 自动迁移表结构
	err = db.AutoMigrate(&TestUser{})
	suite.Require().NoError(err)

	suite.db = db
	suite.testConn = NewConn[TestUser](db)
}

// TearDownSuite 在整个测试套件结束后执行一次
func (suite *ConnRaceTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, err := suite.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// SetupTest 在每个测试方法执行前执行
func (suite *ConnRaceTestSuite) SetupTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM test_users")
}

// TearDownTest 在每个测试方法执行后执行
func (suite *ConnRaceTestSuite) TearDownTest() {
	// 可以在这里添加清理逻辑
}

func (suite *ConnRaceTestSuite) TestMultipleQuery() {
	// 首先创建一些测试数据
	testUsers := []*TestUser{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 25},
		{Name: "David", Age: 35},
		{Name: "Eve", Age: 30},
	}
	dbconn := suite.testConn.Debug()

	// 插入测试数据
	for _, user := range testUsers {
		err := dbconn.Create(user)
		suite.Require().NoError(err)
	}

	// 获取列定义
	ucol := getTestUserColumns()
	queryById := func(conn Repository[TestUser], id int) (*TestUser, error) {
		return conn.Where(ucol.ID.Eq(id)).First()
	}
	queryByName := func(conn Repository[TestUser], name string) ([]*TestUser, error) {
		return conn.Where(ucol.Name.Eq(name)).Find()
	}
	queryByAge := func(conn Repository[TestUser], age int) ([]*TestUser, error) {
		return conn.Where(ucol.Age.Eq(age)).Find()
	}

	// 获取所有用户用于后续查询
	allUsers, err := suite.testConn.Find()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(allUsers)

	for i := 0; i < 10; i++ {
		var result interface{}
		var err error
		var queryDesc string

		ridx := i % 5 // 修改为5种查询类型
		switch ridx {
		case 0:
			// 按ID查询 - 使用实际存在的ID
			userIdx := i % len(allUsers)
			result, err = queryById(dbconn, int(allUsers[userIdx].ID))
			queryDesc = fmt.Sprintf("ById(%d)", allUsers[userIdx].ID)
		case 1:
			// 按名称查询 - 使用实际存在的名称
			names := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
			name := names[i%len(names)]
			result, err = queryByName(dbconn, name)
			queryDesc = fmt.Sprintf("ByName(%s)", name)
		case 2:
			// 按年龄查询 - 使用实际存在的年龄
			ages := []int{25, 30, 35}
			age := ages[i%len(ages)]
			result, err = queryByAge(dbconn, age)
			queryDesc = fmt.Sprintf("ByAge(%d)", age)
		case 3:
			// 查询所有
			result, err = dbconn.Find()
			queryDesc = "FindAll"
		case 4:
			// 带限制的查询
			limit := (i % 3) + 1
			result, err = dbconn.Limit(limit).Find()
			queryDesc = fmt.Sprintf("WithLimit(%d)", limit)
		}

		suite.T().Logf("Query %d: %s", i, queryDesc)

		// 验证查询结果
		suite.Require().NoError(err, "Query %d (%s) should not fail", i, queryDesc)

		switch r := result.(type) {
		case *TestUser:
			suite.NotNil(r, "Single user query should return a result")
			suite.T().Logf("  Found user: ID=%d, Name=%s, Age=%d", r.ID, r.Name, r.Age)
		case []*TestUser:
			suite.T().Logf("  Found %d users", len(r))
			if len(r) > 0 {
				suite.T().Logf("  First user: ID=%d, Name=%s, Age=%d", r[0].ID, r[0].Name, r[0].Age)
			}
		}
	}

	// 测试一些边界情况
	suite.T().Log("=== 测试边界情况 ===")

	// 查询不存在的ID
	_, err = queryById(dbconn, 999)
	suite.Error(err, "Should get error when querying non-existent ID")
	suite.T().Logf("Expected error for non-existent ID: %v", err)

	// 查询不存在的名称
	result, err := queryByName(dbconn, "NonExistentUser")
	suite.NoError(err, "Should not error when querying non-existent name")
	suite.Empty(result, "Should return empty slice for non-existent name")

	// 查询不存在的年龄
	result, err = queryByAge(dbconn, 999)
	suite.NoError(err, "Should not error when querying non-existent age")
	suite.Empty(result, "Should return empty slice for non-existent age")

	// 验证数据完整性
	finalUsers, err := dbconn.Find()
	suite.Require().NoError(err)
	suite.Equal(len(testUsers), len(finalUsers), "All test users should still exist")

	suite.T().Log("=== 多查询操作测试完成 ===")
}

// TestConnRaceCondition 测试conn在并发环境下是否存在竞争条件
func (suite *ConnRaceTestSuite) TestConnRaceCondition() {
	const (
		numGoroutines          = 10 // 并发goroutine数量
		operationsPerGoroutine = 3  // 每个goroutine执行的操作次数
	)

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*operationsPerGoroutine)

	// 启动多个goroutine，每个使用独立的conn实例
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// 为每个goroutine创建独立的conn实例
			userConn := NewConn[TestUser](suite.db)

			for j := 0; j < operationsPerGoroutine; j++ {
				// 创建用户
				user := &TestUser{
					Name: fmt.Sprintf("User_%d_%d", goroutineID, j),
					Age:  20 + (goroutineID+j)%50,
				}

				// 测试创建操作
				if err := userConn.Create(user); err != nil {
					errorChan <- fmt.Errorf("goroutine %d operation %d create failed: %v", goroutineID, j, err)
					return
				}

				// 测试查询操作
				foundUsers, err := userConn.Find()
				if err != nil {
					errorChan <- fmt.Errorf("goroutine %d operation %d find failed: %v", goroutineID, j, err)
					return
				}

				// 验证查询结果
				if len(foundUsers) == 0 {
					errorChan <- fmt.Errorf("goroutine %d operation %d: expected to find users but got empty result", goroutineID, j)
					return
				}

				// 测试更新操作
				if len(foundUsers) > 0 {
					updateUser := foundUsers[0]
					updateUser.Age = updateUser.Age + 1
					if err := userConn.Save(updateUser); err != nil {
						errorChan <- fmt.Errorf("goroutine %d operation %d update failed: %v", goroutineID, j, err)
						return
					}
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误发生
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// 如果有错误，打印前几个错误信息帮助调试
	if len(errors) > 0 {
		suite.T().Logf("Found %d errors during concurrent operations:", len(errors))
		for i, err := range errors {
			if i >= 5 { // 只打印前5个错误
				suite.T().Logf("... and %d more errors", len(errors)-5)
				break
			}
			suite.T().Logf("Error %d: %v", i+1, err)
		}
	}

	// 断言没有发生竞争条件导致的错误
	suite.Empty(errors, "Concurrent operations should not produce race condition errors")

	// 验证最终数据状态
	finalConn := NewConn[TestUser](suite.db)
	allUsers, err := finalConn.Find()
	suite.Require().NoError(err)

	expectedUserCount := numGoroutines * operationsPerGoroutine
	suite.Equal(expectedUserCount, len(allUsers), "Should have created expected number of users")

	suite.T().Logf("Successfully completed %d concurrent operations with %d goroutines",
		numGoroutines*operationsPerGoroutine, numGoroutines)
	suite.T().Logf("Final user count: %d", len(allUsers))
}

// TestConnTransactionRaceCondition 测试conn事务操作的竞争条件
func (suite *ConnRaceTestSuite) TestConnTransactionRaceCondition() {
	const numGoroutines = 20
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 并发执行事务操作
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			err := suite.testConn.Transaction(func(tx *ConnTx) error {
				// 从事务创建新的conn实例来执行操作
				txConn := NewConnFromTx[TestUser](tx)

				// 在事务中创建用户
				user1 := &TestUser{
					Name: fmt.Sprintf("TxUser1_%d", goroutineID),
					Age:  25,
				}

				user2 := &TestUser{
					Name: fmt.Sprintf("TxUser2_%d", goroutineID),
					Age:  30,
				}

				// 创建第一个用户
				if err := txConn.Create(user1); err != nil {
					return fmt.Errorf("failed to create user1: %v", err)
				}

				// 模拟一些处理时间
				time.Sleep(time.Millisecond * 10)

				// 创建第二个用户
				if err := txConn.Create(user2); err != nil {
					return fmt.Errorf("failed to create user2: %v", err)
				}

				return nil
			})

			if err != nil {
				errorChan <- fmt.Errorf("goroutine %d transaction failed: %v", goroutineID, err)
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// 检查事务错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	suite.Empty(errors, "Concurrent transactions should not produce race condition errors")

	// 验证事务结果
	allUsers, err := suite.testConn.Find()
	suite.Require().NoError(err)

	expectedUserCount := numGoroutines * 2 // 每个事务创建2个用户
	suite.Equal(expectedUserCount, len(allUsers), "All transaction users should be created")

	suite.T().Logf("Successfully completed %d concurrent transactions", numGoroutines)
	suite.T().Logf("Final user count from transactions: %d", len(allUsers))
}

// TestConnRepositoryClone 测试conn仓库克隆的竞争条件
func (suite *ConnRaceTestSuite) TestConnRepositoryClone() {
	const numGoroutines = 30
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// 每个goroutine使用WithContext创建新的连接实例
			ctx := context.Background()
			contextConn := suite.testConn.WithContext(ctx)

			// 创建用户
			user := &TestUser{
				Name: fmt.Sprintf("CloneUser_%d", goroutineID),
				Age:  goroutineID + 18,
			}

			if err := contextConn.Create(user); err != nil {
				errorChan <- fmt.Errorf("goroutine %d clone create failed: %v", goroutineID, err)
				return
			}

			// 查询用户 - 使用原生SQL查询
			var users []*TestUser
			err := contextConn.UnderlyingDB().Where("name LIKE ?", fmt.Sprintf("CloneUser_%d", goroutineID)).Find(&users).Error
			if err != nil {
				errorChan <- fmt.Errorf("goroutine %d clone find failed: %v", goroutineID, err)
				return
			}

			if len(users) == 0 {
				errorChan <- fmt.Errorf("goroutine %d: could not find created user", goroutineID)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// 检查错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	suite.Empty(errors, "Concurrent repository clone operations should not produce race condition errors")

	// 验证最终结果
	allUsers, err := suite.testConn.Find()
	suite.Require().NoError(err)
	suite.Equal(numGoroutines, len(allUsers), "Should have created one user per goroutine")

	suite.T().Logf("Successfully completed %d concurrent clone operations", numGoroutines)
}

// TestConnSharedInstanceRaceCondition 专门测试共享conn实例的竞争条件
func (suite *ConnRaceTestSuite) TestConnSharedInstanceRaceCondition() {
	// 创建一个共享的conn实例
	sharedConn := NewConn[TestUser](suite.db)

	const numGoroutines = 10
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 多个goroutine同时使用同一个conn实例
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// 创建用户
			user := &TestUser{
				Name: fmt.Sprintf("SharedUser_%d", goroutineID),
				Age:  20 + goroutineID,
			}

			// 使用共享的conn实例进行操作
			if err := sharedConn.Create(user); err != nil {
				errorChan <- fmt.Errorf("goroutine %d shared create failed: %v", goroutineID, err)
				return
			}

			// 查询操作
			users, err := sharedConn.Find()
			if err != nil {
				errorChan <- fmt.Errorf("goroutine %d shared find failed: %v", goroutineID, err)
				return
			}

			if len(users) == 0 {
				errorChan <- fmt.Errorf("goroutine %d: no users found in shared test", goroutineID)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// 检查错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		suite.T().Logf("Found %d errors in shared conn test:", len(errors))
		for i, err := range errors {
			suite.T().Logf("Error %d: %v", i+1, err)
		}
	}

	// 这里我们预期可能会有竞争条件错误
	// 所以我们记录错误但不让测试失败，以便观察竞争条件的发生
	suite.T().Logf("Shared conn test completed with %d errors (race conditions expected)", len(errors))

	// 验证最终数据状态
	allUsers, err := sharedConn.Find()
	suite.Require().NoError(err)
	suite.T().Logf("Final user count in shared test: %d", len(allUsers))
}

// TestConnAggressiveRaceCondition 更激进的竞争条件测试
func (suite *ConnRaceTestSuite) TestConnAggressiveRaceCondition() {
	// 创建一个共享的conn实例
	sharedConn := NewConn[TestUser](suite.db)

	const numGoroutines = 50
	const operationsPerGoroutine = 10
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*operationsPerGoroutine)

	// 多个goroutine同时使用同一个conn实例进行更复杂的操作
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// 创建用户
				user := &TestUser{
					Name: fmt.Sprintf("AggressiveUser_%d_%d", goroutineID, j),
					Age:  20 + (goroutineID+j)%50,
				}

				// 使用共享的conn实例进行操作，同时进行WithContext调用
				ctx := context.Background()
				contextConn := sharedConn.WithContext(ctx)

				if err := contextConn.Create(user); err != nil {
					errorChan <- fmt.Errorf("goroutine %d operation %d aggressive create failed: %v", goroutineID, j, err)
					continue
				}

				// 查询操作
				users, err := sharedConn.Find()
				if err != nil {
					errorChan <- fmt.Errorf("goroutine %d operation %d aggressive find failed: %v", goroutineID, j, err)
					continue
				}

				// 更新操作
				if len(users) > 0 {
					updateUser := users[0]
					updateUser.Age = updateUser.Age + 1
					if err := sharedConn.Save(updateUser); err != nil {
						errorChan <- fmt.Errorf("goroutine %d operation %d aggressive update failed: %v", goroutineID, j, err)
						continue
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// 检查错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		suite.T().Logf("Found %d errors in aggressive shared conn test:", len(errors))
		for i, err := range errors {
			if i >= 10 { // 只打印前10个错误
				suite.T().Logf("... and %d more errors", len(errors)-10)
				break
			}
			suite.T().Logf("Error %d: %v", i+1, err)
		}
	}

	// 这里我们预期在高并发下可能会有竞争条件错误
	suite.T().Logf("Aggressive shared conn test completed with %d errors (race conditions expected)", len(errors))

	// 验证最终数据状态
	allUsers, err := sharedConn.Find()
	suite.Require().NoError(err)
	suite.T().Logf("Final user count in aggressive test: %d", len(allUsers))
}

// TestConnRaceConditionDemo 专门用于演示竞争条件问题的测试
func (suite *ConnRaceTestSuite) TestConnRaceConditionDemo() {
	suite.T().Log("=== 演示：使用共享conn实例的竞争条件问题 ===")

	// 使用套件的共享conn实例，这会导致竞争条件
	sharedConn := suite.testConn

	const numGoroutines = 20
	const operationsPerGoroutine = 5
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*operationsPerGoroutine)

	suite.T().Logf("启动 %d 个goroutine，每个执行 %d 次操作，使用共享的conn实例",
		numGoroutines, operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// 同时使用WithContext和直接操作
				ctx := context.Background()
				contextConn := sharedConn.WithContext(ctx)

				// 创建用户
				user := &TestUser{
					Name: fmt.Sprintf("RaceDemo_%d_%d", goroutineID, j),
					Age:  20 + (goroutineID+j)%50,
				}

				// 这里会发生竞争条件：多个goroutine同时修改sharedConn的内部状态
				if err := contextConn.Create(user); err != nil {
					errorChan <- fmt.Errorf("goroutine %d operation %d failed: %v", goroutineID, j, err)
					continue
				}

				// 查询操作 - 也会发生竞争条件
				_, err := sharedConn.Find()
				if err != nil {
					errorChan <- fmt.Errorf("goroutine %d operation %d find failed: %v", goroutineID, j, err)
					continue
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// 检查错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	suite.T().Logf("=== 测试结果 ===")
	suite.T().Logf("检测到 %d 个错误（使用 -race 标志运行会显示数据竞争警告）", len(errors))

	if len(errors) > 0 {
		suite.T().Logf("前几个错误示例：")
		for i, err := range errors {
			if i >= 3 {
				suite.T().Logf("... 还有 %d 个错误", len(errors)-3)
				break
			}
			suite.T().Logf("  %d. %v", i+1, err)
		}
	}

	// 验证最终数据状态
	allUsers, err := sharedConn.Find()
	suite.Require().NoError(err)

	expectedCount := numGoroutines * operationsPerGoroutine
	actualCount := len(allUsers)

	suite.T().Logf("期望用户数: %d, 实际用户数: %d", expectedCount, actualCount)

	// 在竞争条件，实际数量可能与期望不符，但这里我们不让测试失败
	// 因为这是用于演示竞争条件问题的测试
	if actualCount != expectedCount {
		suite.T().Logf("⚠️  数据不一致！这表明存在竞争条件问题")
	}

	suite.T().Log("=== 建议 ===")
	suite.T().Log("- 在生产环境中，每个goroutine应该使用独立的conn实例")
	suite.T().Log("- 或者使用NewConn()为每个并发操作创建新的连接")
	suite.T().Log("- 使用 'go test -race' 命令可以检测到数据竞争")
}

// TestConnRaceTestSuite 运行测试套件的入口函数
func TestConnRaceTestSuite(t *testing.T) {
	suite.Run(t, new(ConnRaceTestSuite))
}
