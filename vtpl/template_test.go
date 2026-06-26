package vtpl_test

import (
	"context"
	"errors"
	"html/template"
	"io"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vtpl"
)

func BenchmarkRenderWithTextEngineFacade(b *testing.B) {
	engine := vtpl.NewTextEngine(vtpl.WithEngineFuncMap(map[string]any{"upper": strings.ToUpper}))

	b.ReportAllocs()
	for b.Loop() {
		out, err := vtpl.RenderWithEngine(
			context.Background(),
			engine,
			"hello {{upper .Name}}",
			map[string]string{"Name": "template"},
		)
		if err != nil {
			b.Fatal(err)
		}
		if out != "hello TEMPLATE" {
			b.Fatalf("RenderWithEngine = %q", out)
		}
	}
}

func TestRenderTemplateFacade(t *testing.T) {
	out, err := vtpl.RenderTemplate("hi {{.Name}}", map[string]string{"Name": "gokit"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "hi gokit" {
		t.Fatalf("RenderTemplate() = %q", out)
	}
}

func TestRenderFacadeAndOptions(t *testing.T) {
	out, err := vtpl.Render("hello {{.Name}}", map[string]string{"Name": "tpl"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if out != "hello tpl" {
		t.Fatalf("Render = %q", out)
	}

	out, err = vtpl.RenderWithOptions(
		"hello [[upper .Name]]",
		map[string]string{"Name": "tpl"},
		vtpl.WithTemplateName("custom"),
		vtpl.WithDelims("[[", "]]"),
		vtpl.WithFuncMap(template.FuncMap{"upper": strings.ToUpper}),
	)
	if err != nil {
		t.Fatalf("RenderWithOptions: %v", err)
	}
	if out != "hello TPL" {
		t.Fatalf("RenderWithOptions = %q", out)
	}
}

func TestRenderWithProviderOptions(t *testing.T) {
	factoryCalled := false
	parserCalled := false
	executorCalled := false

	out, err := vtpl.RenderWithOptions(
		"ignored",
		map[string]string{"Name": "provider"},
		vtpl.WithTemplateFactory(func(name string) *template.Template {
			factoryCalled = name == "provider-name"
			return template.New(name)
		}),
		vtpl.WithTemplateName("provider-name"),
		vtpl.WithTemplateParser(func(t *template.Template, source string) (*template.Template, error) {
			parserCalled = source == "ignored"
			return t.Parse("provider {{.Name}}")
		}),
		vtpl.WithTemplateExecutor(func(t *template.Template, w io.Writer, data any) error {
			executorCalled = true
			return t.Execute(w, data)
		}),
	)
	if err != nil {
		t.Fatalf("RenderWithOptions providers: %v", err)
	}
	if out != "provider provider" {
		t.Fatalf("RenderWithOptions providers = %q", out)
	}
	if !factoryCalled || !parserCalled || !executorCalled {
		t.Fatalf("provider calls factory=%v parser=%v executor=%v", factoryCalled, parserCalled, executorCalled)
	}
}

func TestRenderWithOptionsPropagatesProviderErrors(t *testing.T) {
	if _, err := vtpl.RenderWithOptions("bad", nil, vtpl.WithTemplateParser(func(*template.Template, string) (*template.Template, error) {
		return nil, errors.New("parse failed")
	})); err == nil {
		t.Fatal("RenderWithOptions parser error = nil")
	}

	if _, err := vtpl.RenderWithOptions("ok", nil, vtpl.WithTemplateExecutor(func(*template.Template, io.Writer, any) error {
		return errors.New("execute failed")
	})); err == nil {
		t.Fatal("RenderWithOptions executor error = nil")
	}
}

func TestRenderWithEngineFacade(t *testing.T) {
	engine := vtpl.EngineFunc(func(ctx context.Context, req vtpl.RenderRequest) (string, error) {
		if ctx == nil || req.Source != "{{.Name}}" {
			return "", errors.New("unexpected request")
		}
		return "hello " + req.Data.(map[string]string)["Name"], nil
	})

	out, err := vtpl.RenderWithEngine(context.Background(), engine, "{{.Name}}", map[string]string{"Name": "adapter"})
	if err != nil {
		t.Fatalf("RenderWithEngine: %v", err)
	}
	if out != "hello adapter" {
		t.Fatalf("RenderWithEngine = %q", out)
	}
}

func TestBuiltInEngineFacadeContracts(t *testing.T) {
	htmlOut, err := vtpl.RenderWithEngine(context.Background(), vtpl.NewHTMLEngine(), "{{.}}", "<tag>")
	if err != nil {
		t.Fatalf("RenderWithEngine html: %v", err)
	}
	if htmlOut != "&lt;tag&gt;" {
		t.Fatalf("html engine = %q", htmlOut)
	}

	textOut, err := vtpl.RenderWithEngine(context.Background(), vtpl.NewTextEngine(), "{{.}}", "<tag>")
	if err != nil {
		t.Fatalf("RenderWithEngine text: %v", err)
	}
	if textOut != "<tag>" {
		t.Fatalf("text engine = %q", textOut)
	}

	custom := vtpl.NewTextEngine(
		vtpl.WithEngineTemplateName("custom"),
		vtpl.WithEngineDelims("[[", "]]"),
		vtpl.WithEngineFuncMap(map[string]any{"upper": strings.ToUpper}),
	)
	customOut, err := vtpl.RenderWithEngine(context.Background(), custom, "[[upper .Name]]", map[string]string{"Name": "facade"})
	if err != nil {
		t.Fatalf("RenderWithEngine custom: %v", err)
	}
	if customOut != "FACADE" {
		t.Fatalf("custom engine = %q", customOut)
	}
}

func TestRenderWithEngineFacadeErrorContracts(t *testing.T) {
	_, err := vtpl.RenderWithEngine(context.Background(), nil, "{{.}}", "x")
	if !errors.Is(err, vtpl.ErrMissingEngine) {
		t.Fatalf("missing engine error = %v", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("missing engine code = %v", err)
	}

	_, err = vtpl.RenderWithEngine(context.Background(), vtpl.NewTextEngine(), "", nil)
	if !errors.Is(err, vtpl.ErrInvalidRenderRequest) {
		t.Fatalf("empty source error = %v", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("empty source code = %v", err)
	}
}
