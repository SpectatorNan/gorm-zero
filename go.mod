module github.com/SpectatorNan/gorm-zero

go 1.16

require (
	github.com/zeromicro/go-zero v1.5.3
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
	gorm.io/driver/mysql v1.4.7
	gorm.io/driver/postgres v1.5.2
	gorm.io/gorm v1.25.1
)

//replace github.com/zeromicro/go-zero v1.4.2 => github.com/SpectatorNan/go-zero v1.2.5-0.20221201151248-db1f09d9826d
