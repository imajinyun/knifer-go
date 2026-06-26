package vskt_test

import (
	"bytes"
	"testing"

	"github.com/imajinyun/knifer-go/vskt"
)

func TestFacadeFuncDecoderAndEncoder(t *testing.T) {
	decoded := false
	decoder := vskt.FuncDecoder[string](func(session *vskt.AioSession, readBuffer *bytes.Buffer) (string, bool) {
		decoded = true
		if session != nil {
			t.Fatal("session = non-nil, want nil")
		}
		return readBuffer.String(), true
	})
	value, ok := decoder.Decode(nil, bytes.NewBufferString("payload"))
	if !decoded || !ok || value != "payload" {
		t.Fatalf("Decode() = (%q, %v), called=%v", value, ok, decoded)
	}

	encoded := false
	encoder := vskt.FuncEncoder[string](func(session *vskt.AioSession, writeBuffer *bytes.Buffer, data string) {
		encoded = true
		if session != nil {
			t.Fatal("session = non-nil, want nil")
		}
		writeBuffer.WriteString("encoded:" + data)
	})
	var out bytes.Buffer
	encoder.Encode(nil, &out, "payload")
	if !encoded || out.String() != "encoded:payload" {
		t.Fatalf("Encode() output = %q, called=%v", out.String(), encoded)
	}
}
