package pagex

import (
	"context"
	"fmt"
	"testing"

	"github.com/SpectatorNan/gorm-zero/gormc"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

type TestUserModel struct {
	Id       uint64 `gorm:"column:id;primary_key"`
	Age      int8   `gorm:"column:age"`
	Name     string `gorm:"column:name"`     // The username
	Nickname string `gorm:"column:nickname"` // The nickname
	Avatar   string `gorm:"column:avatar"`
	Email    string `gorm:"column:email"`
}

func (TestUserModel) TableName() string {
	return "user"
}

func TestFindPageList(t *testing.T) {
	cfg := mysqlcfg{
		Path:     "localhost",
		Port:     "3306",
		Config:   "charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal",
		Dbname:   "gormzero",
		Username: "root",
		Password: "123456",
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
				Type: "node",
			},
			Weight: 100,
		},
	}
	gormc := gormc.NewConn(db, ccf)
	users, cnt, err := FindPageList[TestUserModel](
		context.Background(),
		gormc,
		&ListReq{Page: 1, PageSize: 5},
		[]OrderBy{
			{OrderKey: "age", Sort: "asc"},
			{OrderKey: "id", Sort: "asc"},
		},
		func(conn *gorm.DB) (*gorm.DB, *gorm.DB) {
			d := db.Model(&TestUserModel{})
			return d, d
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageList Err,%v", err.Error())
	}
	fmt.Println(users, cnt)
}

func TestFindPageListWithCount(t *testing.T) {
	cfg := mysqlcfg{
		Path:     "localhost",
		Port:     "3306",
		Config:   "charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal",
		Dbname:   "gormzero",
		Username: "root",
		Password: "123456",
	}
	mcg := mysql.Config{
		DSN: cfg.Dsn(),
	}

	db, err := gorm.Open(mysql.New(mcg))
	if err != nil {
		t.Error(err)
		return
	}

	users, cnt, err := FindPageListWithCount[TestUserModel](
		context.Background(),
		&ListReq{Page: 1, PageSize: 5},
		[]OrderBy{
			{OrderKey: "age", Sort: "asc"},
			{OrderKey: "id", Sort: "asc"},
		},
		func() (*gorm.DB, *gorm.DB) {
			d := db.Model(&TestUserModel{})
			return d, nil
		},
	)
	if err != nil {
		t.Fatalf("TestFindPageListWithCount Err,%v", err.Error())
	}
	fmt.Println(users, cnt)
}
