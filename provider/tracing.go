package provider

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tpfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
)

const (
	serviceName = "terraform-provider-ctfd"
)

type OTelSetup struct {
	Shutdown       func(context.Context) error
	TracerProvider trace.TracerProvider
}

func SetupOTelSDK(ctx context.Context, version string) (out OTelSetup, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.

	out.Shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = multierr.Append(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned
	handleErr := func(inErr error) {
		err = multierr.Append(inErr, out.Shutdown(ctx))
	}

	// Set up propagator
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	// Ensure default SDK resources and the required service name are set
	r, err := resource.Merge(
		resource.Environment(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return
	}

	// Then create the span exporter
	exp, nerr := autoexport.NewSpanExporter(ctx)
	if err != nil {
		handleErr(nerr)
		return
	}
	shutdownFuncs = append(shutdownFuncs, exp.Shutdown)

	out.TracerProvider = sdktrace.NewTracerProvider(
		// We need to have the burden of a simple span processor as the process might be short-lived
		// because a batch processor can not give enough time to export data...
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exp)),
		sdktrace.WithResource(r),
	)

	return
}

func StartTFSpan(
	ctx context.Context,
	tracer trace.Tracer,
	obj any,
) (context.Context, trace.Span) {
	kind, typeName := "unknown", "unknown"
	if data, ok := obj.(datasource.DataSource); ok {
		kind = "data"

		resp := &datasource.MetadataResponse{}
		data.Metadata(ctx, datasource.MetadataRequest{}, resp)
		typeName = providerTypeName + resp.TypeName
	}
	if r, ok := obj.(tpfresource.Resource); ok {
		kind = "resource"

		resp := &tpfresource.MetadataResponse{}
		r.Metadata(ctx, tpfresource.MetadataRequest{}, resp)
		typeName = providerTypeName + resp.TypeName
	}

	method := getCallerFunctionName()

	return tracer.Start(
		ctx,
		fmt.Sprintf("%s/%s/%s", kind, typeName, method),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

func StartAPISpan(ctx context.Context, tracer trace.Tracer) (context.Context, trace.Span) {
	method := getCallerFunctionName()

	return tracer.Start(
		ctx,
		fmt.Sprintf("api/%s", method),
	)
}

func getCallerFunctionName() string {
	pc, _, _, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	method := "unknown"
	if fn != nil {
		if idx := strings.LastIndex(fn.Name(), "."); idx != -1 {
			method = fn.Name()[idx+1:]
		}
	}
	return method
}
