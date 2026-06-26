package vresty_test

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vresty"
)

func TestFacadeWithTLSConfigAndOptions(t *testing.T) {
	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	_ = vresty.WithTLSConfig(tlsCfg)

	// WithAllowedHosts
	_ = vresty.WithAllowedHosts("example.com", "example.org")

	// WithLookupIP — constructor only stores the function, doesn't call it
	_ = vresty.WithLookupIP(func(ctx context.Context, host string) ([]net.IP, error) {
		return nil, nil
	})
}

func TestFacadeErrorReturningMethods(t *testing.T) {
	// GetWithTimeoutE — invalid URL should return error immediately
	_, err := vresty.GetWithTimeoutE("invalid://not-a-url", time.Second)
	if err == nil {
		t.Fatal("GetWithTimeoutE should return error for invalid URL")
	}

	_, err = vresty.GetWithTimeoutEWithOptions("invalid://not-a-url", time.Second)
	if err == nil {
		t.Fatal("GetWithTimeoutEWithOptions should return error for invalid URL")
	}

	// GetWithParamsE — same check
	_, err = vresty.GetWithParamsE("invalid://not-a-url", nil)
	if err == nil {
		t.Fatal("GetWithParamsE should return error for invalid URL")
	}
	_, err = vresty.GetWithParamsEWithOptions("invalid://not-a-url", nil)
	if err == nil {
		t.Fatal("GetWithParamsEWithOptions should return error for invalid URL")
	}

	// PostStringE
	_, err = vresty.PostStringE("invalid://not-a-url", "body")
	if err == nil {
		t.Fatal("PostStringE should return error for invalid URL")
	}
	_, err = vresty.PostStringEWithOptions("invalid://not-a-url", "body")
	if err == nil {
		t.Fatal("PostStringEWithOptions should return error for invalid URL")
	}

	// PostFormE
	_, err = vresty.PostFormE("invalid://not-a-url", nil)
	if err == nil {
		t.Fatal("PostFormE should return error for invalid URL")
	}
	_, err = vresty.PostFormEWithOptions("invalid://not-a-url", nil)
	if err == nil {
		t.Fatal("PostFormEWithOptions should return error for invalid URL")
	}
	_, err = vresty.PostFormSafeE("invalid://not-a-url", nil)
	if err == nil {
		t.Fatal("PostFormSafeE should return error for invalid URL")
	}

	// PostJSONE
	_, err = vresty.PostJSONE("invalid://not-a-url", "{}")
	if err == nil {
		t.Fatal("PostJSONE should return error for invalid URL")
	}
	_, err = vresty.PostJSONEWithOptions("invalid://not-a-url", "{}")
	if err == nil {
		t.Fatal("PostJSONEWithOptions should return error for invalid URL")
	}
	_, err = vresty.PostJSONSafeE("invalid://not-a-url", "{}")
	if err == nil {
		t.Fatal("PostJSONSafeE should return error for invalid URL")
	}
}

func TestFacadeGetGlobalTimeout(t *testing.T) {
	orig := vresty.GetGlobalTimeout()
	vresty.SetGlobalTimeout(5 * time.Second)
	if vresty.GetGlobalTimeout() != 5*time.Second {
		t.Fatalf("GetGlobalTimeout = %v, want 5s", vresty.GetGlobalTimeout())
	}
	vresty.SetGlobalTimeout(orig)
}
