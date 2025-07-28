package tracing

import (
    "bufio"
    "io"
    "net"
    "net/http"
)

// responseWriter wraps http.ResponseWriter.
type responseWriter struct {
    io.Writer
    http.ResponseWriter
    Status int
}

func (w *responseWriter) WriteHeader(code int) {
    w.ResponseWriter.WriteHeader(code)
    w.Status = code
}

func (w *responseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func (w *responseWriter) Flush() {
    w.ResponseWriter.(http.Flusher).Flush()
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    return w.ResponseWriter.(http.Hijacker).Hijack()
}