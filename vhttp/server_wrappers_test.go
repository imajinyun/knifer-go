package vhttp_test

import (
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeSimpleServerWrappers(t *testing.T) {
	if vhttp.NewSimpleServer(0) == nil {
		t.Fatal("NewSimpleServer returned nil")
	}
	if vhttp.NewSimpleServerAddr("127.0.0.1:0") == nil {
		t.Fatal("NewSimpleServerAddr returned nil")
	}

	static := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0")
	static.SetRootWithOptions(".",
		vhttp.WithStaticFileSystem(http.Dir(".")),
		vhttp.WithStaticFS(os.DirFS(".")),
		vhttp.WithFileServerFactory(func(http.FileSystem) http.Handler {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			})
		}),
		vhttp.WithStaticHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		})),
	)

	listener, listenErr := net.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		t.Fatal(listenErr)
	}
	defer listener.Close()
	if server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0",
		vhttp.WithListener(listener),
		vhttp.WithListenAndServeFunc(func(*http.Server) error {
			return http.ErrServerClosed
		}),
	); server == nil {
		t.Fatal("WithListener server returned nil")
	}
}
