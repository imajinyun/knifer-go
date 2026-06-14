package resty

import (
	"testing"

	grestry "resty.dev/v3"
)

func TestRestyClientFactoryProviderLifecycle(t *testing.T) {
	ResetDefaultRestyClientProvider()
	t.Cleanup(ResetDefaultRestyClientProvider)

	defaultCalled := 0
	ConfigureDefaultRestyClientProvider(func() *grestry.Client {
		defaultCalled++
		return grestry.New()
	})
	client := NewIsolatedRequest(MethodGet, "http://example.com").buildClient()
	if client == nil || defaultCalled != 1 {
		t.Fatalf("default provider client=%v called=%d", client, defaultCalled)
	}

	perCallCalled := 0
	client = NewIsolatedRequest(MethodGet, "http://example.com", WithRestyClientFactory(func() *grestry.Client {
		perCallCalled++
		return grestry.New()
	})).buildClient()
	if client == nil || perCallCalled != 1 || defaultCalled != 1 {
		t.Fatalf("per-call factory client=%v perCall=%d default=%d", client, perCallCalled, defaultCalled)
	}

	client = NewIsolatedRequest(MethodGet, "http://example.com", WithRestyClientFactory(func() *grestry.Client { return nil })).buildClient()
	if client == nil || defaultCalled != 2 {
		t.Fatalf("nil per-call factory client=%v default=%d", client, defaultCalled)
	}

	ResetDefaultRestyClientProvider()
	client = NewIsolatedRequest(MethodGet, "http://example.com").buildClient()
	if client == nil {
		t.Fatal("reset default provider should create a client")
	}
}
