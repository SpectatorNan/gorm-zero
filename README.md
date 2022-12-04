# gorm-zero
 go zero gorm extension

### If you use go zero, and you want to use Gorm. You can use this library.

## It is recommended to use version v1.0.2, the V1.0.3 needs to wait for go-zero to merge core dependencies

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


## Usage Example
- go zero model example link: [gorm-zero-example](https://github.com/SpectatorNan/gorm-zero-example)
