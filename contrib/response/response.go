package response

import (
	"net/http"

	"github.com/honmaple/forest"
)

type PageInfo struct {
	Page     int   `json:"page"  query:"page"`
	Limit    int   `json:"limit" query:"limit"`
	Total    int64 `json:"total"`
	NotLimit bool  `json:"-"`
}

func (s *PageInfo) GetLimit() (int, int) {
	if s.Page < 1 {
		s.Page = 1
	}
	if s.Limit < 1 {
		s.Limit = 10
	}
	offset := (s.Page - 1) * s.Limit
	if offset < 0 {
		offset = 0
	}
	return offset, s.Limit
}

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}

type ListResponse struct {
	PageInfo
	List interface{} `json:"list,omitempty"`
}

func New(code int, data ...interface{}) *Response {
	resp := &Response{
		Code: code,
	}

	if len(data) == 0 {
		resp.Message = http.StatusText(code)
	} else {
		resp.Data = data
	}
	return resp
}

func render(c forest.Context, code int, data ...interface{}) error {
	return c.JSON(code, New(code, data...))
}

func OK(c forest.Context, message string, data interface{}) error {
	return render(c, http.StatusOK, message, data)
}

func BadRequest(c forest.Context, data ...interface{}) error {
	return render(c, http.StatusBadRequest, data)
}

func UnAuthorized(c forest.Context, message string, data interface{}) error {
	return render(c, http.StatusUnauthorized, message, data)
}

func Forbidden(c forest.Context, message string, data interface{}) error {
	return render(c, http.StatusForbidden, message, data)
}

func NotFound(c forest.Context, message string, data interface{}) error {
	return render(c, http.StatusNotFound, message, data)
}

func ServerError(c forest.Context, message string, data interface{}) error {
	return render(c, http.StatusInternalServerError, message, data)
}
