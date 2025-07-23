package example

import (
	"context"
	"fmt"
	mysql2 "github.com/SpectatorNan/gorm-zero/v2/config/mysql"
	"github.com/SpectatorNan/gorm-zero/v2/pagex"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func getGormDb() (*gorm.DB, cache.CacheConf) {

	cfg := mysql2.Mysql{
		Path:     "localhost",
		Port:     3306,
		Config:   "charset=utf8mb4&parseTime=true&loc=Local",
		Dbname:   "gormzero",
		Username: "root",
		Password: "wXrtVWDz374vXKmJ",
		//LogMode:  "info",
		LogMode: "silent",
	}

	db, err := mysql2.Connect(cfg)
	if err != nil {
		panic(err)
	}
	ccf := cache.CacheConf{
		cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: "127.0.0.1:6379",
				Pass: "yourpassword",
				Type: "node",
			},
			Weight: 100,
		},
	}
	return db, ccf
}

// UserModelTestSuite 使用 testify suite 组织测试
type UserModelTestSuite struct {
	suite.Suite
	model *defaultUsersModel
	ctx   context.Context
	db    *gorm.DB
}

// SetupSuite 在整个测试套件开始前执行
func (suite *UserModelTestSuite) SetupSuite() {
	logx.Disable()
	stat.SetReporter(nil)

	db, cc := getGormDb()
	db.Logger.LogMode(logger.Silent)
	suite.db = db
	suite.model = newUsersModel(db, cc)
	suite.ctx = context.Background()
}

// TearDownSuite 在整个测试套件结束后执行
func (suite *UserModelTestSuite) TearDownSuite() {
	// 清理测试数据
	if suite.db != nil {
		// 删除测试期间创建的数据
		suite.db.Exec("DELETE FROM users WHERE account LIKE 'test_%' OR account LIKE 'batch_%' OR account LIKE 'cache_%' OR account LIKE 'updated_%' OR account LIKE 'concurrent_%'")
	}
	//time.Sleep(time.Second)
}

// SetupTest 在每个测试方法执行前执行
func (suite *UserModelTestSuite) SetupTest() {
	// 每个测试前的设置
}

// TearDownTest 在每个测试方法执行后执行
func (suite *UserModelTestSuite) TearDownTest() {
	// 每个测试后的清理
}

// TestUserInsert 测试用户插入
func (suite *UserModelTestSuite) TestUserInsert() {
	user := &Users{
		Account:  "test_suite_user",
		NickName: "Test Suite User",
		Password: "test_password",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Insert should succeed")
	suite.Assert().NotZero(user.Id, "User ID should be generated")
	suite.Assert().Equal("test_suite_user", user.Account, "Account should match")
}

// TestUserFindOne 测试查找单个用户
func (suite *UserModelTestSuite) TestUserFindOne() {
	// 先插入测试数据
	user := &Users{
		Account:  "test_find_user",
		NickName: "Test Find User",
		Password: "find_password",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Setup insert should succeed")

	// 测试查找
	foundUser, err := suite.model.FindOne(suite.ctx, user.Id)
	suite.Require().NoError(err, "FindOne should succeed")
	suite.Assert().NotNil(foundUser, "Found user should not be nil")
	suite.Assert().Equal(user.Id, foundUser.Id, "User ID should match")
	suite.Assert().Equal(user.Account, foundUser.Account, "Account should match")
}

// TestUserUpdate 测试用户更新
func (suite *UserModelTestSuite) TestUserUpdate() {
	// 先插入测试数据
	user := &Users{
		Account:  "test_update_user",
		NickName: "Test Update User",
		Password: "update_password",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Setup insert should succeed")

	// 更新用户信息
	user.NickName = "Updated User"
	user.Password = "updated_password"

	err = suite.model.Update(suite.ctx, nil, user)
	suite.Require().NoError(err, "Update should succeed")

	// 验证更新结果
	foundUser, err := suite.model.FindOne(suite.ctx, user.Id)
	suite.Require().NoError(err, "FindOne after update should succeed")
	suite.Assert().Equal("Updated User", foundUser.NickName, "NickName should be updated")
	suite.Assert().Equal("updated_password", foundUser.Password, "Password should be updated")
}

// TestUserDelete 测试用户删除
func (suite *UserModelTestSuite) TestUserDelete() {
	// 先插入测试数据
	user := &Users{
		Account:  "test_delete_user",
		NickName: "Test Delete User",
		Password: "delete_password",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Setup insert should succeed")

	// 删除用户
	err = suite.model.Delete(suite.ctx, nil, user.Id)
	suite.Require().NoError(err, "Delete should succeed")

	// 验证删除结果 - 软删除，用户应该找不到
	_, err = suite.model.FindOne(suite.ctx, user.Id)
	suite.Assert().Error(err, "FindOne after delete should return error")
}

// TestBatchOperations 测试批量操作
func (suite *UserModelTestSuite) TestBatchOperations() {
	// 批量插入测试数据
	users := []*Users{
		{Account: "test_batch_1", NickName: "Test Batch 1", Password: "password1"},
		{Account: "test_batch_2", NickName: "Test Batch 2", Password: "password2"},
		{Account: "test_batch_3", NickName: "Test Batch 3", Password: "password3"},
	}

	// 测试批量插入
	err := suite.model.BatchInsert(suite.ctx, nil, users)
	suite.Require().NoError(err, "BatchInsert should succeed")
	suite.Assert().Len(users, 3, "Should have 3 users")

	for i, user := range users {
		suite.Assert().NotZero(user.Id, "User %d should have generated ID", i+1)
	}

	// 测试批量更新
	for _, user := range users {
		user.NickName = "Updated " + user.NickName
	}

	err = suite.model.BatchUpdate(suite.ctx, nil, users)
	suite.Require().NoError(err, "BatchUpdate should succeed")

	// 验证更新结果
	for i, user := range users {
		foundUser, err := suite.model.FindOne(suite.ctx, user.Id)
		suite.Require().NoError(err, "Should find updated user %d", i+1)
		suite.Assert().Contains(foundUser.NickName, "Updated", "User %d nickname should be updated", i+1)
	}

	// 测试批量删除
	err = suite.model.BatchDelete(suite.ctx, nil, users)
	suite.Require().NoError(err, "BatchDelete should succeed")

	// 验证删除结果
	for i, user := range users {
		_, err := suite.model.FindOne(suite.ctx, user.Id)
		suite.Assert().Error(err, "User %d should not be found after deletion", i+1)
	}
}

// TestTransactionOperations 测试事务操作
func (suite *UserModelTestSuite) TestTransactionOperations() {
	// 开始事务
	tx := suite.db.Begin()
	suite.Require().NoError(tx.Error, "Transaction should begin successfully")

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			suite.Fail("Transaction panicked", "panic: %v", r)
		}
	}()

	// 在事务中插入用户
	user := &Users{
		Account:  "test_tx_user",
		NickName: "Test Transaction User",
		Password: "tx_password",
	}

	err := suite.model.Insert(suite.ctx, tx, user)
	suite.Require().NoError(err, "Transaction insert should succeed")
	suite.Assert().NotZero(user.Id, "User should have generated ID")

	// 提交事务
	err = tx.Commit().Error
	suite.Require().NoError(err, "Transaction commit should succeed")

	// 验证事务提交后数据存在
	foundUser, err := suite.model.FindOne(suite.ctx, user.Id)
	suite.Require().NoError(err, "Should find committed user")
	suite.Assert().Equal(user.Account, foundUser.Account, "Committed user data should match")
}

// TestTransactionRollback 测试事务回滚
func (suite *UserModelTestSuite) TestTransactionRollback() {
	// 开始事务
	tx := suite.db.Begin()
	suite.Require().NoError(tx.Error, "Transaction should begin successfully")

	// 在事务中插入用户
	user := &Users{
		Account:  "test_rollback_user",
		NickName: "Test Rollback User",
		Password: "rollback_password",
	}

	err := suite.model.Insert(suite.ctx, tx, user)
	suite.Require().NoError(err, "Insert in transaction should succeed")
	suite.Assert().NotZero(user.Id, "User should have generated ID")

	userId := user.Id

	// 回滚事务
	err = tx.Rollback().Error
	suite.Require().NoError(err, "Transaction rollback should succeed")

	// 验证回滚后数据不存在
	_, err = suite.model.FindOne(suite.ctx, userId)
	suite.Assert().Error(err, "User should not exist after transaction rollback")
}

// TestCacheConsistency 测试缓存一致性
func (suite *UserModelTestSuite) TestCacheConsistency() {
	// 插入测试用户
	user := &Users{
		Account:  "test_cache_user",
		NickName: "Test Cache User",
		Password: "cache_password",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Insert should succeed")

	userId := user.Id

	// 第一次查询 - 建立缓存
	user1, err := suite.model.FindOne(suite.ctx, userId)
	suite.Require().NoError(err, "First query should succeed")

	// 第二次查询 - 从缓存读取
	user2, err := suite.model.FindOne(suite.ctx, userId)
	suite.Require().NoError(err, "Second query should succeed")

	// 验证数据一致性
	suite.Assert().Equal(user1.Id, user2.Id, "Cache consistency check - ID")
	suite.Assert().Equal(user1.Account, user2.Account, "Cache consistency check - Account")
	suite.Assert().Equal(user1.NickName, user2.NickName, "Cache consistency check - NickName")

	// 更新操作 - 测试缓存失效
	user1.Account = "test_updated_cache_user"
	err = suite.model.Update(suite.ctx, nil, user1)
	suite.Require().NoError(err, "Update should succeed")

	// 查询更新后的数据
	user3, err := suite.model.FindOne(suite.ctx, userId)
	suite.Require().NoError(err, "Query after update should succeed")
	suite.Assert().Equal("test_updated_cache_user", user3.Account, "Updated data should be reflected")
}

// TestErrorHandling 测试错误处理
func (suite *UserModelTestSuite) TestErrorHandling() {
	// 测试查找不存在的用户
	_, err := suite.model.FindOne(suite.ctx, 999999)
	suite.Assert().Error(err, "Should return error for non-existent user")
}

// TestConcurrentOperations 测试并发操作
func (suite *UserModelTestSuite) TestConcurrentOperations() {
	const goroutineCount = 3
	const usersPerGoroutine = 2

	done := make(chan error, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(routineIndex int) {
			users := make([]*Users, usersPerGoroutine)
			for j := 0; j < usersPerGoroutine; j++ {
				users[j] = &Users{
					Account:  fmt.Sprintf("test_concurrent_%d_%d", routineIndex, j),
					NickName: fmt.Sprintf("Test Concurrent User %d-%d", routineIndex, j),
					Password: fmt.Sprintf("concurrent_pass_%d_%d", routineIndex, j),
				}
			}

			err := suite.model.BatchInsert(suite.ctx, nil, users)
			done <- err
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < goroutineCount; i++ {
		err := <-done
		suite.Assert().NoError(err, "Concurrent operation %d should succeed", i)
	}
}

// TestPageListWithoutOrderKeys 测试不使用 orderkey 映射的分页查询
func (suite *UserModelTestSuite) TestPageListWithoutOrderKeys() {
	// 准备测试数据
	testUsers := []*Users{
		{Account: "page_user_001", NickName: "Page User 1", Password: "pass1"},
		{Account: "page_user_002", NickName: "Page User 2", Password: "pass2"},
		{Account: "page_user_003", NickName: "Page User 3", Password: "pass3"},
		{Account: "page_user_004", NickName: "Page User 4", Password: "pass4"},
		{Account: "page_user_005", NickName: "Page User 5", Password: "pass5"},
	}

	// 批量插入测试数据
	for _, user := range testUsers {
		err := suite.model.Insert(suite.ctx, nil, user)
		suite.Require().NoError(err, "Insert test data should succeed")
	}

	// 测试分页查询 - 第1页，每页2条
	page := &pagex.PagePrams{
		Page:     1,
		PageSize: 2,
	}
	orderBy := pagex.OrderParams{
		OrderKey: "account",
		Sort:     pagex.Asc(),
	}

	users, total, err := suite.model.FindPageList(suite.ctx, page, orderBy, nil)
	suite.Require().NoError(err, "FindPageList should succeed")
	suite.Assert().GreaterOrEqual(total, int64(5), "Total should be at least 5")
	suite.Assert().LessOrEqual(len(users), 2, "Should return at most 2 users per page")

	// 验证排序结果 - 按 account 升序
	if len(users) >= 2 {
		suite.Assert().True(users[0].Account <= users[1].Account, "Results should be sorted by account ASC")
	}
}

// TestPageListWithOrderKeys 测试使用 orderkey 映射的分页查询
func (suite *UserModelTestSuite) TestPageListWithOrderKeys() {
	// 准备测试数据
	testUsers := []*Users{
		{Account: "map_user_001", NickName: "Map User A", Password: "pass1"},
		{Account: "map_user_002", NickName: "Map User B", Password: "pass2"},
		{Account: "map_user_003", NickName: "Map User C", Password: "pass3"},
	}

	// 批量插入测试数据
	for _, user := range testUsers {
		err := suite.model.Insert(suite.ctx, nil, user)
		suite.Require().NoError(err, "Insert test data should succeed")
	}

	// 设置 orderkey 映射 - 将前端字段名映射到数据库字段名
	orderKeys := map[string]string{
		"username":   "account",    // 前端用 username，映射到数据库的 account 字段
		"nickname":   "nick_name",  // 前端用 nickname，映射到数据库的 nick_name 字段
		"deleteTime": "deleted_at", // 前端用 deleteTime，映射到数据库的 deleted_at 字段
	}

	// 测试使用映射的排序键 - 前端传 username，实际按 account 字段排序
	page := &pagex.PagePrams{
		Page:     1,
		PageSize: 10,
	}
	orderBy := pagex.OrderParams{
		OrderKey: "username", // 前端字段名
		Sort:     pagex.Desc(),
	}

	users, total, err := suite.model.FindPageList(suite.ctx, page, orderBy, orderKeys)
	suite.Require().NoError(err, "FindPageList with orderKeys should succeed")
	suite.Assert().GreaterOrEqual(total, int64(3), "Total should be at least 3")

	// 验证排序结果 - 按 account 降序（通过 username 映射）
	if len(users) >= 2 {
		suite.Assert().True(users[0].Account >= users[1].Account, "Results should be sorted by account DESC via username mapping")
	}
}

// TestPageListWithNickNameMapping 测试昵称字段的 orderkey 映射
func (suite *UserModelTestSuite) TestPageListWithNickNameMapping() {
	// 准备测试数据
	testUsers := []*Users{
		{Account: "nick_user_001", NickName: "Alice", Password: "pass1"},
		{Account: "nick_user_002", NickName: "Bob", Password: "pass2"},
		{Account: "nick_user_003", NickName: "Charlie", Password: "pass3"},
	}

	// 批量插入测试数据
	for _, user := range testUsers {
		err := suite.model.Insert(suite.ctx, nil, user)
		suite.Require().NoError(err, "Insert test data should succeed")
	}

	// 设置 orderkey 映射
	orderKeys := map[string]string{
		"displayName": "nick_name", // 前端用 displayName，映射到数据库的 nick_name 字段
	}

	// 测试使用映射的排序键 - 前端传 displayName，实际按 nick_name 字段排序
	page := &pagex.PagePrams{
		Page:     1,
		PageSize: 10,
	}
	orderBy := pagex.OrderParams{
		OrderKey: "displayName", // 前端字段名
		Sort:     pagex.Asc(),
	}

	users, total, err := suite.model.FindPageList(suite.ctx, page, orderBy, orderKeys)
	suite.Require().NoError(err, "FindPageList with nickname mapping should succeed")
	suite.Assert().GreaterOrEqual(total, int64(3), "Total should be at least 3")

	// 验证排序结果 - 按 nick_name 升序（通过 displayName 映射）
	if len(users) >= 2 {
		suite.Assert().True(users[0].NickName <= users[1].NickName, "Results should be sorted by nick_name ASC via displayName mapping")
	}
}

// TestPageListWithInvalidOrderKey 测试无效的 orderkey 映射
func (suite *UserModelTestSuite) TestPageListWithInvalidOrderKey() {
	// 准备测试数据
	user := &Users{
		Account:  "invalid_order_user",
		NickName: "Invalid Order User",
		Password: "pass",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Insert test data should succeed")

	// 设置包含无效映射的 orderKeys
	orderKeys := map[string]string{
		"invalidField": "non_existent_field", // 映射到不存在的字段
	}

	// 测试使用无效的排序键 - 前端传 invalidField，实际按 non_existent_field 排序
	page := &pagex.PagePrams{
		Page:     1,
		PageSize: 10,
	}
	orderBy := pagex.OrderParams{
		OrderKey: "invalidField", // 无效的字段名
		Sort:     pagex.Asc(),
	}

	users, total, err := suite.model.FindPageList(suite.ctx, page, orderBy, orderKeys)
	// 即使 orderkey 无效，查询也应该成功，只是不会应用排序
	suite.Require().NoError(err, "FindPageList with invalid orderKey should still succeed")
	suite.Assert().GreaterOrEqual(total, int64(1), "Total should be at least 1")
	suite.Assert().GreaterOrEqual(len(users), 1, "Should return at least 1 user")
}

// TestPageListWithEmptyOrderKeys 测试空的 orderkey 映射
func (suite *UserModelTestSuite) TestPageListWithEmptyOrderKeys() {
	// 准备测试数据
	user := &Users{
		Account:  "empty_order_user",
		NickName: "Empty Order User",
		Password: "pass",
	}

	err := suite.model.Insert(suite.ctx, nil, user)
	suite.Require().NoError(err, "Insert test data should succeed")

	// 测试使用空的 orderKeys 映射
	page := &pagex.PagePrams{
		Page:     1,
		PageSize: 10,
	}
	orderBy := pagex.OrderParams{
		OrderKey: "account", // 直接使用数据库字段名
		Sort:     pagex.Asc(),
	}

	users, total, err := suite.model.FindPageList(suite.ctx, page, orderBy, map[string]string{})
	suite.Require().NoError(err, "FindPageList with empty orderKeys should succeed")
	suite.Assert().GreaterOrEqual(total, int64(1), "Total should be at least 1")

	// 再测试传入 nil 的情况
	users2, total2, err2 := suite.model.FindPageList(suite.ctx, page, orderBy, nil)
	suite.Require().NoError(err2, "FindPageList with nil orderKeys should succeed")
	suite.Assert().GreaterOrEqual(total2, int64(1), "Total should be at least 1")
	suite.Assert().Equal(len(users), len(users2), "Results should be the same for empty map and nil")
}

// 运行测试套件
func TestUserModelSuite(t *testing.T) {
	suite.Run(t, new(UserModelTestSuite))
}
