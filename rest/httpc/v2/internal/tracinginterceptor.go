package internal

import (
	"net/http"

	ztrace "github.com/shippomx/zard/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TracingInterceptor(r *http.Request, _ ExtendInfo) (*http.Request, ResponseHandler) {
	if r == nil {
		return nil, func(_ *http.Response, _ error) {}
	}
	tracer := otel.Tracer(ztrace.TraceName)
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	spanCtx, span := tracer.Start(
		ctx,
		r.Host,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
			"httpc-"+r.Host, r.Host, r)...),
	)
	propagator.Inject(spanCtx, propagation.HeaderCarrier(r.Header))
	r = r.WithContext(spanCtx)
	return r, func(resp *http.Response, err error) {
		defer span.End()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		if resp == nil {
			return
		}
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(resp.StatusCode, oteltrace.SpanKindClient))
	}
}
