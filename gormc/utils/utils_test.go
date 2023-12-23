package utils

import (
	"strings"
	"testing"
)

func TestFileWithLineNum(t *testing.T) {
	t.Log(FileWithLineNum())
}

func TestContain(t *testing.T) {
	str := "/Users/spec/Documents/go/pkg/mod/gorm.io/gorm@v1.25.5/callbacks.go"
	contain := strings.Contains(str, "gorm.io")
	t.Log(contain)
	str2 := "/Users/spec/Documents/GitHub/gorm-zero/gormc/pagex/pageCombine.go"
	contain2 := strings.Contains(str2, "/Users/spec/Documents/GitHub/gorm-zero/")
	t.Log(contain2)
}
