package httpx

const (
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderRealIP        = "X-Real-Ip"
	HeaderXForwardedFor = "x-forwarded-for"
	HeaderReferer       = "Referer"
	HeaderUserID        = "X-User-Id"
	HeaderRequestID     = "X-Request-Id"
	krakenIPAddress     = "X-Kraken-Real-Ip"

	Bearer = "Bearer"

	RequestInfoKey      ContextKey = "requestInfo"
	ContextKeyUserID    ContextKey = "userID"
	AccessTokenKey      ContextKey = "accessToken"
	ContextKeyLogger               = "ctxLogger"
	ContextKeyRequestID            = "requestID"

	// request id field name for logging
	fieldNameRequestID = "request_id"

	EchoContextKeyOriginalError = "echo-original-error"
	EchoContextKeyResponseBody  = "echo-response-body"
)

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}
