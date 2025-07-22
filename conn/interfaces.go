package conn

import (
	"context"
	"database/sql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Repository[T any] interface {
	// gen parent interface
	gen.SubQuery

	// DB
	Debug() Repository[T]
	WithContext(ctx context.Context) Repository[T]
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() Repository[T]
	WriteDB() Repository[T]
	As(alias string) gen.Dao
	Session(config *gorm.Session) Repository[T]

	// operations
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) Repository[T]
	Not(conds ...gen.Condition) Repository[T]
	Or(conds ...gen.Condition) Repository[T]
	Select(conds ...field.Expr) Repository[T]
	Where(conds ...gen.Condition) Repository[T]
	Order(conds ...field.Expr) Repository[T]
	Distinct(cols ...field.Expr) Repository[T]
	Omit(cols ...field.Expr) Repository[T]
	Join(table schema.Tabler, on ...field.Expr) Repository[T]
	LeftJoin(table schema.Tabler, on ...field.Expr) Repository[T]
	RightJoin(table schema.Tabler, on ...field.Expr) Repository[T]
	Group(cols ...field.Expr) Repository[T]
	Having(conds ...gen.Condition) Repository[T]
	Limit(limit int) Repository[T]
	Offset(offset int) Repository[T]
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) Repository[T]
	Unscoped() Repository[T]

	// execute
	Create(values ...*T) error
	CreateInBatches(values []*T, batchSize int) error
	Save(values ...*T) error
	First() (*T, error)
	Take() (*T, error)
	Last() (*T, error)
	Find() ([]*T, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*T, err error)
	FindInBatches(result *[]*T, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*T) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao

	//
	Attrs(attrs ...field.AssignExpr) Repository[T]
	Assign(attrs ...field.AssignExpr) Repository[T]
	Joins(fields ...field.RelationField) Repository[T]
	Preload(fields ...field.RelationField) Repository[T]
	FirstOrInit() (*T, error)
	FirstOrCreate() (*T, error)
	FindByPage(offset int, limit int) (result []*T, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) Repository[T]
	UnderlyingDB() *gorm.DB
	schema.Tabler
}
