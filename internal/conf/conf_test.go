package conf

import (
	"reflect"
	"testing"
)

func TestParseSetting(t *testing.T) {
	s, err := Parse("name = gokit\n[server]\nport=8080\ndebug=true")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("name"); got != "gokit" {
		t.Fatalf("Get(name) = %q", got)
	}
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("GetByGroup(server, port) = %q", got)
	}
	if got := s.GetOrDefault("missing", "def"); got != "def" {
		t.Fatalf("GetOrDefault() = %q", got)
	}
}

func TestParseYAML(t *testing.T) {
	s, err := ParseYAML("app: gokit\nserver:\n  port: 8080")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("app"); got != "gokit" {
		t.Fatalf("Get(app) = %q", got)
	}
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("GetByGroup(server, port) = %q", got)
	}
}

func TestNilConfReadMethodsAreEmptyAndSafe(t *testing.T) {
	var s *Conf

	if got := s.Groups(); len(got) != 0 {
		t.Fatalf("Groups() = %v, want empty", got)
	}
	if got := s.Keys("missing"); len(got) != 0 {
		t.Fatalf("Keys(missing) = %v, want empty", got)
	}
	if got := s.ToMap(); len(got) != 0 {
		t.Fatalf("ToMap() = %v, want empty", got)
	}
}

func TestGroupsKeysAndToMapKeepStableSemantics(t *testing.T) {
	s := New()
	s.Set("root", "value")
	s.SetByGroup("server", "port", "8080")
	s.SetByGroup("server", "host", "localhost")
	s.SetByGroup("app", "name", "gokit")

	if got, want := s.Groups(), []string{"", "app", "server"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Groups() = %v, want %v", got, want)
	}
	if got, want := s.Keys("server"), []string{"host", "port"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys(server) = %v, want %v", got, want)
	}

	m := s.ToMap()
	m["server"]["port"] = "9090"
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("ToMap() returned shallow copy, source port = %q", got)
	}
}
