package conf

import (
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestLoadWithOptionsReadFileProvider(t *testing.T) {
	c, err := LoadWithOptions("virtual.setting", LoadOptions{
		MaxBytes: 16,
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			if path != "virtual.setting" || maxBytes != 16 {
				t.Fatalf("read path=%q maxBytes=%d", path, maxBytes)
			}
			return []byte("name=fake"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "fake" {
		t.Fatalf("loaded name = %q", got)
	}
}

func TestLoadWithOptionsReadFileProviderUsesDefaultMaxBytes(t *testing.T) {
	_, err := LoadWithOptions("virtual.setting", LoadOptions{
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			if maxBytes != DefaultMaxBytes {
				t.Fatalf("default maxBytes=%d, want %d", maxBytes, DefaultMaxBytes)
			}
			return []byte("name=fake"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadWithOptionsAllowsExplicitUnlimitedMaxBytes(t *testing.T) {
	_, err := LoadWithOptions("virtual.setting", LoadOptions{
		MaxBytes: -1,
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			if maxBytes != -1 {
				t.Fatalf("maxBytes=%d, want -1", maxBytes)
			}
			return []byte("name=fake"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadWithOptionsReadFileProviderEnforcesMaxBytes(t *testing.T) {
	_, err := LoadWithOptions("virtual.setting", LoadOptions{
		MaxBytes: 4,
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			return []byte("name=fake"), nil
		},
	})
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}
