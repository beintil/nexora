package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

type HTMLFile struct {
	name string
}

var (
	HTMLFileRegister = HTMLFile{
		name: "register.html",
		//map[string]any{
		//	"Link": "",
		//},
	}
)

//go:embed email/*.html
var emailFS embed.FS

type Renderer struct {
	tpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	temlpates, err := template.ParseFS(emailFS, "email/*.html")
	if err != nil {
		return nil, fmt.Errorf("unable to parse templates: %w", err)
	}

	return &Renderer{tpl: temlpates}, nil
}

func (r *Renderer) Render(file HTMLFile, data map[string]any) (bytes.Buffer, error) {
	var buf bytes.Buffer
	err := r.tpl.ExecuteTemplate(&buf, file.name, data)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("unable to render template: %w", err)
	}
	return buf, nil
}
