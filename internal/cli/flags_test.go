package cli

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestFlagParserParsesTypedValuesAndRemainingArgs(t *testing.T) {
	parser := NewFlagParser("serve")
	host := parser.String("host", "127.0.0.1", "host to bind")
	port := parser.Int("port", 8080, "port to bind")
	debug := parser.Bool("debug", false, "enable debug")
	timeout := parser.Duration("timeout", time.Second, "request timeout")
	result, err := parser.Parse([]string{
		"--host", "0.0.0.0",
		"--port", "9090",
		"--debug",
		"--timeout", "2s",
		"api",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if *host != "0.0.0.0" || *port != 9090 || !*debug || *timeout != 2*time.Second {
		t.Fatalf("values host=%q port=%d debug=%v timeout=%s", *host, *port, *debug, *timeout)
	}
	if len(result.Args) != 1 || result.Args[0] != "api" {
		t.Fatalf("remaining args = %#v", result.Args)
	}
}

func TestFlagParserKeepsDefaults(t *testing.T) {
	parser := NewFlagParser("serve")
	host := parser.String("host", "127.0.0.1", "host to bind")
	port := parser.Int("port", 8080, "port to bind")
	_, err := parser.Parse(nil)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if *host != "127.0.0.1" || *port != 8080 {
		t.Fatalf("defaults host=%q port=%d", *host, *port)
	}
}

func TestFlagParserReturnsUsageErrorForUnknownFlag(t *testing.T) {
	parser := NewFlagParser("serve")
	_, err := parser.Parse([]string{"--missing"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("Parse unknown flag error = %v, want ErrUsage", err)
	}
}

func TestFlagParserWritesUsageToInjectedWriter(t *testing.T) {
	var out bytes.Buffer
	parser := NewFlagParser("serve", WithFlagOutput(&out))
	parser.String("host", "127.0.0.1", "host to bind")
	parser.Usage()
	if !bytes.Contains(out.Bytes(), []byte("host to bind")) {
		t.Fatalf("usage output = %q", out.String())
	}
}

func TestNilFlagOutputDoesNotClearPreviousWriter(t *testing.T) {
	var out bytes.Buffer
	parser := NewFlagParser("serve", WithFlagOutput(&out), WithFlagOutput(nil))
	parser.String("host", "127.0.0.1", "host to bind")
	parser.Usage()
	if !bytes.Contains(out.Bytes(), []byte("host to bind")) {
		t.Fatalf("usage output = %q", out.String())
	}
}
