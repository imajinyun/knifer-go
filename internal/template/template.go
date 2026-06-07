package template

import (
	"bytes"
	"html/template"
	"io"
)

type renderConfig struct {
	name     string
	funcMap  template.FuncMap
	delims   [2]string
	factory  func(string) *template.Template
	parser   func(*template.Template, string) (*template.Template, error)
	executor func(*template.Template, io.Writer, any) error
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
