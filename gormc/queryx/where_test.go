package queryx

import (
	"fmt"
	"testing"
)

func TestChangeInSqlToEq(t *testing.T) {
	example := "select * from table where column in ? and another_column in ?"
	result := replaceLastInWithEqual(example)
	t.Logf("result: %v", result)
}

func TestWhereInt64s(t *testing.T) {
	where := "column in ?"
	val := []int64{1}
	result := whereInt64s(where, val)
	t.Logf("result: %v", result)
}

func whereInt64s(whereStr string, val []int64) string {
	if len(val) > 1 {
		return whereStr
	} else if len(val) == 1 {
		whereStr = replaceLastInWithEqual(whereStr)
		fmt.Println("whereStr", whereStr)
		return whereStr
	}
	return whereStr
}
