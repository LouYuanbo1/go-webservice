package gormx

import (
	"context"
	"errors"

	"github.com/LouYuanbo1/go-webservice/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func startSpan(ctx context.Context, method string) (context.Context, oteltrace.Span) {
	tracer := trace.TracerFromContext(ctx)
	start, span := tracer.Start(ctx, method, oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	span.SetAttributes(attribute.Key("gormx.method").String(method))

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
