package vtpl

import (
	"html/template"
	"io"

	templateimpl "github.com/imajinyun/go-knifer/internal/template"
)

// RenderOption customizes template rendering per call.
type RenderOption = templateimpl.RenderOption

// WithTemplateName sets the template name used while parsing.
func WithTemplateName(name string) RenderOption { return templateimpl.WithTemplateName(name) }

// WithFuncMap sets functions available to the template.
func WithFuncMap(funcMap template.FuncMap) RenderOption { return templateimpl.WithFuncMap(funcMap) }

// WithDelims sets template action delimiters.
func WithDelims(left, right string) RenderOption { return templateimpl.WithDelims(left, right) }

// WithTemplateFactory sets the template constructor used before parsing.
func WithTemplateFactory(factory func(string) *template.Template) RenderOption {
	return templateimpl.WithTemplateFactory(factory)
}

// WithTemplateParser sets the parser used after template construction.
func WithTemplateParser(parser func(*template.Template, string) (*template.Template, error)) RenderOption {
	return templateimpl.WithTemplateParser(parser)
}

// WithTemplateExecutor sets the executor used after parsing.
func WithTemplateExecutor(executor func(*template.Template, io.Writer, any) error) RenderOption {
	return templateimpl.WithTemplateExecutor(executor)
}

// Render renders a Go html/template string with data.
func Render(tpl string, data any) (string, error) { return templateimpl.Render(tpl, data) }

// RenderWithOptions renders a Go html/template string with per-call options.
func RenderWithOptions(tpl string, data any, opts ...RenderOption) (string, error) {
	return templateimpl.RenderWithOptions(tpl, data, opts...)
}

// RenderTemplate renders a Go html/template string with data.
func RenderTemplate(tpl string, data any) (string, error) {
	return templateimpl.RenderTemplate(tpl, data)
}
