[![Build Status](https://travis-ci.org/d7561985/opentracefasthttp.svg?branch=master)](https://travis-ci.org/d7561985/opentracefasthttp)

# opentracefasthttp
[Opentracing](github.com/opentracing/opentracing-go) carrier for [fasthttp](https://github.com/valyala/fasthttp) server. Gives possibility to use span extract/inject options


# examples

## client send with request's header
```go
	req := fasthttp.AcquireRequest()
	...
	
	carrier := opentracefasthttp.New(&req.Header)
	err = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
```

## 

## server read from request's header
```go
	func(ctx *fasthttp.RequestCtx) {
		carrier := opentracefasthttp.New(&ctx.Request.Header)
		clientContext, err := tr.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil{
			...
		}
		span := trace.StartSpan("HTTP "+string(ctx.Method())+" "+ctx.Request.URI().String(), ext.RPCServerOption(clientContext))
	}
```