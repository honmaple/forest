package render

import (
	"encoding/json"
	"net/http"
)

const (
	charsetUTF8                          = "charset=UTF-8"
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
)

type Renderer interface {
	Render(http.ResponseWriter) error
}

func Blob(w http.ResponseWriter, contentType string, data []byte) (err error) {
	writeContentType(w, contentType)

	if len(data) > 0 {
		_, err = w.Write(data)
	}
	return
}

func Text(w http.ResponseWriter, data string) error {
	return Blob(w, MIMETextPlainCharsetUTF8, []byte(data))
}

func HTML(w http.ResponseWriter, data string) error {
	return Blob(w, MIMETextHTMLCharsetUTF8, []byte(data))
}

func JSON(w http.ResponseWriter, data interface{}) error {
	writeContentType(w, MIMEApplicationJSONCharsetUTF8)
	return json.NewEncoder(w).Encode(data)
}

func JSONP(w http.ResponseWriter, callback string, data interface{}) (err error) {
	writeContentType(w, MIMEApplicationJavaScriptCharsetUTF8)
	if _, err = w.Write([]byte(callback + "(")); err != nil {
		return err
	}
	if err = json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	_, err = w.Write([]byte(");"))
	return err
}

func XML(w http.ResponseWriter, data interface{}) error {
	writeContentType(w, MIMEApplicationJSONCharsetUTF8)
	return json.NewEncoder(w).Encode(data)
}

func writeContentType(w http.ResponseWriter, v string) {
	header := w.Header()
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", v)
	}
}
