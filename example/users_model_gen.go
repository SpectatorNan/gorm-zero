package example

import (
	"context"
	"errors"
	"fmt"
	"github.com/SpectatorNan/gorm-zero/v2/conn"
	"github.com/SpectatorNan/gorm-zero/v2/connx"
	"github.com/SpectatorNan/gorm-zero/v2/executor"
	"github.com/SpectatorNan/gorm-zero/v2/helper"
	"github.com/SpectatorNan/gorm-zero/v2/pagex"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/gorm"
)

var (
	cacheGormzeroUsersIdPrefix = "cache:gormzero:users:id:"
)

// 缓存keys管理器
type (
	usersModel interface {
		Insert(ctx context.Context, tx *conn.ConnTx, data *Users) error
		BatchInsert(ctx context.Context, tx *conn.ConnTx, news []*Users) error
		FindOne(ctx context.Context, id int64) (*Users, error)
		FindPageList(ctx context.Context, page *pagex.PagePrams, orderBy pagex.OrderParams, orderKeys map[string]string, conds ...gen.Condition) ([]*Users, int64, error)
		Update(ctx context.Context, tx *conn.ConnTx, data *Users) error
		BatchUpdate(ctx context.Context, tx *conn.ConnTx, news []*Users) error
		BatchDelete(ctx context.Context, tx *conn.ConnTx, datas []*Users) error
		Delete(ctx context.Context, tx *conn.ConnTx, id int64) error
	}

	defaultUsersModel struct {
		usersDo      connx.CachedConn[Users]
		pageExecutor *executor.PageExecutor[Users]

		ALL       field.Asterisk
		Id        field.Int64
		Account   field.String
		NickName  field.String
		Password  field.String
		CreatedAt field.Time
		UpdatedAt field.Time
		DeletedAt field.Time

		fieldMap map[string]field.Expr
	}

	Users struct {
		Id        int64          `gorm:"column:id;primary_key;autoIncrement:true"`
		Account   string         `gorm:"column:account"`
		NickName  string         `gorm:"column:nick_name"`
		Password  string         `gorm:"column:password"`
		CreatedAt time.Time      `gorm:"column:created_at"`
		UpdatedAt time.Time      `gorm:"column:updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	}
)

const TableNameUsers = "`users`"

func (Users) TableName() string {
	return TableNameUsers
}

func newUsersModel(db *gorm.DB, c cache.CacheConf) *defaultUsersModel {
	return newUsersWithOption(db, c, connx.Option{})
}

func newUsersWithOption(db *gorm.DB, c cache.CacheConf, opt connx.Option) *defaultUsersModel {

	_users := &defaultUsersModel{}

	// 创建 users 实例，它将作为缓存键提供者
	do := connx.NewCachedConn[Users](db, c, _users, opt)
	_users.usersDo = do

	tableName := _users.usersDo.Conn.TableName()
	_users.ALL = field.NewAsterisk(tableName)
	_users.Id = field.NewInt64(tableName, "id")
	_users.Account = field.NewString(tableName, "account")
	_users.NickName = field.NewString(tableName, "nick_name")
	_users.Password = field.NewString(tableName, "password")
	_users.CreatedAt = field.NewTime(tableName, "created_at")
	_users.UpdatedAt = field.NewTime(tableName, "updated_at")
	_users.DeletedAt = field.NewTime(tableName, "deleted_at")
	_users.fillFieldMap()

	return _users
}
func (fd defaultUsersModel) Table(newTableName string) *defaultUsersModel {
	fd.usersDo.Conn.UseTable(newTableName)
	return fd.updateTableName(newTableName)
}

func (fd defaultUsersModel) As(alias string) *defaultUsersModel {
	fd.usersDo.Conn.DO = *(fd.usersDo.Conn.DO.As(alias).(*gen.DO))
	return fd.updateTableName(alias)
}

func (fd *defaultUsersModel) updateTableName(newTableName string) *defaultUsersModel {
	fd.ALL = field.NewAsterisk(newTableName)

	fd.Id = field.NewInt64(newTableName, "id")
	fd.Account = field.NewString(newTableName, "account")
	fd.NickName = field.NewString(newTableName, "nick_name")
	fd.Password = field.NewString(newTableName, "password")
	fd.DeletedAt = field.NewTime(newTableName, "deleted_at")

	fd.fillFieldMap()
	return fd
}

func (fd *defaultUsersModel) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := fd.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (fd *defaultUsersModel) fillFieldMap() {
	fd.fieldMap = make(map[string]field.Expr, 4)
	fd.fieldMap["account"] = fd.Account
	fd.fieldMap["nick_name"] = fd.NickName
	fd.fieldMap["password"] = fd.Password
	fd.fieldMap["deleted_at"] = fd.DeletedAt

	// 构建 OrderExpr 映射
	orderExprMap := make(map[string]field.OrderExpr, len(fd.fieldMap))
	for name, expr := range fd.fieldMap {
		if oe, ok := expr.(field.OrderExpr); ok {
			orderExprMap[name] = oe
		}
	}

	// 如果 pageExecutor 不存在，则创建；存在则更新其 fieldMap
	if fd.pageExecutor == nil {
		fd.pageExecutor = executor.NewPageExecutor[Users](orderExprMap, make(map[string]string))
	} else {
		fd.pageExecutor = executor.NewPageExecutor[Users](orderExprMap, fd.pageExecutor.GetDefaultOrderKeys())
	}
}

func (fd defaultUsersModel) clone(db *gorm.DB) defaultUsersModel {
	fd.usersDo.Conn.ReplaceConnPool(db.Statement.ConnPool)
	return fd
}

func (fd defaultUsersModel) replaceDB(db *gorm.DB) defaultUsersModel {
	fd.usersDo.Conn.ReplaceDB(db)
	return fd
}

func (m *defaultUsersModel) GetCacheKeys(data *Users) []string {
	if data == nil {
		return []string{}
	}
	gormzeroUsersIdKey := fmt.Sprintf("%s%v", cacheGormzeroUsersIdPrefix, data.Id)
	cacheKeys := []string{
		gormzeroUsersIdKey,
	}
	cacheKeys = append(cacheKeys, m.customCacheKeys(data)...)
	return cacheKeys
}

func (m *defaultUsersModel) Insert(ctx context.Context, tx *conn.ConnTx, data *Users) error {

	return m.usersDo.DoCtx(ctx, func(repo conn.Repository[Users]) error {
		repodb := repo
		if tx != nil {
			repodb.ReplaceDB(tx.UnderlyingDB())
		}
		return repodb.Create(data)
	}, m.GetCacheKeys(data)...)
}

func (m *defaultUsersModel) BatchInsert(ctx context.Context, tx *conn.ConnTx, news []*Users) error {

	err := m.usersDo.DoCtx(ctx, func(repo conn.Repository[Users]) error {
		repodb := repo
		if tx != nil {
			repodb.ReplaceDB(tx.UnderlyingDB())
		}
		return repodb.CreateInBatches(news, 500)
	})
	return err

}

func (m *defaultUsersModel) FindOne(ctx context.Context, id int64) (*Users, error) {

	forResult := func(repo conn.Repository[Users]) conn.Repository[Users] {
		return repo.Where(m.Id.Eq(id))
	}
	var resp Users
	gormzeroUsersIdKey := fmt.Sprintf("%s%v", cacheGormzeroUsersIdPrefix, id)

	err := m.usersDo.QueryCtx(ctx, gormzeroUsersIdKey, &resp, func(repo conn.Repository[Users]) error {

		return helper.Execute(forResult(repo).First, &resp)

	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
	//return result, nil
}

func (m *defaultUsersModel) FindPageList(ctx context.Context, page *pagex.PagePrams, orderBy pagex.OrderParams,
	orderKeys map[string]string, conds ...gen.Condition) ([]*Users, int64, error) {

	forResult := func(repo conn.Repository[Users]) conn.Repository[Users] {
		query := repo
		query = query.Where(conds...)
		return query
	}
	var resp []*Users
	var count int64

	err := m.usersDo.QueryNoCache(ctx, func(repo conn.Repository[Users]) error {
		query := forResult(repo)

		return m.pageExecutor.ExecutePageWithConditions(query, page, []pagex.OrderParams{orderBy}, orderKeys, &resp, &count)

	})
	if err != nil {
		return nil, 0, err
	}

	return resp, count, nil

}

func (m *defaultUsersModel) Update(ctx context.Context, tx *conn.ConnTx, data *Users) error {
	// 先获取旧数据用于缓存清理
	old, err := m.FindOne(ctx, data.Id)
	if err != nil {
		return err
	}

	// 收集需要清理的缓存键
	clearKeys := append(m.GetCacheKeys(old), m.GetCacheKeys(data)...)

	return m.usersDo.DoCtx(ctx, func(repo conn.Repository[Users]) error {
		query := repo
		if tx != nil {
			query.ReplaceDB(tx.UnderlyingDB())
		}
		// 使用 Save 方法进行更新
		return query.Save(data)
	}, clearKeys...)
}

func (m *defaultUsersModel) BatchUpdate(ctx context.Context, tx *conn.ConnTx, news []*Users) error {
	var opts []connx.DoBatchOption
	if tx != nil {
		opts = append(opts, connx.WithTx(tx))
	}

	return m.usersDo.DoBatchUpdate(ctx, news, func(repo conn.CommonRepo[Users], data *Users) error {
		return repo.Save(data)
	}, opts...)
}

func (m *defaultUsersModel) Delete(ctx context.Context, tx *conn.ConnTx, id int64) error {
	// 先获取数据用于缓存清理
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, connx.ErrNotFound) {
			return nil // 数据不存在，认为删除成功
		}
		return err
	}

	return m.usersDo.DoCtx(ctx, func(repo conn.Repository[Users]) error {
		query := repo
		if tx != nil {
			query.ReplaceDB(tx.UnderlyingDB())
		}
		// 使用软删除
		_, err := query.Where(m.Id.Eq(id)).Delete()
		return err
	}, m.GetCacheKeys(data)...)
}

func (m *defaultUsersModel) BatchDelete(ctx context.Context, tx *conn.ConnTx, datas []*Users) error {
	var opts []connx.DoBatchOption
	if tx != nil {
		opts = append(opts, connx.WithTx(tx))
	}

	return m.usersDo.DoBatchDelete(ctx, datas, func(repo conn.ConnOld[Users], data *Users) error {
		_, err := repo.Where(m.Id.Eq(data.Id)).Delete()
		return err
	}, opts...)
}
