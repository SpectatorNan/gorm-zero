package plugins

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const (
	spanName           = "gorm-zero"
	gormSpanKey        = "gorm-zero-span"
	callBackBeforeName = "gorm-zero-trace:before"
	callBackAfterName  = "gorm-zero-trace:after"

	spanEventName        = "gorm-zero-event"
	spanAttrTable        = attribute.Key("gorm.table")
	spanAttrSql          = attribute.Key("gorm.sql")
	spanAttrRowsAffected = attribute.Key("gorm.rowsAffected")
)

type TracingPlugin struct{}

func (gp *TracingPlugin) Name() string {
	return "gorm-zero-tracing-plugin"
}

func (gp *TracingPlugin) Initialize(db *gorm.DB) (err error) {
	// 开始前 - 并不是都用相同的方法，可以自己自定义
	db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before)
	db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before)
	db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before)
	db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, before)
	db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before)
	db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before)

	// 结束后 - 并不是都用相同的方法，可以自己自定义
	db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after)
	db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after)
	db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after)
	db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after)
	db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)
	db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)
	return
}

// 告诉编译器这个结构体实现了gorm.Plugin接口
var _ gorm.Plugin = &TracingPlugin{}

func before(db *gorm.DB) {
	_, span := startSpan(db.Statement.Context)
	// 利用db实例去传递span
	db.InstanceSet(gormSpanKey, span)
}

func after(db *gorm.DB) {
	// 从GORM的DB实例中取出span
	_span, isExist := db.InstanceGet(gormSpanKey)
	if !isExist {
		// 不存在就直接抛弃掉
		return
	}

	// 断言进行类型转换
	span, ok := _span.(oteltrace.Span)
	if !ok {
		return
	}

	defer func() {
		endSpan(span, db.Error)
	}()

	span.AddEvent(spanEventName, oteltrace.WithAttributes(
		spanAttrTable.String(db.Statement.Table),
		spanAttrRowsAffected.Int64(db.RowsAffected),
		spanAttrSql.String(db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)),
	))
}

func startSpan(ctx context.Context) (context.Context, oteltrace.Span) {
	// tracer := otel.Tracer(trace.TraceName)
	tracer := trace.TracerFromContext(ctx)
	start, span := tracer.Start(ctx,
		spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)

	return start, span
}

func endSpan(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
