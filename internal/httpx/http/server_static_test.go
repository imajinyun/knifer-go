package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSimpleServerSetRootWithOptions(t *testing.T) {
	srv := NewSimpleServer(0)
	called := false
	srv.SetRootWithOptions("ignored",
		WithStaticFS(fstest.MapFS{"static.txt": {Data: []byte("ok")}}),
		WithFileServerFactory(func(fileSystem http.FileSystem) http.Handler {
			called = true
			if fileSystem == nil {
				t.Fatal("file system should be set")
			}
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("static"))
			})
		}),
	)
	if !called {
		t.Fatal("custom file server factory was not used")
	}
}

func TestSimpleServerSetRootWithStaticHandler(t *testing.T) {
	srv := NewSimpleServer(0)
	called := false
	srv.SetRootWithOptions("ignored", WithStaticHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})))
	req, err := http.NewRequest(string(MethodGet), "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.mux.ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Fatal("static handler was not used")
	}
}
