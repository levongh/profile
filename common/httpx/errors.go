package httpx

import (
	"errors"
	"fmt"
)

var (
	ErrUserIDIsMissing    = errors.New("x-user-id is missing from header")
	ErrUnauthorized       = errors.New("authentication failed")
	ErrUserAgentIsMissing = errors.New("user-agent is missing from header")
)

var (
	ErrMsgFailedToCreateRequest = "failed to create request: %w"
)

// ErrHTTPRequest is returned if Request.Send was not able to send request
type ErrHTTPRequest struct {
	Err error
}

func (e ErrHTTPRequest) Error() string {
	return fmt.Sprintf("http request failure %s", e.Err.Error())
}

// ErrHTTPResponse is returned if Request.Send sent http request but response code is not 200
type ErrHTTPResponse struct {
	URL  string
	Code int
	Body []byte
}

func NewErrHTTPResponse(url string, code int, body []byte) ErrHTTPResponse {
	return ErrHTTPResponse{
		URL:  url,
		Code: code,
		Body: body,
	}
}

func (e ErrHTTPResponse) Error() string {
	return fmt.Sprintf("url='%s' status='%d' response='%s'", e.URL, e.Code, e.Body)
}
