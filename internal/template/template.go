package template

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io"
	texttemplate "text/template"

	knifer "github.com/imajinyun/go-knifer"
)

// ErrMissingEngine reports that no template engine was provided.
var ErrMissingEngine = errors.New("template: missing engine")

// ErrInvalidRenderRequest reports an invalid engine render request.
var ErrInvalidRenderRequest = errors.New("template: invalid render request")

type renderConfig struct {
	name     string
	funcMap  template.FuncMap
	delims   [2]string
	factory  func(string) *template.Template
	parser   func(*template.Template, string) (*template.Template, error)
	executor func(*template.Template, io.Writer, any) error
}

type engineConfig struct {
	name    string
	funcMap map[string]any
	delims  [2]string
}

// RenderRequest is the engine-neutral input for rendering a template source.
type RenderRequest struct {
	// Source is the template source to parse and render.
	Source string
	// Data is passed to the selected engine during execution.
	Data any
}

// Engine renders template requests with an adapter-selected implementation.
type Engine interface {
	Render(ctx context.Context, req RenderRequest) (string, error)
}

// EngineFunc adapts a function to Engine.
type EngineFunc func(context.Context, RenderRequest) (string, error)

// Render calls f(ctx, req).
func (f EngineFunc) Render(ctx context.Context, req RenderRequest) (string, error) {
	return f(ctx, req)
}

// EngineOption customizes built-in template engines.
type EngineOption func(*engineConfig)

// WithEngineTemplateName sets the template name used by built-in engines.
func WithEngineTemplateName(name string) EngineOption {
	return func(c *engineConfig) { c.name = name }
}

// WithEngineFuncMap sets functions available to built-in engines.
func WithEngineFuncMap(funcMap map[string]any) EngineOption {
	return func(c *engineConfig) { c.funcMap = funcMap }
}

// WithEngineDelims sets template action delimiters for built-in engines.
func WithEngineDelims(left, right string) EngineOption {
	return func(c *engineConfig) { c.delims = [2]string{left, right} }
}

func applyEngineOptions(opts []EngineOption) engineConfig {
	cfg := engineConfig{name: "go-knifer-template"}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.name == "" {
		cfg.name = "go-knifer-template"
	}
	return cfg
}

// NewHTMLEngine creates an Engine backed by html/template.
func NewHTMLEngine(opts ...EngineOption) Engine {
	cfg := applyEngineOptions(opts)
	return EngineFunc(func(ctx context.Context, req RenderRequest) (string, error) {
		if err := validateRenderRequest(ctx, req); err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}

		t := template.New(cfg.name)
		if cfg.funcMap != nil {
			t = t.Funcs(template.FuncMap(cfg.funcMap))
		}
		if cfg.delims[0] != "" || cfg.delims[1] != "" {
			t = t.Delims(cfg.delims[0], cfg.delims[1])
		}
		parsed, err := t.Parse(req.Source)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := parsed.Execute(&buf, req.Data); err != nil {
			return "", err
		}
		return buf.String(), nil
	})
}

// NewTextEngine creates an Engine backed by text/template.
func NewTextEngine(opts ...EngineOption) Engine {
	cfg := applyEngineOptions(opts)
	return EngineFunc(func(ctx context.Context, req RenderRequest) (string, error) {
		if err := validateRenderRequest(ctx, req); err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}

		t := texttemplate.New(cfg.name)
		if cfg.funcMap != nil {
			t = t.Funcs(texttemplate.FuncMap(cfg.funcMap))
		}
		if cfg.delims[0] != "" || cfg.delims[1] != "" {
			t = t.Delims(cfg.delims[0], cfg.delims[1])
		}
		parsed, err := t.Parse(req.Source)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := parsed.Execute(&buf, req.Data); err != nil {
			return "", err
		}
		return buf.String(), nil
	})
}

// RenderWithEngine renders a template source through a caller-selected engine.
func RenderWithEngine(ctx context.Context, engine Engine, source string, data any) (string, error) {
	if engine == nil {
		return "", knifer.WrapError(
			knifer.ErrCodeInvalidInput,
			"template: missing engine",
			ErrMissingEngine,
		)
	}
	return engine.Render(ctx, RenderRequest{Source: source, Data: data})
}

func validateRenderRequest(ctx context.Context, req RenderRequest) error {
	if ctx == nil {
		return knifer.WrapError(
			knifer.ErrCodeInvalidInput,
			"template: nil context",
			ErrInvalidRenderRequest,
		)
	}
	if req.Source == "" {
		return knifer.WrapError(
			knifer.ErrCodeInvalidInput,
			"template: empty source",
			ErrInvalidRenderRequest,
		)
	}
	return nil
}

// RenderOption customizes template rendering per call.
type RenderOption func(*renderConfig)

// WithTemplateName sets the template name used while parsing.
func WithTemplateName(name string) RenderOption { return func(c *renderConfig) { c.name = name } }

// WithFuncMap sets functions available to the template.
func WithFuncMap(funcMap template.FuncMap) RenderOption {
	return func(c *renderConfig) { c.funcMap = funcMap }
}

// WithDelims sets template action delimiters.
func WithDelims(left, right string) RenderOption {
	return func(c *renderConfig) { c.delims = [2]string{left, right} }
}

// WithTemplateFactory sets the template constructor used before parsing.
func WithTemplateFactory(factory func(string) *template.Template) RenderOption {
	return func(c *renderConfig) { c.factory = factory }
}

// WithTemplateParser sets the parser used after template construction.
func WithTemplateParser(parser func(*template.Template, string) (*template.Template, error)) RenderOption {
	return func(c *renderConfig) { c.parser = parser }
}

// WithTemplateExecutor sets the executor used after parsing.
func WithTemplateExecutor(executor func(*template.Template, io.Writer, any) error) RenderOption {
	return func(c *renderConfig) { c.executor = executor }
}

func applyRenderOptions(opts []RenderOption) renderConfig {
	cfg := renderConfig{name: "go-knifer-template", factory: template.New}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.name == "" {
		cfg.name = "go-knifer-template"
	}
	if cfg.factory == nil {
		cfg.factory = template.New
	}
	if cfg.parser == nil {
		cfg.parser = func(t *template.Template, tpl string) (*template.Template, error) { return t.Parse(tpl) }
	}
	if cfg.executor == nil {
		cfg.executor = func(t *template.Template, w io.Writer, data any) error { return t.Execute(w, data) }
	}
	return cfg
}

// Render renders a Go html/template string with data.
func Render(tpl string, data any) (string, error) {
	return RenderWithOptions(tpl, data)
}

// RenderWithOptions renders a Go html/template string with per-call options.
func RenderWithOptions(tpl string, data any, opts ...RenderOption) (string, error) {
	cfg := applyRenderOptions(opts)
	t := cfg.factory(cfg.name)
	if cfg.funcMap != nil {
		t = t.Funcs(cfg.funcMap)
	}
	if cfg.delims[0] != "" || cfg.delims[1] != "" {
		t = t.Delims(cfg.delims[0], cfg.delims[1])
	}
	t, err := cfg.parser(t, tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := cfg.executor(t, &buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTemplate renders a Go html/template string with data.
func RenderTemplate(tpl string, data any) (string, error) { return Render(tpl, data) }
