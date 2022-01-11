package gormx

import (
	"github.com/tal-tech/go-zero/core/logx"
	"strings"
)

func desensitize(datasource string) string {
	// remove account
	pos := strings.LastIndex(datasource, "@")
	if 0 <= pos && pos+1 < len(datasource) {
		datasource = datasource[pos+1:]
	}

	return datasource
}

func logInstanceError(datasource string, err error) {
	datasource = desensitize(datasource)
	logx.Errorf("Error on getting sql instance of %s: %v", datasource, err)
}