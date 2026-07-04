package template

import (
	"context"
	"errors"
	"html/template"
	"io"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func BenchmarkRenderWithTextEngine(b *testing.B) {
	engine := NewTextEngine(WithEngineFuncMap(map[string]any{"upper": strings.ToUpper}))
	req := RenderRequest{
		Source: "hello {{upper .Name}}",
		Data:   map[string]string{"Name": "template"},
	}

	b.ReportAllocs()
	for b.Loop() {
		out, err := engine.Render(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
		if out != "hello TEMPLATE" {
			b.Fatalf("Render = %q", out)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	out, err := RenderTemplate("hello {{.Name}}", map[string]string{"Name": "gokit"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello gokit" {
		t.Fatalf("RenderTemplate() = %q", out)
	}
}

func TestRenderWithTemplateParser(t *testing.T) {
	called := false
	out, err := RenderWithOptions("ignored", map[string]string{"Name": "gokit"}, WithTemplateParser(func(t *template.Template, _ string) (*template.Template, error) {
		called = true
		return t.Parse("hi {{.Name}}")
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !called || out != "hi gokit" {
		t.Fatalf("parser called=%v out=%q", called, out)
	}
}

func TestRenderWithOptionsUsesNameFactoryFuncMapAndDelims(t *testing.T) {
	factoryName := ""
	parserName := ""
	out, err := RenderWithOptions("Hello <<upper .Name>>",
		map[string]string{"Name": "gokit"},
		WithTemplateName("custom-template"),
		WithTemplateFactory(func(name string) *template.Template {
			factoryName = name
			return template.New(name)
		}),
		WithFuncMap(template.FuncMap{"upper": strings.ToUpper}),
		WithDelims("<<", ">>"),
		WithTemplateParser(func(t *template.Template, tpl string) (*template.Template, error) {
			parserName = t.Name()
			return t.Parse(tpl)
		}),
	)
	if err != nil {
		t.Fatalf("RenderWithOptions: %v", err)
	}
	if out != "Hello GOKIT" || factoryName != "custom-template" || parserName != "custom-template" {
		t.Fatalf("out=%q factory=%q parser=%q", out, factoryName, parserName)
	}
}

func TestRenderWithOptionsExecutorAndFallbacks(t *testing.T) {
	executed := false
	out, err := RenderWithOptions("ignored", map[string]string{"Name": "gokit"},
		WithTemplateName(""),
		WithTemplateFactory(nil),
		WithTemplateParser(nil),
		WithTemplateExecutor(func(tpl *template.Template, w io.Writer, data any) error {
			executed = true
			if tpl.Name() != "knifer-go-template" {
				t.Fatalf("template name = %q", tpl.Name())
			}
			_, err := io.WriteString(w, "executor-output")
			return err
		}),
	)
	if err != nil {
		t.Fatalf("RenderWithOptions executor: %v", err)
	}
	if !executed || out != "executor-output" {
		t.Fatalf("executed=%v out=%q", executed, out)
	}
}

func TestNilRenderProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	parser := func(tpl *template.Template, source string) (*template.Template, error) {
		return tpl.Parse(source)
	}
	executor := func(tpl *template.Template, w io.Writer, data any) error {
		return tpl.Execute(w, data)
	}
	cfg := applyRenderOptions([]RenderOption{
		WithTemplateFactory(template.New),
		WithTemplateFactory(nil),
		WithTemplateParser(parser),
		WithTemplateParser(nil),
		WithTemplateExecutor(executor),
		WithTemplateExecutor(nil),
	})
	if cfg.factory == nil || cfg.parser == nil || cfg.executor == nil {
		t.Fatalf("nil render provider option overwrote configured provider: %#v", cfg)
	}
}

func TestRenderWithOptionsReturnsParserAndExecutorErrors(t *testing.T) {
	parserErr := errors.New("parse failed")
	if _, err := RenderWithOptions("ignored", nil, WithTemplateParser(func(*template.Template, string) (*template.Template, error) {
		return nil, parserErr
	})); !errors.Is(err, parserErr) {
		t.Fatalf("parser error = %v", err)
	}

	executorErr := errors.New("execute failed")
	if _, err := RenderWithOptions("ok", nil, WithTemplateExecutor(func(*template.Template, io.Writer, any) error {
		return executorErr
	})); !errors.Is(err, executorErr) {
		t.Fatalf("executor error = %v", err)
	}
}

func TestRenderWithEngineUsesInjectedEngine(t *testing.T) {
	called := false
	out, err := RenderWithEngine(context.Background(), EngineFunc(func(ctx context.Context, req RenderRequest) (string, error) {
		called = ctx != nil && req.Source == "{{.Name}}" && req.Data.(map[string]string)["Name"] == "adapter"
		return "engine-output", nil
	}), "{{.Name}}", map[string]string{"Name": "adapter"})
	if err != nil {
		t.Fatalf("RenderWithEngine: %v", err)
	}
	if !called || out != "engine-output" {
		t.Fatalf("called=%v out=%q", called, out)
	}
}

func TestRenderWithEngineRejectsMissingEngine(t *testing.T) {
	_, err := RenderWithEngine(context.Background(), nil, "{{.}}", "x")
	if !errors.Is(err, ErrMissingEngine) {
		t.Fatalf("RenderWithEngine missing engine error = %v", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("RenderWithEngine missing engine code = %v", err)
	}
}

func TestBuiltInEngines(t *testing.T) {
	tests := []struct {
		name   string
		engine Engine
		want   string
	}{
		{
			name:   "html engine escapes html",
			engine: NewHTMLEngine(),
			want:   "&lt;tag&gt;",
		},
		{
			name:   "text engine keeps text raw",
			engine: NewTextEngine(),
			want:   "<tag>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderWithEngine(context.Background(), tt.engine, "{{.}}", "<tag>")
			if err != nil {
				t.Fatalf("RenderWithEngine: %v", err)
			}
			if got != tt.want {
				t.Fatalf("RenderWithEngine = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuiltInEngineOptionsAndValidation(t *testing.T) {
	engine := NewTextEngine(
		WithEngineTemplateName("custom"),
		WithEngineDelims("[[", "]]"),
		WithEngineFuncMap(map[string]any{"upper": strings.ToUpper}),
	)
	out, err := RenderWithEngine(context.Background(), engine, "[[upper .Name]]", map[string]string{"Name": "tpl"})
	if err != nil {
		t.Fatalf("RenderWithEngine options: %v", err)
	}
	if out != "TPL" {
		t.Fatalf("RenderWithEngine options = %q", out)
	}

	_, err = RenderWithEngine(context.Background(), engine, "", nil)
	if !errors.Is(err, ErrInvalidRenderRequest) {
		t.Fatalf("empty source error = %v", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("empty source code = %v", err)
	}

	var nilCtx context.Context
	_, err = RenderWithEngine(nilCtx, engine, "{{.}}", "x")
	if !errors.Is(err, ErrInvalidRenderRequest) {
		t.Fatalf("nil context error = %v", err)
	}
}

func TestBuiltInEngineErrorBranches(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	for _, tt := range []struct {
		name   string
		engine Engine
	}{
		{name: "html", engine: NewHTMLEngine(WithEngineTemplateName(""))},
		{name: "text", engine: NewTextEngine(WithEngineTemplateName(""))},
	} {
		t.Run(tt.name+" canceled context", func(t *testing.T) {
			_, err := tt.engine.Render(canceled, RenderRequest{Source: "{{.}}", Data: "x"})
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled error = %v", err)
			}
		})

		t.Run(tt.name+" parse error", func(t *testing.T) {
			_, err := tt.engine.Render(context.Background(), RenderRequest{Source: "{{", Data: nil})
			if err == nil {
				t.Fatal("parse error = nil")
			}
		})

		t.Run(tt.name+" execute error", func(t *testing.T) {
			_, err := tt.engine.Render(context.Background(), RenderRequest{Source: "{{call .}}", Data: "not-a-function"})
			if err == nil {
				t.Fatal("execute error = nil")
			}
		})
	}
}

func TestHTMLEngineOptions(t *testing.T) {
	engine := NewHTMLEngine(
		WithEngineDelims("[[", "]]"),
		WithEngineFuncMap(map[string]any{"upper": strings.ToUpper}),
	)

	out, err := engine.Render(context.Background(), RenderRequest{
		Source: "[[upper .]]",
		Data:   "html",
	})
	if err != nil {
		t.Fatalf("html engine options: %v", err)
	}
	if out != "HTML" {
		t.Fatalf("html engine options = %q", out)
	}
}
