package opentracefasthttp_test

import (
	"net"
	"testing"

	"github.com/d7561985/opentracefasthttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestCarrier(t *testing.T) {
	tracer, closer := jaeger.NewTracer("fasthttp-carrier-tester", jaeger.NewConstSampler(true), jaeger.NewNullReporter())
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	ok := false
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close() //nolint

	srv := fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		carrier := opentracefasthttp.New(&ctx.Request.Header)
		clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		assert.NoError(t, err)
		span := tracer.StartSpan("HTTP "+string(ctx.Method())+" "+ctx.Request.URI().String(), ext.RPCServerOption(clientContext))
		assert.NotNil(t, span)
		span.LogFields(log.String("server", "request ok"))
		defer span.Finish()

		ok = true
	}}
	go srv.Serve(ln) //nolint

	span := opentracing.GlobalTracer().StartSpan("client-request")
	defer span.Finish()

	span.SetTag("test", "test")
	span.LogFields(log.String("test", "test"))

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://example.com")

	carrier := opentracefasthttp.New(&req.Header)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	assert.NoError(t, err)

	client := fasthttp.Client{Dial: func(addr string) (net.Conn, error) {
		return ln.Dial()
	}}
	err = client.Do(req, nil)
	assert.NoError(t, err)

	assert.True(t, ok)
}
