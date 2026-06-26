package conf

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestLoadRemoteWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Config-Token") != "secret" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte("app:\n  name: remote"))
	}))
	defer server.Close()
	calledFactory := false
	c, err := LoadRemoteWithOptions(server.URL+"/app.yaml", LoadOptions{
		Timeout: time.Second,
		Headers: http.Header{"X-Config-Token": []string{"secret"}},
		RequestFactory: func(ctx context.Context, rawURL string) (*http.Request, error) {
			calledFactory = true
			return http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledFactory {
		t.Fatal("request factory was not called")
	}
	if got := c.GetByGroup("app", "name"); got != "remote" {
		t.Fatalf("remote app.name = %q", got)
	}
	if _, err := LoadRemoteWithOptions(server.URL+"/app.yaml", LoadOptions{MaxBytes: 3, Headers: http.Header{"X-Config-Token": []string{"secret"}}}); err == nil {
		t.Fatal("LoadRemoteWithOptions max bytes error = nil")
	}
}

func TestLoadRemoteWithOptionsRejectsInvalidProviderResponses(t *testing.T) {
	_, err := LoadRemoteWithOptions("https://config.example/app.yaml", LoadOptions{
		RemoteClient: &http.Client{Transport: confRoundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, nil
		})},
	})
	assertConfCode(t, err, knifer.ErrCodeInternal)

	_, err = LoadRemoteWithOptions("https://config.example/app.yaml", LoadOptions{
		RemoteClient: &http.Client{Transport: confRoundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("app:\n  name: remote"))}, nil
		})},
		RequestFactory: func(context.Context, string) (*http.Request, error) { return nil, nil },
	})
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = readAllLimit(nil, 1)
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}
