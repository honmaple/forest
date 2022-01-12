package forest

import (
	"net/http"
)

type Response struct {
	http.ResponseWriter
	Size   int
	Status int
}

const noWritten = -1

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
