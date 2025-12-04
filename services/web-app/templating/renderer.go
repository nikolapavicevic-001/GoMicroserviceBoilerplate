package templating

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	Pages    map[string]*template.Template
	Partials map[string]*template.Template
}

func NewRenderer(templates *LoadedTemplates) *TemplateRenderer {
	return &TemplateRenderer{
		Pages:    templates.Pages,
		Partials: templates.Partials,
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Partial?
	if strings.HasPrefix(name, "partials/") {
		tmpl, ok := t.Partials[name]
		if !ok {
			return fmt.Errorf("partial %s not found", name)
		}
		return tmpl.ExecuteTemplate(w, filepath.Base(name), data)
	}

	// Page?
	tmpl, ok := t.Pages[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}
