package tracing

import (
    "context"
    "fmt"
    "net/http"
    "runtime"

    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/log"
)

// Span is copy of opentracing.Span interface. It helps us to
// wrap the opentracing.Span and expose the local span to other
// layers.
//
// Span represents an active, un-finished span in the OpenTracing
// system. The function that create the span is responsible to
// finish it.
type Span interface {
    Finish()
    FinishWithOptions(opts opentracing.FinishOptions)
    Context() opentracing.SpanContext
    SetOperationName(operationName string) opentracing.Span
    SetTag(key string, value interface{}) opentracing.Span
    LogFields(fields ...log.Field)
    LogKV(alternatingKeyValues ...interface{})
    SetBaggageItem(restrictedKey string, value string) opentracing.Span
    BaggageItem(restrictedKey string) string
    Tracer() opentracing.Tracer
    LogEvent(event string)
    LogEventWithPayload(event string, payload interface{})
    Log(data opentracing.LogData)
}

// CreateSpan creates a new opentracing span adding tags for the
// span name and caller details. It accepts a context and an
// operation name. The span will be attached to the context, in
// this way we can create a nested spans. The operation name can
// be a function name, service name, or application layer name.
//
// User must call defer `sp.Finish()`
func CreateSpan(ctx context.Context, operationName string) (Span, context.Context) {
    sp, ctx := opentracing.StartSpanFromContext(ctx, operationName)
    sp.SetTag("name", operationName)

    // Get caller function name, file and line
    pc := make([]uintptr, 15)
    n := runtime.Callers(2, pc)
    frames := runtime.CallersFrames(pc[:n])
    frame, _ := frames.Next()
    callerDetails := fmt.Sprintf("%s - %s#%d", frame.Function, frame.File, frame.Line)
    sp.SetTag("caller", callerDetails)

    return sp, ctx
}

// LogSpanError attaches message and error to the span.
func LogSpanError(span Span, message string, err error) {
    if err != nil {
        span.SetTag("error", true)
        span.LogFields(
            log.String("message", message),
            log.Error(err),
            log.String("stack", fmt.Sprintf("%+v", err)),
        )
    }
}

// InjectHTTPCarrier extracts trace ID from span and injects it
// into the http header.
func InjectHTTPCarrier(span Span, req *http.Request) error {
    return opentracing.GlobalTracer().Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(req.Header))
}