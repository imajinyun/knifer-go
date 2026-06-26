package bean

import (
	"errors"
	"slices"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

type embeddedProfile struct {
	Trace string `bean:"trace_id"`
}

func assertEqualStrings(t *testing.T, want, got []string) {
	t.Helper()
	if !slices.Equal(want, got) {
		t.Fatalf("strings = %#v, want %#v", got, want)
	}
}

type sourceProfile struct {
	embeddedProfile
	Name  string `bean:"name,alias=full_name|displayName"`
	Age   string `bean:"age"`
	Admin string `bean:"admin"`
	Skip  string `bean:"-"`
	Empty string `bean:"empty"`
}

type targetProfile struct {
	Name  string `bean:"name,alias=full_name|displayName" json:"full_name"`
	Age   int    `json:"age"`
	Admin bool   `json:"admin"`
	Trace string `json:"trace_id"`
	Empty string `json:"empty"`
}

func assertBeanInvalidInput(t *testing.T, err error) {
	t.Helper()
	const code = knifer.ErrCodeInvalidInput
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
	var beanErr *BeanError
	if !errors.As(err, &beanErr) {
		t.Fatalf("errors.As(err, *BeanError) = false: %v", err)
	}
}
