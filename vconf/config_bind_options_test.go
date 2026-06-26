package vconf_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeBindAndSchemaParserOptions(t *testing.T) {
	s, err := vconf.Parse(`
flag=yes
count=custom-int
amount=custom-uint
ratio=custom-float
items=1,2,3
schema_bool=yes
schema_float=custom-float
choice=blue
`)
	if err != nil {
		t.Fatal(err)
	}

	type bindConfig struct {
		Flag   bool    `conf:"flag"`
		Count  int     `conf:"count"`
		Amount uint    `conf:"amount"`
		Ratio  float64 `conf:"ratio"`
		Items  []int   `conf:"items"`
	}
	var cfg bindConfig
	if err := s.BindWithOptions(&cfg,
		vconf.WithBindBoolParser(func(value string) (bool, error) {
			return value == "yes", nil
		}),
		vconf.WithBindIntParser(func(value string, base int, bitSize int) (int64, error) {
			if value == "custom-int" {
				return 42, nil
			}
			return 7, nil
		}),
		vconf.WithBindUintParser(func(value string, base int, bitSize int) (uint64, error) {
			if value == "custom-uint" {
				return 9, nil
			}
			return 3, nil
		}),
		vconf.WithBindFloatParser(func(value string, bitSize int) (float64, error) {
			if value == "custom-float" {
				return 1.5, nil
			}
			return 0, errors.New("unexpected float")
		}),
	); err != nil {
		t.Fatalf("BindWithOptions() error = %v", err)
	}
	if !cfg.Flag || cfg.Count != 42 || cfg.Amount != 9 || cfg.Ratio != 1.5 || !reflect.DeepEqual(cfg.Items, []int{7, 7, 7}) {
		t.Fatalf("BindWithOptions cfg = %#v", cfg)
	}

	err = s.ValidateSchemaWithOptions(vconf.Schema{Fields: []vconf.FieldRule{
		{Key: "schema_bool", Required: true, Type: vconf.TypeBool},
		{Key: "schema_float", Required: true, Type: vconf.TypeFloat},
		{Key: "choice", Required: true, Choices: []string{"red", "blue"}},
	}},
		vconf.WithSchemaBoolParser(func(value string) (bool, error) {
			return value == "yes", nil
		}),
		vconf.WithSchemaFloatParser(func(value string, bitSize int) (float64, error) {
			if value == "custom-float" {
				return 2.5, nil
			}
			return 0, errors.New("unexpected schema float")
		}),
	)
	if err != nil {
		t.Fatalf("ValidateSchemaWithOptions() error = %v", err)
	}
}

func TestFacadeBindDecodeHookApplied(t *testing.T) {
	s := vconf.New()
	s.Set("start", "2026-06-22")
	type target struct {
		Start time.Time `conf:"start"`
	}
	var cfg target
	err := s.BindWithOptions(&cfg,
		vconf.WithBindDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
				return time.Parse(time.DateOnly, value.(string))
			}
			return value, nil
		}),
	)
	if err != nil {
		t.Fatalf("BindWithOptions() with hook error = %v", err)
	}
	if got := cfg.Start.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Start = %q", got)
	}
}

func TestDynamicConfigContractMatrix(t *testing.T) {
	cfg, err := vconf.Parse(`
name=knifer-go
[server]
port=8080
enabled=true
tags=api,admin
`)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "string", got: cfg.Get("name"), want: "knifer-go"},
		{name: "missing default", got: cfg.GetOrDefault("missing", "fallback"), want: "fallback"},
		{name: "group int", got: cfg.GetIntByGroup("server", "port", 0), want: 8080},
		{name: "group bool", got: cfg.GetBoolByGroup("server", "enabled", false), want: true},
		{name: "group string", got: cfg.GetByGroup("server", "tags"), want: "api,admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %#v, want %#v", tt.got, tt.want)
			}
		})
	}

	type serverConfig struct {
		Port    int      `conf:"port"`
		Enabled bool     `conf:"enabled"`
		Tags    []string `conf:"tags"`
	}
	var bound serverConfig
	if err := cfg.BindGroup("server", &bound); err != nil {
		t.Fatalf("BindGroup() error = %v", err)
	}
	if !reflect.DeepEqual(bound, serverConfig{Port: 8080, Enabled: true, Tags: []string{"api", "admin"}}) {
		t.Fatalf("BindGroup() = %#v", bound)
	}
}

func FuzzDynamicConfigScalarContract(f *testing.F) {
	f.Add("42")
	f.Add("true")
	f.Add("knifer-go")
	f.Fuzz(func(t *testing.T, value string) {
		cfg := vconf.New()
		cfg.Set("value", value)
		if got := cfg.Get("value"); got != value {
			t.Fatalf("Get(value) = %q, want %q", got, value)
		}
		_ = cfg.GetInt("value", -1)
		_ = cfg.GetBool("value", false)
	})
}
