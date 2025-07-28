package tracing

import (
    "bytes"
    "io"
    "io/ioutil"
    "net/http"
    "time"

    "github.com/gorilla/mux"

    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/uber/jaeger-client-go/config"
)

// New creates an Opentracing tracer and attaches it to mux middleware.
// Returns Closer do be added to caller function as `defer closer.Close()`
func New(router *mux.Router) io.Closer {
    // Add Opentracing instrumentation
    defaultCfg := config.Configuration{
        ServiceName: "http-tracer",
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:            true,
            BufferFlushInterval: 1 * time.Second,
        },
    }
    cfg, err := defaultCfg.FromEnv()
    if err != nil {
        panic("Could not parse Jaeger env vars: " + err.Error())
    }
    tracer, closer, err := cfg.NewTracer()
    if err != nil {
        panic("Could not initialize jaeger tracer: " + err.Error())
    }

    opentracing.SetGlobalTracer(tracer)
    router.Use(TraceWithConfig(TraceConfig{
        Tracer: tracer,
    }))
    return closer
}

// Trace returns a Trace middleware.
// Trace middleware traces http requests and reporting errors.
func Trace(tracer opentracing.Tracer) mux.MiddlewareFunc {
    c := DefaultTraceConfig
    c.Tracer = tracer
    c.ComponentName = defaultComponentName
    return TraceWithConfig(c)
}

// TraceWithConfig returns a Trace middleware with config.
// See: `Trace()`.
func TraceWithConfig(config TraceConfig) mux.MiddlewareFunc {
    if config.Tracer == nil {
        panic("trace middleware requires opentracing tracer")
    }
    if config.ComponentName == "" {
        config.ComponentName = defaultComponentName
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            operationName := "HTTP " + r.Method + " URL: " + r.URL.Path
            var sp opentracing.Span
            tr := config.Tracer
            if ctx, err := tr.Extract(opentracing.HTTPHeaders,
                opentracing.HTTPHeadersCarrier(r.Header)); err != nil {
                sp = tr.StartSpan(operationName)
            } else {
                sp = tr.StartSpan(operationName, ext.RPCServerOption(ctx))
            }

            ext.HTTPMethod.Set(sp, r.Method)
            ext.HTTPUrl.Set(sp, r.URL.String())
            ext.Component.Set(sp, config.ComponentName)

            // response
            resBody := new(bytes.Buffer)
            mw := io.MultiWriter(w, resBody)
            writer := &responseWriter{Writer: mw, ResponseWriter: w}
            w = writer

            // Dump request
            if config.IsBodyDump {
                // request
                var reqBody []byte
                if r.Body != nil { // Read
                    reqBody, _ = ioutil.ReadAll(r.Body)
                    sp.SetTag("http.req.body", string(reqBody))
                }

                r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset
            }

            r = r.WithContext(opentracing.ContextWithSpan(r.Context(), sp))

            defer func() {
                status := writer.Status
                ext.HTTPStatusCode.Set(sp, uint16(status))
                if status >= http.StatusInternalServerError {
                    ext.Error.Set(sp, true)
                }

                // Dump response body
                if config.IsBodyDump {
                    sp.SetTag("http.resp.body", resBody.String())
                }

                sp.Finish()
            }()

            next.ServeHTTP(w, r)
        })
    }
}