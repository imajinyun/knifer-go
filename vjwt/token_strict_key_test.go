package vjwt_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vjwt"
)

func TestCreateTokenWithOptionsStrictKey(t *testing.T) {
	if token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(map[string]any{vjwt.JWTPayloadSubject: "alice"}),
		vjwt.WithTokenKey([]byte("weak")),
		vjwt.WithTokenStrictKey(),
	); err == nil || token != "" {
		t.Fatalf("strict weak key token=%q err=%v, want error", token, err)
	}
}
