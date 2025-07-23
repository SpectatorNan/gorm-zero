package executor

import (
	"github.com/SpectatorNan/gorm-zero/v2/conn"
	"github.com/SpectatorNan/gorm-zero/v2/pagex"
	"gorm.io/gen/field"
)

type PageExecutor[T any] struct {
	defaultOrderKeys map[string]string
	fieldMap         map[string]field.OrderExpr
}

func NewPageExecutor[T any](fieldMap map[string]field.OrderExpr, defaultOrderKeys map[string]string) *PageExecutor[T] {
	return &PageExecutor[T]{
		fieldMap:         fieldMap,
		defaultOrderKeys: defaultOrderKeys,
	}
}

// GetDefaultOrderKeys 获取默认的排序键映射
func (pe *PageExecutor[T]) GetDefaultOrderKeys() map[string]string {
	return pe.defaultOrderKeys
}

// 执行带条件的分页查询
func (pe *PageExecutor[T]) ExecutePageWithConditions(
	query conn.Repository[T],
	page *pagex.PagePrams,
	orderBy []pagex.OrderParams,
	orderKeys map[string]string, // 如果为空则使用默认的
	target *[]*T,
	count *int64,
) error {
	// 使用传入的 orderKeys，如果为空则使用默认的
	keys := orderKeys
	if keys == nil {
		keys = pe.defaultOrderKeys
	}

	// 应用排序
	sortedQuery := pe.applyOrder(query, orderBy, keys)

	// 执行分页查询
	res, total, err := sortedQuery.FindByPage(page.Offset(), page.Limit())
	if err != nil {
		return err
	}

	*target = res
	*count = total
	return nil
}

// 应用排序逻辑
func (pe *PageExecutor[T]) applyOrder(
	query conn.Repository[T],
	orderBys []pagex.OrderParams,
	orderKeys map[string]string,
) conn.Repository[T] {
	orderExprs := applyOrderBys(pe.fieldMap, orderBys, orderKeys)
	if len(orderExprs) > 0 {
		query = query.Order(orderExprs...)
	}
	return query
}

func applyOrderBys(fieldMap map[string]field.OrderExpr, orderBys []pagex.OrderParams, orderKeys map[string]string) []field.Expr {
	var exprs []field.Expr
	for _, orderBy := range orderBys {
		// 获取用于查找字段的 key
		fieldKey := orderBy.OrderKey

		// 如果 orderKeys 不为空，先检查是否有映射配置
		if orderKeys != nil && len(orderKeys) > 0 {
			if mappedKey, exists := orderKeys[orderBy.OrderKey]; exists {
				fieldKey = mappedKey
			}
		}

		// 使用 fieldKey 在 fieldMap 中查找字段表达式
		if fieldExpr, ok := fieldMap[fieldKey]; ok {
			if orderBy.Sort == pagex.Desc() {
				exprs = append(exprs, fieldExpr.Desc())
			} else if orderBy.Sort == pagex.Asc() {
				exprs = append(exprs, fieldExpr.Asc())
			}
		}
	}
	return exprs
}

// 支持 Scan 的分页查询
func (pe *PageExecutor[T]) ExecuteScanPageWithConditions(
	query conn.Repository[T],
	result interface{},
	page *pagex.PagePrams,
	orderBys []pagex.OrderParams,
	orderKeys map[string]string,
	count *int64,
) error {
	keys := orderKeys
	if keys == nil {
		keys = pe.defaultOrderKeys
	}

	sortedQuery := pe.applyOrder(query, orderBys, keys)
	total, err := sortedQuery.ScanByPage(result, page.Offset(), page.Limit())
	if err != nil {
		return err
	}

	*count = total
	return nil
}
