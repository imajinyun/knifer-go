package vhttp_test

import (
	"io"
	"net/http"

	"github.com/imajinyun/knifer-go/vhttp"
)

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

type closeErrorWriteCloser struct {
	io.Writer
	err error
}

func (w closeErrorWriteCloser) Close() error { return w.err }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func executeRequest(req *vhttp.Request) *vhttp.Response {
	return req.Execute()
}

func errorString(err *vhttp.Error) string {
	return err.Error()
}
