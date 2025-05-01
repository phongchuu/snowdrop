package snowdrop

import (
	"context"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type tracerContextKey int

const tracerCtxID tracerContextKey = iota

var ErrNoTracer = errors.New("there is no trace.Tracer in the given context")

func WithTracer(r *http.Request, tracer trace.Tracer) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), tracerCtxID, tracer))
}

//nolint:ireturn
func GetTracer(r *http.Request) (trace.Tracer, error) {
	return GetTracerFromCtx(r.Context())
}

//nolint:ireturn
func GetTracerFromCtx(ctx context.Context) (trace.Tracer, error) {
	if tracer, ok := ctx.Value(tracerCtxID).(trace.Tracer); ok {
		return tracer, nil
	}

	return nil, ErrNoTracer
}

//nolint:ireturn
func MustGetTracer(r *http.Request) trace.Tracer {
	return MustGetTracerFromCtx(r.Context())
}

//nolint:ireturn
func MustGetTracerFromCtx(ctx context.Context) trace.Tracer {
	tracer, err := GetTracerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	return tracer
}
