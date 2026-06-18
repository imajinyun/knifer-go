package template

import (
	"errors"
	"html/template"
	"io"
	"strings"
	"testing"
)

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
			if tpl.Name() != "go-knifer-template" {
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
