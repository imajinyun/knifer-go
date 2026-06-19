package http

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestWithStaticFileSystem(t *testing.T) {
	cfg := staticConfig{}
	opt := WithStaticFileSystem(http.Dir("."))
	opt(&cfg)
	if cfg.fileSystem == nil {
		t.Fatal("WithStaticFileSystem should set fileSystem")
	}
}

func TestNewSimpleServerAddr(t *testing.T) {
	srv := NewSimpleServerAddr(":0")
	if srv == nil {
		t.Fatal("NewSimpleServerAddr returned nil")
	}
	_ = srv.Stop(time.Second)
}

func TestWithServerErrorLog(t *testing.T) {
	var s http.Server
	logger := log.New(io.Discard, "", 0)
	opt := WithServerErrorLog(logger)
	opt(&s)
	if s.ErrorLog != logger {
		t.Fatal("WithServerErrorLog did not set ErrorLog")
	}
}

func TestWithHTTPServer(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		var s http.Server
		s.Addr = ":8080"
		opt := WithHTTPServer(nil)
		opt(&s)
		if s.Addr != ":8080" {
			t.Fatal("WithHTTPServer(nil) should not modify s")
		}
	})

	t.Run("copies fields", func(t *testing.T) {
		src := &http.Server{
			Addr:                         ":9090",
			Handler:                      http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			DisableGeneralOptionsHandler: true,
			ReadTimeout:                  5 * time.Second,
			ReadHeaderTimeout:            3 * time.Second,
			WriteTimeout:                 10 * time.Second,
			IdleTimeout:                  60 * time.Second,
			MaxHeaderBytes:               8192,
			ErrorLog:                     log.New(io.Discard, "", 0),
		}
		var dst http.Server
		WithHTTPServer(src)(&dst)
		if dst.Addr != src.Addr ||
			dst.ReadTimeout != src.ReadTimeout ||
			dst.WriteTimeout != src.WriteTimeout ||
			dst.IdleTimeout != src.IdleTimeout ||
			dst.MaxHeaderBytes != src.MaxHeaderBytes ||
			dst.ErrorLog != src.ErrorLog {
			t.Fatal("WithHTTPServer did not copy all fields")
		}
	})
}

func TestSimpleServerAddHandler(t *testing.T) {
	srv := NewSimpleServer(0)

	var called bool
	srv.AddHandler("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.mux.ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Fatal("AddHandler handler was not called")
	}
}

func TestSimpleServerSetRoot(t *testing.T) {
	srv := NewSimpleServer(0)

	result := srv.SetRoot(".")
	if result != srv {
		t.Fatal("SetRoot should return self")
	}
}

func TestWithSaveStat(t *testing.T) {
	cfg := saveConfig{}
	statFn := func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	opt := WithSaveStat(statFn)
	opt(&cfg)
	if cfg.stat == nil {
		t.Fatal("WithSaveStat should set stat")
	}
}

func TestHTTPResponseContentLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "5")
		_, _ = w.Write([]byte("hello"))
	}))
	defer ts.Close()

	client := NewClient()
	resp := client.Get(ts.URL).Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute error = %v", resp.Err())
	}
	if n := resp.ContentLength(); n != 5 {
		t.Fatalf("ContentLength = %d, want 5", n)
	}
}

func TestMustExecute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
		defer ts.Close()

		client := NewClient()
		resp := client.Get(ts.URL).MustExecute()
		if resp.Body() != "ok" {
			t.Fatalf("MustExecute body = %q, want %q", resp.Body(), "ok")
		}
	})

	t.Run("panic on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("MustExecute should panic on error")
			}
		}()
		NewRequest("INVALID", "://bad").MustExecute()
	})
}

func TestClientGetSafeAndPostSafe(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		_, _ = w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	client := NewClient()
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})

	resp := client.GetSafe(ts.URL, allowLocal).Execute()
	if resp.Err() != nil {
		t.Fatalf("GetSafe Execute error = %v", resp.Err())
	}
	if got := resp.Headers()["X-Method"]; len(got) != 1 || got[0] != "GET" {
		t.Fatalf("GetSafe method = %v", got)
	}

	resp = client.PostSafe(ts.URL, allowLocal).Execute()
	if resp.Err() != nil {
		t.Fatalf("PostSafe Execute error = %v", resp.Err())
	}
	if got := resp.Headers()["X-Method"]; len(got) != 1 || got[0] != "POST" {
		t.Fatalf("PostSafe method = %v", got)
	}
}

func TestClientGetAndPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer ts.Close()

	client := NewClient()

	resp := client.Get(ts.URL).Execute()
	if resp.Body() != "GET" {
		t.Fatalf("Get body = %q", resp.Body())
	}
	resp = client.Post(ts.URL).Execute()
	if resp.Body() != "POST" {
		t.Fatalf("Post body = %q", resp.Body())
	}
}

func TestPostFormE(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		_, _ = w.Write([]byte(r.FormValue("key")))
	}))
	defer ts.Close()

	body, err := PostFormE(ts.URL, map[string]any{"key": "val"})
	if err != nil {
		t.Fatalf("PostFormE error = %v", err)
	}
	if body != "val" {
		t.Fatalf("PostFormE body = %q, want %q", body, "val")
	}
}

func TestContentLengthNilResponse(t *testing.T) {
	resp := &HTTPResponse{}
	if n := resp.ContentLength(); n != -1 {
		t.Fatalf("ContentLength = %d, want -1 for nil response", n)
	}
}

func TestSimpleServerAddHandlerLeadingSlash(t *testing.T) {
	srv := NewSimpleServer(0)

	srv.AddHandler("test-no-slash", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req, err := http.NewRequest("GET", "/test-no-slash", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

func TestSimpleServerAddAction(t *testing.T) {
	srv := NewSimpleServer(0)

	var called bool
	srv.AddAction("/action", func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req, err := http.NewRequest("GET", "/action", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.mux.ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Fatal("AddAction handler was not called")
	}
}

func TestDownloadStringSafeE(t *testing.T) {
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-data"))
	}))
	defer ts.Close()

	body, err := DownloadStringSafeE(ts.URL, "", allowLocal)
	if err != nil {
		t.Fatalf("DownloadStringSafeE error = %v", err)
	}
	if body != "safe-data" {
		t.Fatalf("body = %q, want %q", body, "safe-data")
	}
}

func TestDownloadSafe(t *testing.T) {
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("writer-data"))
	}))
	defer ts.Close()

	var buf strings.Builder
	if _, err := DownloadSafe(ts.URL, &buf, allowLocal); err != nil {
		t.Fatalf("DownloadSafe error = %v", err)
	}
	if buf.String() != "writer-data" {
		t.Fatalf("written = %q, want %q", buf.String(), "writer-data")
	}
}

func TestDownloadBytesSafeE(t *testing.T) {
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("bytes-data"))
	}))
	defer ts.Close()

	data, err := DownloadBytesSafeE(ts.URL, allowLocal)
	if err != nil {
		t.Fatalf("DownloadBytesSafeE error = %v", err)
	}
	if string(data) != "bytes-data" {
		t.Fatalf("data = %q, want %q", string(data), "bytes-data")
	}
}

func TestPostFormSafeE(t *testing.T) {
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		_, _ = w.Write([]byte(r.FormValue("k")))
	}))
	defer ts.Close()

	body, err := PostFormSafeE(ts.URL, map[string]any{"k": "v"}, allowLocal)
	if err != nil {
		t.Fatalf("PostFormSafeE error = %v", err)
	}
	if body != "v" {
		t.Fatalf("PostFormSafeE body = %q, want %q", body, "v")
	}
}

func TestPostJSONSafeE(t *testing.T) {
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("json-ok"))
	}))
	defer ts.Close()

	body, err := PostJSONSafeE(ts.URL, `{"msg":"hi"}`, allowLocal)
	if err != nil {
		t.Fatalf("PostJSONSafeE error = %v", err)
	}
	if body != "json-ok" {
		t.Fatalf("PostJSONSafeE body = %q, want %q", body, "json-ok")
	}
}
