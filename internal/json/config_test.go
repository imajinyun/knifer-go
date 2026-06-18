package json

import (
	"testing"
)

func TestCreateConfig(t *testing.T) {
	cfg := CreateConfig()
	if cfg == nil {
		t.Fatal("CreateConfig() = nil")
	}
	if cfg.IndentFactor != 4 {
		t.Fatalf("default IndentFactor = %d", cfg.IndentFactor)
	}
}

func TestDefaultParseBool(t *testing.T) {
	if ok, _ := defaultParseBool("true"); !ok {
		t.Fatal("defaultParseBool('true') = false")
	}
	if ok, _ := defaultParseBool("1"); !ok {
		t.Fatal("defaultParseBool('1') = false")
	}
	if ok, _ := defaultParseBool("yes"); !ok {
		t.Fatal("defaultParseBool('yes') = false")
	}
	if ok, _ := defaultParseBool("false"); ok {
		t.Fatal("defaultParseBool('false') = true")
	}
	if ok, _ := defaultParseBool("0"); ok {
		t.Fatal("defaultParseBool('0') = true")
	}
	if ok, _ := defaultParseBool(""); ok {
		t.Fatal("defaultParseBool('') = true")
	}
	if ok, _ := defaultParseBool("no"); ok {
		t.Fatal("defaultParseBool('no') = true")
	}
	if _, err := defaultParseBool("unknown"); err == nil {
		t.Fatal("defaultParseBool('unknown') error = nil")
	}
}
