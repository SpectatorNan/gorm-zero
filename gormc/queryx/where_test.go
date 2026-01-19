package queryx

import (
	"fmt"
	"testing"
)

func TestChangeInSqlToEq(t *testing.T) {
	example := "select * from table where column in ? and another_column in ?"
	result := normalizeINToEquals(example)
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
		whereStr = normalizeINToEquals(whereStr)
		fmt.Println("whereStr", whereStr)
		return whereStr
	}
	return whereStr
}

func TestNormalizeINToEquals(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  string // expected for normalizeINToEquals (SQL-oriented)
	}

	cases := []testCase{
		{"id IN ?", "id IN ?", "id = ?"},
		{"pin in tin", "pin in tin", "pin = tin"},
		{"pin=inline", "pin=inline", "pin=inline"},
		{"pin In ?", "pin In ?", "pin = ?"},
		{"begin in end", "begin in end", "begin = end"},
		// non-standard or edge cases: still handled if 'in' is a standalone token
		{"IN foo (no leading space)", "IN foo", "IN foo"},
		{"something in (no trailing space)", "something in", "something in"},
		{"extra spaces", "col   IN    ?", "col = ?"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeINToEquals(tc.input)
			if got != tc.want {
				t.Errorf("normalizeINToEquals(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
