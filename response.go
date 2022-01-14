package forest

import (
	"fmt"
	"net/http"
)

type (
	Error struct {
		Code    int         `json:"-"`
		Message interface{} `json:"message"`
	}
	Response struct {
		http.ResponseWriter
		Size   int
		Status int
	}
)

const noWritten = -1

func (e *Error) Error() string {
	return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
}

func NewError(code int, message ...interface{}) *Error {
	e := &Error{
		Code: code,
	}
	if len(message) > 0 {
		e.Message = message[0]
	} else {
		e.Message = http.StatusText(code)
	}
	return e
}

func (r *Response) Written() bool {
	return r.Size != noWritten
}

func (r *Response) WriteHeader(code int) {
	if r.Written() {
		return
	}
	r.Size = 0
	r.Status = code
	r.ResponseWriter.WriteHeader(r.Status)
}

func (r *Response) Write(b []byte) (n int, err error) {
	if !r.Written() {
		if r.Status == 0 {
			r.Status = http.StatusOK
		}
		r.WriteHeader(r.Status)
	}
	n, err = r.ResponseWriter.Write(b)
	r.Size += n
	return
}

func (r *Response) reset(w http.ResponseWriter) {
	r.Size = noWritten
	r.Status = http.StatusOK
	r.ResponseWriter = w
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{ResponseWriter: w}
}
