package forest

import (
	"net/http"
)

type Response struct {
	http.ResponseWriter
	Status int
}

func (r *Response) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(r.Status)
}

func (r *Response) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
}

func NewResponse(w http.ResponseWriter, e *Engine) *Response {
	return &Response{ResponseWriter: w}
}
