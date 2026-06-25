package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const TraceName = "go-webservice"

func NewTracerProvider() (*sdkTrace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(
			traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			sdkTrace.WithBatchTimeout(time.Second),
		),
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

func TracerFromContext(ctx context.Context) (tracer oteltrace.Tracer) {
	if span := oteltrace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		tracer = span.TracerProvider().Tracer(TraceName)
	} else {
		tracer = otel.Tracer(TraceName)
	}

	return
}
