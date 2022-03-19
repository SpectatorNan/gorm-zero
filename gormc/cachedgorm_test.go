package gormc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stat"
)

func init() {
	logx.Disable()
	stat.SetReporter(nil)
}
