// +build jaeger

// test
package opentracefasthttp_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/d7561985/opentracefasthttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

const JaegerURL = "http://localhost:16686"

func TestCarrier(t *testing.T) {
	waitJaeger()

	// we need real trace connection
	cfg, err := config.FromEnv()
	assert.NoError(t, err)
	cfg.ServiceName = "fasthttp-carrier-tester"
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	tr, cl, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	assert.NoError(t, err)
	defer cl.Close() //nolint
	opentracing.SetGlobalTracer(tr)

	ok := false
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close() //nolint

	srv := fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		carrier := opentracefasthttp.New(&ctx.Request.Header)
		clientContext, err := tr.Extract(opentracing.HTTPHeaders, carrier)
		assert.NoError(t, err)
		span := tr.StartSpan("HTTP "+string(ctx.Method())+" "+ctx.Request.URI().String(), ext.RPCServerOption(clientContext))
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
	err = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	assert.NoError(t, err)

	client := fasthttp.Client{Dial: func(addr string) (net.Conn, error) {
		return ln.Dial()
	}}
	err = client.Do(req, nil)
	assert.NoError(t, err)

	assert.True(t, ok)
}

func waitJaeger() {
	res := make(chan struct{})
	go checkJaeger(res)

	select {

	case <-time.After(time.Minute):
		panic("no connection")
	case <-res:
		fmt.Println("Jaeger has found")
		return
	}
}

func checkJaeger(res chan<- struct{}) {
	request, err := http.NewRequest("GET", JaegerURL, nil)
	if err != nil {
		panic(err)
	}

	for {
		if response, err := http.DefaultClient.Do(request); err == nil && response.StatusCode == http.StatusOK {
			res <- struct{}{}
			return
		}
		time.Sleep(time.Second)
	}
}
