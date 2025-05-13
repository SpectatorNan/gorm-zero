package batchx

import (
	"context"
	"errors"
	"testing"

	"github.com/SpectatorNan/gorm-zero/gormc"
	"github.com/SpectatorNan/gorm-zero/gormx"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel is a test model for batch operations
type TestModel struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:255"`
}

// Mock model with cache
type MockCachedModel struct {
	db *gorm.DB
}

func (m *MockCachedModel) GetCacheKeys(data *TestModel) []string {
	return []string{
		"cache:testmodel:" + data.Name,
		"cache:testmodel:id:" + string(rune(data.ID)),
	}
}

func (m *MockCachedModel) ExecCtx(ctx context.Context, execCtx gormc.ExecCtxFn, keys ...string) error {
	return execCtx(m.db)
}

// Mock model without cache
type MockNoCacheModel struct {
	db *gorm.DB
}

func (m *MockNoCacheModel) ExecCtx(ctx context.Context, execCtx gormx.ExecCtxFn) error {
	return execCtx(m.db)
}

// Setup test database
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&TestModel{})

	return db
}

func TestBatchExecCtx(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	// Create test data
	testData := []TestModel{
		{Name: "test1"},
		{Name: "test2"},
		{Name: "test3"},
	}

	// Insert test data
	for _, d := range testData {
		db.Create(&d)
	}

	// Create model
	mockModel := &MockCachedModel{db: db}

	t.Run("Success without tx", func(t *testing.T) {
		err := BatchExecCtx[TestModel](ctx, mockModel, testData, func(db *gorm.DB) error {
			return db.Model(&TestModel{}).Where("name = ?", "test1").Update("name", "updated1").Error
		}, nil)

		assert.NoError(t, err)

		// Verify update
		var result TestModel
		db.Where("name = ?", "updated1").First(&result)
		assert.Equal(t, "updated1", result.Name)
	})

	t.Run("Success with tx", func(t *testing.T) {
		// Start transaction
		tx := db.Begin()

		err := BatchExecCtx[TestModel](ctx, mockModel, testData, func(db *gorm.DB) error {
			return db.Model(&TestModel{}).Where("name = ?", "test2").Update("name", "updated2").Error
		}, tx)

		assert.NoError(t, err)

		// Commit transaction
		tx.Commit()

		// Verify update
		var result TestModel
		db.Where("name = ?", "updated2").First(&result)
		assert.Equal(t, "updated2", result.Name)
	})

	t.Run("Rollback on error without tx", func(t *testing.T) {
		// Get current count
		var count int64
		db.Model(&TestModel{}).Count(&count)
		initialCount := count

		err := BatchExecCtx[TestModel](ctx, mockModel, testData, func(db *gorm.DB) error {
			// Do some operation
			db.Create(&TestModel{Name: "to-be-rolled-back"})
			// Return error to trigger rollback
			return errors.New("forced error")
		}, nil)

		assert.Error(t, err)

		// Verify rollback (count should be the same)
		db.Model(&TestModel{}).Count(&count)
		assert.Equal(t, initialCount, count)
	})
}

func TestBatchNoCacheExecCtx(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	// Create model
	mockModel := &MockNoCacheModel{db: db}

	t.Run("Success without tx", func(t *testing.T) {
		err := BatchNoCacheExecCtx[TestModel](ctx, mockModel, func(db *gorm.DB) error {
			return db.Create(&TestModel{Name: "nocache1"}).Error
		}, nil)

		assert.NoError(t, err)

		// Verify creation
		var result TestModel
		db.Where("name = ?", "nocache1").First(&result)
		assert.Equal(t, "nocache1", result.Name)
	})

	t.Run("Success with tx", func(t *testing.T) {
		// Start transaction
		tx := db.Begin()

		err := BatchNoCacheExecCtx[TestModel](ctx, mockModel, func(db *gorm.DB) error {
			return db.Create(&TestModel{Name: "nocache2"}).Error
		}, tx)

		assert.NoError(t, err)

		// Commit transaction
		tx.Commit()

		// Verify creation
		var result TestModel
		db.Where("name = ?", "nocache2").First(&result)
		assert.Equal(t, "nocache2", result.Name)
	})

	t.Run("Rollback on error without tx", func(t *testing.T) {
		// Get current count
		var count int64
		db.Model(&TestModel{}).Count(&count)
		initialCount := count

		err := BatchNoCacheExecCtx[TestModel](ctx, mockModel, func(db *gorm.DB) error {
			// Do some operation
			db.Create(&TestModel{Name: "nocache-rollback"})
			// Return error to trigger rollback
			return errors.New("forced error")
		}, nil)

		assert.Error(t, err)

		// Verify rollback (count should be the same)
		db.Model(&TestModel{}).Count(&count)
		assert.Equal(t, initialCount, count)
	})
}

func TestGetCacheKeysByMultiData(t *testing.T) {
	mockModel := &MockCachedModel{db: nil}
	testData := []TestModel{
		{ID: 1, Name: "test1"},
		{ID: 2, Name: "test2"},
		{ID: 3, Name: "test1"}, // Duplicate name to test uniqueness
	}

	keys := getCacheKeysByMultiData(mockModel, testData)

	// We should have unique keys
	assert.Equal(t, 6, len(keys))

	// Test with empty data
	emptyKeys := getCacheKeysByMultiData(mockModel, []TestModel{})
	assert.Equal(t, 0, len(emptyKeys))
}

func TestUniqueKeys(t *testing.T) {
	keys := []string{
		"key1",
		"key2",
		"key1", // Duplicate
		"key3",
		"key2", // Duplicate
	}

	uniqueKeys := uniqueKeys(keys)

	// We should have 3 unique keys
	assert.Equal(t, 3, len(uniqueKeys))

	// Check if all keys are present
	keyMap := make(map[string]bool)
	for _, k := range uniqueKeys {
		keyMap[k] = true
	}

	assert.True(t, keyMap["key1"])
	assert.True(t, keyMap["key2"])
	assert.True(t, keyMap["key3"])
}
