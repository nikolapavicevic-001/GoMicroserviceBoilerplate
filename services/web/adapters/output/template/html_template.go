package template

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/microserviceboilerplate/web/domain/port/output"
)

// HTMLTemplateRenderer implements the TemplateRepository interface
type HTMLTemplateRenderer struct {
	templates *template.Template
}

// NewHTMLTemplateRenderer creates a new HTML template renderer
func NewHTMLTemplateRenderer(templatesPath string) (output.TemplateRepository, error) {
	funcMap := template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"upper": strings.ToUpper,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &HTMLTemplateRenderer{
		templates: tmpl,
	}, nil
}

// Render renders a template with the given data
func (r *HTMLTemplateRenderer) Render(ctx context.Context, templateName string, data interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	return &buf, nil
}

// RenderString renders a template and returns it as a string
func (r *HTMLTemplateRenderer) RenderString(ctx context.Context, templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	return buf.String(), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

