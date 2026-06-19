package vbean_test

import (
	"errors"
	"slices"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vbean"
)

type userDTO struct {
	Name string `bean:"name,alias=full_name"`
	Age  string `bean:"age"`
}

func assertEqualStrings(t *testing.T, want, got []string) {
	t.Helper()
	if !slices.Equal(want, got) {
		t.Fatalf("strings = %#v, want %#v", got, want)
	}
}

type userModel struct {
	Name string `json:"full_name"`
	Age  int    `json:"age"`
}

func assertFacadeBeanCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
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
	var beanErr *vbean.Error
	if !errors.As(err, &beanErr) {
		t.Fatalf("errors.As(err, *vbean.Error) = false: %v", err)
	}
}
