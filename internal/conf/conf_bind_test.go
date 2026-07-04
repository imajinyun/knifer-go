package conf

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestBindWithOptionsUsesParsers(t *testing.T) {
	s := New()
	s.SetByGroup("server", "port", "custom-int")
	s.SetByGroup("server", "debug", "custom-bool")
	s.SetByGroup("server", "ratio", "custom-float")
	s.SetByGroup("server", "ids", "a,b")

	type serverConf struct {
		Port  int     `conf:"port"`
		Debug bool    `conf:"debug"`
		Ratio float64 `conf:"ratio"`
		IDs   []uint  `conf:"ids"`
	}
	var cfg serverConf
	var intCalled, boolCalled, floatCalled, uintCalled int
	err := s.BindGroupWithOptions("server", &cfg,
		WithBindIntParser(func(text string, base, bitSize int) (int64, error) {
			intCalled++
			if text == "custom-int" {
				return 8080, nil
			}
			return strconv.ParseInt(text, base, bitSize)
		}),
		WithBindBoolParser(func(text string) (bool, error) {
			boolCalled++
			return text == "custom-bool", nil
		}),
		WithBindFloatParser(func(text string, bitSize int) (float64, error) {
			floatCalled++
			if text == "custom-float" {
				return 0.75, nil
			}
			return strconv.ParseFloat(text, bitSize)
		}),
		WithBindUintParser(func(text string, base, bitSize int) (uint64, error) {
			uintCalled++
			switch text {
			case "a":
				return 1, nil
			case "b":
				return 2, nil
			default:
				return strconv.ParseUint(text, base, bitSize)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, serverConf{Port: 8080, Debug: true, Ratio: 0.75, IDs: []uint{1, 2}}) {
		t.Fatalf("BindGroupWithOptions = %#v", cfg)
	}
	if intCalled != 1 || boolCalled != 1 || floatCalled != 1 || uintCalled != 2 {
		t.Fatalf("parser calls int=%d bool=%d float=%d uint=%d", intCalled, boolCalled, floatCalled, uintCalled)
	}
}

func TestBindWithDecodeHook(t *testing.T) {
	s := New()
	s.Set("start", "2026-06-22")
	type target struct {
		Start time.Time `conf:"start"`
	}
	var cfg target
	called := 0
	err := s.BindWithOptions(&cfg,
		WithBindDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			called++
			if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
				return time.Parse(time.DateOnly, value.(string))
			}
			return value, nil
		}),
	)
	if err != nil {
		t.Fatalf("BindWithOptions() with hook error = %v", err)
	}
	if called != 1 {
		t.Fatalf("hook calls = %d, want 1", called)
	}
	if got := cfg.Start.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Start = %q", got)
	}
}

func TestBindDecodeHookRejectsUnsafeReflectNumericConversion(t *testing.T) {
	s := New()
	s.Set("count", "ignored")
	type target struct {
		Count int8 `conf:"count"`
	}
	var cfg target
	err := s.BindWithOptions(&cfg,
		WithBindDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			if to.Kind() == reflect.Int8 {
				return int16(128), nil
			}
			return value, nil
		}),
	)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("BindWithOptions() error = %v, want integer overflow", err)
	}
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)

	err = s.BindWithOptions(&cfg,
		WithBindDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			if to.Kind() == reflect.Int8 {
				return int16(127), nil
			}
			return value, nil
		}),
	)
	if err != nil {
		t.Fatalf("BindWithOptions() safe conversion error = %v", err)
	}
	if cfg.Count != 127 {
		t.Fatalf("Count = %d, want 127", cfg.Count)
	}
}

func TestBindReportsNestedFieldPath(t *testing.T) {
	s := New()
	s.Set("server.port", "bad")
	type server struct {
		Port int `conf:"port"`
	}
	type target struct {
		Server server `conf:"server"`
	}
	var cfg target
	err := s.Bind(&cfg)
	if err == nil {
		t.Fatal("Bind() error = nil, want nested parse error")
	}
	if got := err.Error(); !strings.Contains(got, "bind server.port") {
		t.Fatalf("Bind() error = %q, want nested key path", got)
	}
}
