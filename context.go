package opentracefasthttp

import (
	"github.com/opentracing/opentracing-go"
	"github.com/valyala/fasthttp"
)

const contextKey = "github.com/wawan93/opentracefasthttp span"

// ContextWithSpan returns a new `*fasthttp.RequestCtx` that holds a reference to
// the span. If span is nil, a new context without an active span is returned.
func ContextWithSpan(ctx *fasthttp.RequestCtx, span opentracing.Span) *fasthttp.RequestCtx {
	if ctx == nil {
		return nil
	}
	ctx.SetUserValue(contextKey, span)
	return ctx
}

// SpanFromContext returns the `opentracing.Span` previously associated with `ctx`, or
// `nil` if no such `opentracing.Span` could be found.
func SpanFromContext(ctx *fasthttp.RequestCtx) opentracing.Span {
	v := ctx.UserValue(contextKey)
	if span, ok := v.(opentracing.Span); ok {
		return span
	}
	return nil
}
