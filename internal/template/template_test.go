package template

import (
	"html/template"
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
