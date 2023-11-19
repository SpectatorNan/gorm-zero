# gorm-zero
 go zero gorm extension

### If you use go zero, and you want to use Gorm. You can use this library.


# Usage

- add the dependent
```shell
go get github.com/SpectatorNan/gorm-zero
```
- replace  template/model in your project with gorm-zero/template/{goctl version}/model
- generate
```shell
goctl model mysql -src={patterns} -dir={dir} -cache --home ./template
```

## Mysql
### Config
```go
import (
    "github.com/SpectatorNan/gorm-zero/gormc/config/mysql"
)
type Config struct {
    Mysql mysql.Mysql
    ...
}
```
## Initialization
```go
import (
"github.com/SpectatorNan/gorm-zero/gormc/config/mysql"
)
func NewServiceContext(c config.Config) *ServiceContext {
    db, err := mysql.Connect(c.Mysql)
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## PgSql
### Config
```go
import (
"github.com/SpectatorNan/gorm-zero/gormc/config/pg"
)
type Config struct {
    PgSql pg.PgSql
    ...
}
```
## Initialization
```go
import (
"github.com/SpectatorNan/gorm-zero/gormc/config/pg"
)
func NewServiceContext(c config.Config) *ServiceContext {
    db, err := pg.Connect(c.PgSql)
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## Useage

### Query With Cache And Custom Expire Duration
```go
    gormzeroUsersIdKey := fmt.Sprintf("%s%v", cacheGormzeroUsersIdExpirePrefix, id)
    var resp Users
    err := m.QueryWithExpireCtx(ctx, &resp, gormzeroUsersIdKey, expire, func(conn *gorm.DB, v interface{}) error {
        return conn.Model(&Users{}).Where("`id` = ?", id).First(&resp).Error
    })
    switch err {
        case nil:
            return &resp, nil
        case gormc.ErrNotFound:
            return nil, ErrNotFound
        default:
            return nil, err
    }
```

### Query With Cache And Default Expire Duration
```go
    gormzeroUsersIdKey := fmt.Sprintf("%s%v", cacheGormzeroUsersIdPrefix, id)
    var resp Users
    err := m.QueryCtx(ctx, &resp, gormzeroUsersIdKey, func(conn *gorm.DB, v interface{}) error {
        return conn.Model(&Users{}).Where("`id` = ?", id).First(&resp).Error
    })
    switch err {
        case nil:
            return &resp, nil
        case gormc.ErrNotFound:
            return nil, ErrNotFound
        default:
            return nil, err
    }
```


## Usage Example
- go zero model example link: [gorm-zero-example](https://github.com/SpectatorNan/gorm-zero-example)
