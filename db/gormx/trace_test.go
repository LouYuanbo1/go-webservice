package gormx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func setupTracerProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	t.Helper()

	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	otel.SetTracerProvider(tp)

	return tp, spanRecorder
}

func TestStartSpan(t *testing.T) {
	tp, spanRecorder := setupTracerProvider(t)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := context.Background()
	method := "Create"

	newCtx, span := startSpan(ctx, method)

	assert.NotNil(t, span)
	assert.NotEqual(t, oteltrace.SpanContext{}, span.SpanContext())
	assert.True(t, span.SpanContext().IsValid())

	span.End()

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	s := spans[0]
	assert.Equal(t, "Create", s.Name())
	assert.Equal(t, oteltrace.SpanKindClient, s.SpanKind())

	attrs := s.Attributes()
	var found bool
	for _, attr := range attrs {
		if string(attr.Key) == "gormx.method" {
			found = true
			assert.Equal(t, attribute.String("gormx.method", method), attr)
			break
		}
	}
	assert.True(t, found, "expected attribute gormx.method to be set")

	parentSpan := oteltrace.SpanFromContext(newCtx)
	assert.Equal(t, span.SpanContext().SpanID(), parentSpan.SpanContext().SpanID())
}

func TestStartSpan_WithParentContext(t *testing.T) {
	tp, spanRecorder := setupTracerProvider(t)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")
	ctx, parentSpan := tracer.Start(context.Background(), "parent")

	newCtx, childSpan := startSpan(ctx, "Find")

	childSpan.End()
	parentSpan.End()

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 2)

	var childSpanData, parentSpanData sdktrace.ReadOnlySpan
	for _, s := range spans {
		if s.Name() == "Find" {
			childSpanData = s
		}
		if s.Name() == "parent" {
			parentSpanData = s
		}
	}

	assert.NotNil(t, childSpanData)
	assert.NotNil(t, parentSpanData)
	assert.Equal(t, parentSpanData.SpanContext().SpanID(), childSpanData.Parent().SpanID())

	_ = newCtx
}

func TestEndSpan_NilError(t *testing.T) {
	tp, spanRecorder := setupTracerProvider(t)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test")

	endSpan(span, nil)

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	s := spans[0]
	assert.Equal(t, codes.Ok, s.Status().Code)
}

func TestEndSpan_ErrRecordNotFound(t *testing.T) {
	tp, spanRecorder := setupTracerProvider(t)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test")

	endSpan(span, gorm.ErrRecordNotFound)

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	s := spans[0]
	assert.Equal(t, codes.Ok, s.Status().Code)
}

func TestEndSpan_GeneralError(t *testing.T) {
	tp, spanRecorder := setupTracerProvider(t)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test")

	err := errors.New("something went wrong")
	endSpan(span, err)

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	s := spans[0]
	assert.Equal(t, codes.Error, s.Status().Code)
	assert.Equal(t, "something went wrong", s.Status().Description)
}

func TestEndSpan_WrappedErrRecordNotFound(t *testing.T) {
	tp, spanRecorder := setupTracerProvider(t)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test")

	wrappedErr := errors.Join(gorm.ErrRecordNotFound, errors.New("extra"))
	endSpan(span, wrappedErr)

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	s := spans[0]
	assert.Equal(t, codes.Ok, s.Status().Code)
}
