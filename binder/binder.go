package binder

import (
	"encoding/json"
	"encoding/xml"
	"github.com/honmaple/forest/render"
	"net/http"
	"strings"
)

const defaultMemory = 32 << 20

type Binder interface {
	Bind(*http.Request, interface{}) error
}

var (
	XML           = XMLBinder{}
	JSON          = JSONBinder{}
	Form          = FormBinder{"form"}
	Query         = QueryBinder{"query"}
	Header        = HeaderBinder{"header"}
	MultipartForm = MultipartFormBinder{"form"}
)

type JSONBinder struct{}

func (b JSONBinder) Bind(req *http.Request, dst interface{}) error {
	return json.NewDecoder(req.Body).Decode(dst)
}

type XMLBinder struct{}

func (b XMLBinder) Bind(req *http.Request, dst interface{}) error {
	return xml.NewDecoder(req.Body).Decode(dst)
}

type QueryBinder struct {
	TagName string
}

func (b QueryBinder) Bind(req *http.Request, dst interface{}) error {
	return bindData(dst, req.URL.Query(), b.TagName)
}

type FormBinder struct {
	TagName string
}

func (b FormBinder) Bind(req *http.Request, dst interface{}) (err error) {
	if err = req.ParseForm(); err != nil {
		return err
	}
	return bindData(dst, req.Form, b.TagName)
}

type MultipartFormBinder struct {
	TagName string
}

func (b MultipartFormBinder) Bind(req *http.Request, dst interface{}) (err error) {
	if err = req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	return bindData(dst, req.PostForm, b.TagName)
}

type HeaderBinder struct {
	TagName string
}

func (b HeaderBinder) Bind(req *http.Request, dst interface{}) error {
	return bindData(dst, req.Header, b.TagName)
}

func Bind(req *http.Request, dst interface{}) (err error) {
	method := req.Method
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
		return Query.Bind(req, dst)
	}
	ctype := req.Header.Get("Content-Type")
	if strings.Contains(ctype, "/x-www-form-urlencoded") {
		return Form.Bind(req, dst)
	}
	if strings.HasPrefix(ctype, render.MIMEApplicationJSON) {
		return JSON.Bind(req, dst)
	}
	if strings.HasPrefix(ctype, render.MIMEApplicationXML) {
		return XML.Bind(req, dst)
	}
	return nil
}
