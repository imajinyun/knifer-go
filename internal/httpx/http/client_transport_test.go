package http

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClientUsesCapturedConfig(t *testing.T) {
	oldUA := GetGlobalUserAgent()
	oldFollow := GetGlobalFollowRedirects()
	defer SetGlobalUserAgent(oldUA)
	defer SetGlobalFollowRedirects(oldFollow)

	SetGlobalUserAgent("client-agent")
	SetGlobalFollowRedirects(false)
	client := NewClient()
	SetGlobalUserAgent("mutated-agent")
	SetGlobalFollowRedirects(true)

	req := client.Get("https://example.com")
	if req.userAgent != "client-agent" {
		t.Fatalf("client request userAgent = %q, want captured client-agent", req.userAgent)
	}
	if req.followRedir == nil || *req.followRedir {
		t.Fatalf("client request followRedirects = %v, want captured false", req.followRedir)
	}

	isolated := NewIsolatedClient().Get("https://example.com")
	if isolated.userAgent != "" || isolated.followRedir == nil || !*isolated.followRedir {
		t.Fatalf("isolated client defaults ua=%q follow=%v", isolated.userAgent, isolated.followRedir)
	}
}

func TestRequestCustomTransport(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := io.NopCloser(strings.NewReader("intercepted"))
		return &http.Response{
			StatusCode: 200,
			Body:       body,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Request:    req,
		}, nil
	})
	body := Get("http://will-not-call/").Transport(rt).Execute().Body()
	if body != "intercepted" {
		t.Fatalf("body: %q", body)
	}
}

func TestDefaultTransportIsReused(t *testing.T) {
	clientA := Get("https://example.com").buildClient()
	clientB := Post("https://example.com").Timeout(time.Second).buildClient()
	shared := getDefaultTransport()

	if clientA.Transport != shared {
		t.Fatalf("default request transport = %p, want shared default transport %p", clientA.Transport, shared)
	}
	if clientB.Transport != shared {
		t.Fatalf("request with timeout transport = %p, want shared default transport %p", clientB.Transport, shared)
	}
}

func TestTransportProviderEvaluatedWhenBuildingClient(t *testing.T) {
	calls := 0
	custom := &http.Transport{}
	req := Get("https://example.com", WithTransportProvider(func() http.RoundTripper {
		calls++
		return custom
	}))
	if calls != 0 {
		t.Fatalf("transport provider called during construction: %d", calls)
	}
	client := req.buildClient()
	if calls != 1 || client.Transport != custom {
		t.Fatalf("transport provider calls=%d transport=%#v, want custom", calls, client.Transport)
	}
}

func TestDefaultTransportProviderCanBeConfiguredAndReset(t *testing.T) {
	custom := &http.Transport{MaxIdleConnsPerHost: 7}
	ConfigureDefaultTransportProvider(func() *http.Transport { return custom })
	t.Cleanup(ResetDefaultTransport)

	client := Get("https://example.com").buildClient()
	if client.Transport != custom {
		t.Fatalf("configured default transport = %p, want %p", client.Transport, custom)
	}

	ResetDefaultTransport()
	client = Get("https://example.com").buildClient()
	if client.Transport == custom {
		t.Fatal("ResetDefaultTransport should clear configured transport")
	}
	if _, ok := client.Transport.(*http.Transport); !ok {
		t.Fatalf("reset default transport type = %T, want *http.Transport", client.Transport)
	}
}

func TestDefaultTransportProviderConcurrentConfigureAndUse(t *testing.T) {
	ResetDefaultTransport()
	t.Cleanup(ResetDefaultTransport)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ConfigureDefaultTransportProvider(func() *http.Transport {
					return &http.Transport{MaxIdleConnsPerHost: 7}
				})
				if client := Get("https://example.com").buildClient(); client.Transport == nil {
					t.Error("configured default transport produced nil transport")
				}
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ResetDefaultTransport()
				if client := Post("https://example.com").Timeout(time.Second).buildClient(); client.Transport == nil {
					t.Error("reset default transport produced nil transport")
				}
			}
		}()
	}
	wg.Wait()
}
