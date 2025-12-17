package output

import (
	"context"
	"io"
)

// TemplateRepository defines the template rendering interface
type TemplateRepository interface {
	// Render renders a template with the given data
	Render(ctx context.Context, templateName string, data interface{}) (io.Reader, error)
	
	// RenderString renders a template and returns it as a string
	RenderString(ctx context.Context, templateName string, data interface{}) (string, error)
}

