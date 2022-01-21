package render

import (
	"encoding/json"
	"net/http"
)

const (
	charsetUTF8                 = "charset=UTF-8"
	ContentType                 = "Content-Type"
	ContentTypeText             = "text/plain"
	ContentTypeTextCharsetUTF8  = ContentTypeText + "; " + charsetUTF8
	ContentTypeHTML             = "text/html"
	ContentTypeHTMLCharsetUTF8  = ContentTypeHTML + "; " + charsetUTF8
	ContentTypeXML              = "application/xml"
	ContentTypeXMLCharsetUTF8   = ContentTypeXML + "; " + charsetUTF8
	ContentTypeJSON             = "application/json"
	ContentTypeJSONCharsetUTF8  = ContentTypeJSON + "; " + charsetUTF8
	ContentTypeJSONP            = "application/javascript"
	ContentTypeJSONPCharsetUTF8 = ContentTypeJSONP + "; " + charsetUTF8
)

type Renderer interface {
	Render(http.ResponseWriter) error
}

func writeContentType(w http.ResponseWriter, v string) {
	header := w.Header()
	if header.Get(ContentType) == "" {
		header.Set(ContentType, v)
	}
}

func Blob(w http.ResponseWriter, code int, contentType string, data []byte) (err error) {
	writeContentType(w, contentType)
	if code > 0 {
		w.WriteHeader(code)
	}

	if len(data) > 0 {
		_, err = w.Write(data)
	}
	return
}

func Text(w http.ResponseWriter, code int, data string) error {
	return Blob(w, code, ContentTypeTextCharsetUTF8, []byte(data))
}

func HTML(w http.ResponseWriter, code int, data string) error {
	return Blob(w, code, ContentTypeHTMLCharsetUTF8, []byte(data))
}

func JSON(w http.ResponseWriter, code int, data interface{}) error {
	writeContentType(w, ContentTypeJSONCharsetUTF8)
	if code > 0 {
		w.WriteHeader(code)
	}
	return json.NewEncoder(w).Encode(data)
}

func JSONP(w http.ResponseWriter, code int, callback string, data interface{}) (err error) {
	writeContentType(w, ContentTypeJSONPCharsetUTF8)
	if code > 0 {
		w.WriteHeader(code)
	}

	if _, err = w.Write([]byte(callback + "(")); err != nil {
		return err
	}
	if err = json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	_, err = w.Write([]byte(");"))
	return err
}

func XML(w http.ResponseWriter, code int, data interface{}) error {
	writeContentType(w, ContentTypeJSONCharsetUTF8)
	if code > 0 {
		w.WriteHeader(code)
	}
	return json.NewEncoder(w).Encode(data)
}
