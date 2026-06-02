package vbean_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vbean"
)

type userDTO struct {
	Name string `bean:"name,alias=full_name"`
	Age  string `bean:"age"`
}

type userModel struct {
	Name string `json:"full_name"`
	Age  int    `json:"age"`
}

func TestFacadeCopyProperties(t *testing.T) {
	var dst userModel
	if err := vbean.CopyProperties(userDTO{Name: "alice", Age: "18"}, &dst); err != nil {
		t.Fatalf("CopyProperties() error = %v", err)
	}
	if dst.Name != "alice" || dst.Age != 18 {
		t.Fatalf("dst = %+v", dst)
	}
}

func TestFacadeToMap(t *testing.T) {
	got, err := vbean.ToMap(userDTO{Name: "bob", Age: "20"})
	if err != nil {
		t.Fatalf("ToMap() error = %v", err)
	}
	if got["name"] != "bob" || got["age"] != "20" {
		t.Fatalf("map = %#v", got)
	}
}
