package bean

import "testing"

type embeddedProfile struct {
	Trace string `bean:"trace_id"`
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

func TestCopyPropertiesStructToStructWithAliasAndWeakConversion(t *testing.T) {
	src := sourceProfile{
		embeddedProfile: embeddedProfile{Trace: "t-1"},
		Name:            "alice",
		Age:             "42",
		Admin:           "yes",
		Skip:            "ignored",
	}
	var dst targetProfile
	if err := CopyProperties(src, &dst, WithIgnoreEmpty(true)); err != nil {
		t.Fatalf("CopyProperties() error = %v", err)
	}
	if dst.Name != "alice" || dst.Age != 42 || !dst.Admin || dst.Trace != "t-1" || dst.Empty != "" {
		t.Fatalf("dst = %+v", dst)
	}
}

func TestCopyPropertiesMapToStruct(t *testing.T) {
	src := map[string]any{
		"displayName": "bob",
		"age":         7.9,
		"admin":       1,
		"trace_id":    "t-2",
	}
	var dst targetProfile
	if err := Copy(src, &dst); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if dst.Name != "bob" || dst.Age != 7 || !dst.Admin || dst.Trace != "t-2" {
		t.Fatalf("dst = %+v", dst)
	}
}

func TestToMapUsesPrimaryTagAndOmit(t *testing.T) {
	got, err := ToMap(sourceProfile{Name: "alice", Age: "18", Skip: "hidden"})
	if err != nil {
		t.Fatalf("ToMap() error = %v", err)
	}
	if got["name"] != "alice" || got["age"] != "18" {
		t.Fatalf("map = %#v", got)
	}
	if _, ok := got["Skip"]; ok {
		t.Fatalf("omit field leaked: %#v", got)
	}
}

func TestWeaklyTypedDisabled(t *testing.T) {
	var dst targetProfile
	err := CopyProperties(map[string]any{"age": "42"}, &dst, WithWeaklyTyped(false))
	if err == nil {
		t.Fatal("expected strict assignment error")
	}
}
