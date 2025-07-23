package conn

import (
	"context"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

type CommonRepo[T any] struct {
	gen.DO
}

func (r CommonRepo[T]) Debug() Repository[T] {
	return r.withDO(r.DO.Debug())
}

func (r CommonRepo[T]) WithContext(ctx context.Context) Repository[T] {
	return r.withDO(r.DO.WithContext(ctx))
}

func (r CommonRepo[T]) ReadDB() Repository[T] {
	return r.Clauses(dbresolver.Read)
}

func (r CommonRepo[T]) WriteDB() Repository[T] {
	return r.Clauses(dbresolver.Write)
}

func (r CommonRepo[T]) Session(config *gorm.Session) Repository[T] {
	return r.withDO(r.DO.Session(config))
}

func (r CommonRepo[T]) Clauses(conds ...clause.Expression) Repository[T] {
	return r.withDO(r.DO.Clauses(conds...))
}

func (r CommonRepo[T]) Returning(value interface{}, columns ...string) Repository[T] {
	return r.withDO(r.DO.Returning(value, columns...))
}

func (r CommonRepo[T]) Not(conds ...gen.Condition) Repository[T] {
	return r.withDO(r.DO.Not(conds...))
}

func (r CommonRepo[T]) Or(conds ...gen.Condition) Repository[T] {
	return r.withDO(r.DO.Or(conds...))
}

func (r CommonRepo[T]) Select(conds ...field.Expr) Repository[T] {
	return r.withDO(r.DO.Select(conds...))
}

func (r CommonRepo[T]) Where(conds ...gen.Condition) Repository[T] {
	return r.withDO(r.DO.Where(conds...))
}

func (r CommonRepo[T]) Order(conds ...field.Expr) Repository[T] {
	return r.withDO(r.DO.Order(conds...))
}

func (r CommonRepo[T]) Distinct(cols ...field.Expr) Repository[T] {
	return r.withDO(r.DO.Distinct(cols...))
}

func (r CommonRepo[T]) Omit(cols ...field.Expr) Repository[T] {
	return r.withDO(r.DO.Omit(cols...))
}

func (r CommonRepo[T]) Join(table schema.Tabler, on ...field.Expr) Repository[T] {
	return r.withDO(r.DO.Join(table, on...))
}

func (r CommonRepo[T]) LeftJoin(table schema.Tabler, on ...field.Expr) Repository[T] {
	return r.withDO(r.DO.LeftJoin(table, on...))
}

func (r CommonRepo[T]) RightJoin(table schema.Tabler, on ...field.Expr) Repository[T] {
	return r.withDO(r.DO.RightJoin(table, on...))
}

func (r CommonRepo[T]) Group(cols ...field.Expr) Repository[T] {
	return r.withDO(r.DO.Group(cols...))
}

func (r CommonRepo[T]) Having(conds ...gen.Condition) Repository[T] {
	return r.withDO(r.DO.Having(conds...))
}

func (r CommonRepo[T]) Limit(limit int) Repository[T] {
	return r.withDO(r.DO.Limit(limit))
}

func (r CommonRepo[T]) Offset(offset int) Repository[T] {
	return r.withDO(r.DO.Offset(offset))
}

func (r CommonRepo[T]) Scopes(funcs ...func(gen.Dao) gen.Dao) Repository[T] {
	return r.withDO(r.DO.Scopes(funcs...))
}

func (r CommonRepo[T]) Unscoped() Repository[T] {
	return r.withDO(r.DO.Unscoped())
}

func (r CommonRepo[T]) Create(values ...*T) error {
	if len(values) == 0 {
		return nil
	}
	return r.DO.Create(values)
}

func (r CommonRepo[T]) CreateInBatches(values []*T, batchSize int) error {
	return r.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (r CommonRepo[T]) Save(values ...*T) error {
	if len(values) == 0 {
		return nil
	}
	return r.DO.Save(values)
}

func (r CommonRepo[T]) First() (*T, error) {
	if result, err := r.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*T), nil
	}
}

func (r CommonRepo[T]) Take() (*T, error) {
	if result, err := r.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*T), nil
	}
}

func (r CommonRepo[T]) Last() (*T, error) {
	if result, err := r.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*T), nil
	}
}

func (r CommonRepo[T]) Find() ([]*T, error) {
	result, err := r.DO.Find()
	return result.([]*T), err
}

func (r CommonRepo[T]) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*T, err error) {
	buf := make([]*T, 0, batchSize)
	err = r.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (r CommonRepo[T]) FindInBatches(result *[]*T, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return r.DO.FindInBatches(result, batchSize, fc)
}

func (r CommonRepo[T]) Attrs(attrs ...field.AssignExpr) Repository[T] {
	return r.withDO(r.DO.Attrs(attrs...))
}

func (r CommonRepo[T]) Assign(attrs ...field.AssignExpr) Repository[T] {
	return r.withDO(r.DO.Assign(attrs...))
}

func (r CommonRepo[T]) Joins(fields ...field.RelationField) Repository[T] {
	for _, _f := range fields {
		r = *r.withDO(r.DO.Joins(_f))
	}
	return &r
}

func (r CommonRepo[T]) Preload(fields ...field.RelationField) Repository[T] {
	for _, _f := range fields {
		r = *r.withDO(r.DO.Preload(_f))
	}
	return &r
}

func (r CommonRepo[T]) FirstOrInit() (*T, error) {
	if result, err := r.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*T), nil
	}
}

func (r CommonRepo[T]) FirstOrCreate() (*T, error) {
	if result, err := r.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*T), nil
	}
}

func (r CommonRepo[T]) FindByPage(offset int, limit int) (result []*T, count int64, err error) {
	result, err = r.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = r.Offset(-1).Limit(-1).Count()
	return
}

func (r CommonRepo[T]) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = r.Count()
	if err != nil {
		return
	}

	err = r.Offset(offset).Limit(limit).Scan(result)
	return
}

func (r CommonRepo[T]) Scan(result interface{}) (err error) {
	return r.DO.Scan(result)
}

func (r CommonRepo[T]) Delete(models ...*T) (result gen.ResultInfo, err error) {
	return r.DO.Delete(models)
}

func (r *CommonRepo[T]) withDO(do gen.Dao) *CommonRepo[T] {
	// 创建新的 CommonRepo 实例而不是修改现有的
	r.DO = *do.(*gen.DO)
	return r
	//newRepo := &CommonRepo[T]{}
	//newRepo.DO = *do.(*gen.DO)
	//return newRepo
}
