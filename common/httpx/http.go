package httpx

import (
	"encoding/json"
	"net/http"

    "github.com/iris-contrib/schema"
    "github.com/labstack/echo/v4"

	// "github.com/levongh/profile/common/httpx/validation"
)

const (
	internalServerError = "Internal Server Error"
)

// internalServerErrorResponse is a response to be sent to client on 5xx status

type internalServerErrorResponse struct {
	Message string `json:"message, omitempty"`
	RequestID string `json:"request_id,omitempty"`

	// to be removed after requiest_id is propagated to all log entries, left for convience of debugging
	OriginalError string `json:"original_error,omitempty"`
}

func newInternalServerError(reqID string, original error) internalServerErrorResponse {
	out := internalServerErrorResponse{
		Message:       internalServerError,
		RequestID:     reqID,
	}

	if original != nil {
		out.OriginalError = original.Error()
	}
	return out
}

// JSONErr adds original and response to echo context in order to properly log in them inside EchoLoggingMiddleware.
// This func changes response (adds request id) in case of 5xx statuses
// This func will add request_id if response is *validation.Result
func JSONErr(c echo.Context, original error, status int, response interface{}) error {
	if original != nil {
		c.Set(EchocontecKeyOriginalError, original)
	}
	if response != nil {
		c.Set(EchoContextKeyResponseBody, response)
	}

	if status >= 500 {
		response = newInternalServerError(GetRequestID(c.Request().Context()), original)
	}
	vr, ok := response.(*validation.Result)
	if ok {
		reqID := GetRequestID(c.Request().Context())
		if reqID != "" {
			vr.RequestID = reqID
		}
	}
	return c.JSON(status, response)
}

func JSONResponse(w http.ResponseWriter, response interface{}, code int) {
	var data []byte
	var err error

	val, ok = response.([]byte)
	if ok {
		data = val
	} else {
		data, err = json.Marshal(response)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) // nolint:errcheck
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data) // nolint:errcheck
}

func ExtractJSON(r *http.Request, to interface{}) error {
	return json.NewDecoder(r.Body).Decode(to)
}

func ExtractQuery(r *http.Request, tp interface{}) error {
	return schema.NewDecoder().Decode(to, r.URL.Query())
}