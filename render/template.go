package render

import (
	"html/template"
	"io/fs"
	"net/http"
)

type (
	Renderer interface {
		Render(http.ResponseWriter) error
	}
	TemplateRenderer interface {
		Render(http.ResponseWriter, string, interface{}) error
	}
	Template struct {
		templates *template.Template
	}
)

func (t *Template) Render(w http.ResponseWriter, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate(f fs.FS, patterns ...string) *Template {
	return &Template{template.Must(template.ParseFS(f, patterns...))}
}
