package tracing

import (
    "github.com/opentracing/opentracing-go"
)

const (
    defaultComponentName = "http/middleware"
)

var (
    // DefaultTraceConfig is the default Trace middleware config.
    DefaultTraceConfig = TraceConfig{
        ComponentName: defaultComponentName,
        IsBodyDump:    false,
    }
)

// TraceConfig defines the config for Trace middleware.
type TraceConfig struct {
    // OpenTracing Tracer instance which should be got before
    Tracer opentracing.Tracer

    // ComponentName used for describing the tracing component name
    ComponentName string

    // Add req body & resp body to tracing tags
    IsBodyDump bool
}