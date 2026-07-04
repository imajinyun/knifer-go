package resty

import (
	"io"
	"net/http"
)

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

type closeErrorWriteCloser struct {
	io.Writer
	err error
}

func (w closeErrorWriteCloser) Close() error { return w.err }

type restyRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f restyRoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
