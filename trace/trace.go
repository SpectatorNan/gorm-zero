package gormx

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// traceName is used to identify the trace name for the SQL execution.
	traceName = "gorm-zero"

	// spanName is used to identify the span name for the SQL execution.
	//spanName = "sql"
)

type Span struct {
	name         string
	attributeKey attribute.Key
	ignoreError  error
}

// var sqlAttributeKey = attribute.Key("gorm-zero.conn.method")
func SpanFrom(name string, key string) Span {
	return Span{
		name:         name,
		attributeKey: attribute.Key(key),
	}
}

func (s Span) With(ctx context.Context, method string, fn func(ctx context.Context) error) error {
	var err error
	ctx, span := s.start(ctx, method)
	defer s.end(span, nil)
	err = fn(ctx)
	return err
}

func (s Span) start(ctx context.Context, method string) (context.Context, oteltrace.Span) {
	tracer := otel.Tracer(traceName)
	start, span := tracer.Start(ctx,
		s.name,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)
	span.SetAttributes(s.attributeKey.String(method))

	return start, span
}

func (s Span) end(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil || errors.Is(err, s.ignoreError) {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
