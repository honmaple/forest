package template

import (
	"errors"
	"io"
	"io/fs"

	"github.com/flosch/pongo2/v5"
	"github.com/honmaple/forest"
)

type Pongo2Template struct {
	templateSet *pongo2.TemplateSet
}

func (s *Pongo2Template) Render(w io.Writer, name string, data interface{}, c forest.Context) error {
	template, err := s.templateSet.FromCache(name)
	if err != nil {
		return errors.New("template not found")
	}
	if m, ok := data.(pongo2.Context); ok {
		return template.ExecuteWriter(m, w)
	}
	return template.ExecuteWriter(nil, w)
}

func NewPongo2(f fs.FS, patterns ...string) *Pongo2Template {
	return &Pongo2Template{
		templateSet: pongo2.NewSet("app", NewEmbedLoader(f)),
	}
}
