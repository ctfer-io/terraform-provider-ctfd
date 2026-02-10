package provider

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Option interface {
	apply(*options)
}

type options struct {
	tracer trace.TracerProvider
}

type tracerOption struct {
	tracer trace.TracerProvider
}

func (opt tracerOption) apply(opts *options) {
	opts.tracer = opt.tracer
}

func WithTracerProvider(tracer trace.TracerProvider) Option {
	return &tracerOption{
		tracer: tracer,
	}
}

func getTracer(opts ...Option) trace.Tracer {
	o := &options{
		tracer: nil,
	}
	for _, opt := range opts {
		opt.apply(o)
	}

	if o.tracer == nil {
		o.tracer = otel.GetTracerProvider()
	}
	return o.tracer.Tracer(serviceName)
}
