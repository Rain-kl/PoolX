// Package response provides the Wavelet-aligned HTTP API envelope.
//
// Success and error bodies always use:
//
//	{ "error_msg": "", "data": <T|null> }
//
// HTTP status carries the outcome; the body does not include a separate error code object.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the canonical API response body (matches Wavelet).
type Envelope struct {
	ErrorMsg string `json:"error_msg"`
	Data     any    `json:"data"`
}

// APIError is an abortable business error for handlers that prefer the abort helpers.
type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string { return e.Msg }

// NewError constructs an APIError.
func NewError(code int, msg string) *APIError {
	return &APIError{Code: code, Msg: msg}
}

// OK constructs a success envelope.
func OK(data any) Envelope {
	return Envelope{ErrorMsg: "", Data: data}
}

// OKNil constructs a success envelope with null data.
func OKNil() Envelope {
	return Envelope{ErrorMsg: "", Data: nil}
}

// Err constructs an error envelope (data is always null).
func Err(msg string) Envelope {
	return Envelope{ErrorMsg: msg, Data: nil}
}

// Success writes a success JSON body with the Wavelet envelope.
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, OK(data))
}

// Error writes an error JSON body and aborts.
// code is retained for call-site documentation but is not emitted in the body (Wavelet style).
func Error(c *gin.Context, status int, code, message string) {
	_ = code
	c.AbortWithStatusJSON(status, Err(message))
}

// AbortWithError attaches an APIError and aborts; use with ErrorHandlerMiddleware if registered.
func AbortWithError(c *gin.Context, code int, msg string) {
	_ = c.Error(NewError(code, msg))
	c.Abort()
}

// AbortBadRequest aborts with HTTP 400.
func AbortBadRequest(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusBadRequest, msg)
}

// AbortUnauthorized aborts with HTTP 401.
func AbortUnauthorized(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusUnauthorized, msg)
}

// AbortForbidden aborts with HTTP 403.
func AbortForbidden(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusForbidden, msg)
}

// AbortNotFound aborts with HTTP 404.
func AbortNotFound(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusNotFound, msg)
}

// AbortConflict aborts with HTTP 409.
func AbortConflict(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusConflict, msg)
}

// AbortInternal aborts with HTTP 500.
func AbortInternal(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusInternalServerError, msg)
}

// ErrorHandlerMiddleware formats c.Errors as Wavelet envelopes (optional global exit).
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		err := c.Errors.Last().Err
		if apiErr, ok := err.(*APIError); ok {
			c.JSON(apiErr.Code, Err(apiErr.Msg))
			return
		}
		c.JSON(http.StatusInternalServerError, Err("内部系统错误"))
	}
}
