package gormc

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"time"
)

func init() {
	logx.Disable()
	stat.SetReporter(nil)
}

type mysqlcfg struct {
	Path         string // 服务器地址
	Port         string `json:",default=3306"`                                               // 端口
	Config       string `json:",default=charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal"` // 高级配置
	Dbname       string // 数据库名
	Username     string // 数据库用户名
	Password     string // 数据库密码
	MaxIdleConns int    `json:",default=10"` // 空闲中的最大连接数
	MaxOpenConns int    `json:",default=10"` // 打开到数据库的最大连接数
	LogMode      string `json:",default="`   // 是否开启Gorm全局日志
	LogZap       bool   // 是否通过zap写入日志文件
}

func (m *mysqlcfg) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}
func TestGormc_QueryWithExpire(t *testing.T) {

	cfg := mysqlcfg{
		Path:     "localhost",
		Port:     "3306",
		Config:   "charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal",
		Dbname:   "gormzero",
		Username: "root",
		Password: "root",
	}
	mcg := mysql.Config{
		DSN: cfg.Dsn(),
	}
	db, err := gorm.Open(mysql.New(mcg))
	if err != nil {
		t.Error(err)
		return
	}
	ccf := cache.CacheConf{
		cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: "127.0.0.1:6379",
				Pass: "",
			},
			Weight: 100,
		},
	}
	gormc := NewConn(db, ccf)
	var str string
	err = gormc.QueryWithExpireCtx(context.Background(), &str, "any", time.Second*5, func(conn *gorm.DB, v interface{}) error {
		*v.(*string) = "value"
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}

}
