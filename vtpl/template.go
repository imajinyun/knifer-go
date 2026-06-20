package vtpl

import (
	"context"
	"html/template"
	"io"

	templateimpl "github.com/imajinyun/go-knifer/internal/template"
)

// ErrMissingEngine reports that no template engine was provided.
var ErrMissingEngine = templateimpl.ErrMissingEngine

// ErrInvalidRenderRequest reports an invalid engine render request.
var ErrInvalidRenderRequest = templateimpl.ErrInvalidRenderRequest

// RenderRequest is the engine-neutral input for rendering a template source.
type RenderRequest = templateimpl.RenderRequest

// Engine renders template requests with an adapter-selected implementation.
type Engine = templateimpl.Engine

// EngineFunc adapts a function to Engine.
type EngineFunc = templateimpl.EngineFunc

// EngineOption customizes built-in template engines.
type EngineOption = templateimpl.EngineOption

// WithEngineTemplateName sets the template name used by built-in engines.
func WithEngineTemplateName(name string) EngineOption {
	return templateimpl.WithEngineTemplateName(name)
}

// WithEngineFuncMap sets functions available to built-in engines.
func WithEngineFuncMap(funcMap map[string]any) EngineOption {
	return templateimpl.WithEngineFuncMap(funcMap)
}

// WithEngineDelims sets template action delimiters for built-in engines.
func WithEngineDelims(left, right string) EngineOption {
	return templateimpl.WithEngineDelims(left, right)
}

// NewHTMLEngine creates an Engine backed by html/template.
func NewHTMLEngine(opts ...EngineOption) Engine { return templateimpl.NewHTMLEngine(opts...) }

// NewTextEngine creates an Engine backed by text/template.
func NewTextEngine(opts ...EngineOption) Engine { return templateimpl.NewTextEngine(opts...) }

// RenderWithEngine renders a template source through a caller-selected engine.
func RenderWithEngine(ctx context.Context, engine Engine, source string, data any) (string, error) {
	return templateimpl.RenderWithEngine(ctx, engine, source, data)
}

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
