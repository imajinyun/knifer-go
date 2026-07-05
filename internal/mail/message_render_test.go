package mail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewMessageRendersMixedRelatedAlternative(t *testing.T) {
	msg, err := NewMessage(
		WithFrom("Sender <sender@example.com>"),
		WithTo("Receiver <receiver@example.com>"),
		WithCc("copy@example.com"),
		WithBcc("hidden@example.com"),
		WithSubject("hello 世界"),
		WithText("plain body"),
		WithHTML(`<p><img src="cid:logo">html body</p>`),
		WithInline("logo.png", "logo", []byte("inline-data"), "image/png"),
		WithAttachment("report.txt", []byte("attachment-data"), TypeTextPlain),
		WithDate(time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)),
		WithMessageID("message@example.com"),
		WithBoundaryGenerator(sequenceBoundary("mixed-boundary", "related-boundary", "alternative-boundary")),
	)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	raw := buf.String()

	assertContains(t, raw, "From: \"Sender\" <sender@example.com>\r\n")
	assertContains(t, raw, "To: \"Receiver\" <receiver@example.com>\r\n")
	assertContains(t, raw, "Cc: <copy@example.com>\r\n")
	assertContains(t, raw, "Subject: =?UTF-8?b?")
	assertContains(t, raw, "Message-ID: <message@example.com>\r\n")
	assertContains(t, raw, `Content-Type: multipart/mixed; boundary="mixed-boundary"`)
	assertContains(t, raw, `Content-Type: multipart/related; boundary="related-boundary"`)
	assertContains(t, raw, `Content-Type: multipart/alternative; boundary="alternative-boundary"`)
	assertContains(t, raw, "Content-Id: <logo>")
	assertContains(t, raw, "Content-Disposition: attachment;")
	if strings.Contains(raw, "Bcc:") {
		t.Fatalf("rendered message leaked Bcc header:\n%s", raw)
	}

	recipients := strings.Join(msg.Recipients(), ",")
	assertContains(t, recipients, "receiver@example.com")
	assertContains(t, recipients, "copy@example.com")
	assertContains(t, recipients, "hidden@example.com")
}

func TestMessageOptionsAndRenderingPaths(t *testing.T) {
	from := &Address{Name: "Sender", Email: "sender@example.com"}
	msg, err := NewMessage(
		WithFromAddress(from),
		WithTo("to@example.com"),
		WithReplyTo("reply@example.com"),
		WithSubject("html"),
		WithHTML("<strong>Hello</strong>"),
		WithHeader("X-Custom", "a", "b"),
		WithCharset(CharsetASCII),
		WithEncoding(EncodingBase64),
	)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	from.Email = "changed@example.com"
	raw, err := msg.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	text := string(raw)
	assertContains(t, text, `From: "Sender" <sender@example.com>`)
	assertContains(t, text, "Reply-To: <reply@example.com>\r\n")
	assertContains(t, text, "X-Custom: a, b\r\n")
	assertContains(t, text, "Content-Type: text/html; charset=US-ASCII\r\n")
	assertContains(t, text, "PHN0cm9uZz5IZWxsbzwvc3Ryb25nPg==")
}

func TestMessageEncodingAndBoundaryErrors(t *testing.T) {
	for _, tt := range []struct {
		name     string
		encoding Encoding
		body     string
	}{
		{name: "seven bit", encoding: Encoding7Bit, body: "plain"},
		{name: "eight bit", encoding: Encoding8Bit, body: "héllo"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewMessage(
				WithFrom("from@example.com"),
				WithTo("to@example.com"),
				WithText(tt.body),
				WithEncoding(tt.encoding),
			)
			if err != nil {
				t.Fatalf("NewMessage() error = %v", err)
			}
			raw, err := msg.Bytes()
			if err != nil {
				t.Fatalf("Bytes() error = %v", err)
			}
			assertContains(t, string(raw), tt.body)
		})
	}

	msg, err := NewMessage(
		WithFrom("from@example.com"),
		WithTo("to@example.com"),
		WithText("plain"),
		WithHTML("<p>html</p>"),
	)
	if err != nil {
		t.Fatalf("NewMessage(alternative) error = %v", err)
	}
	raw, err := msg.Bytes()
	if err != nil {
		t.Fatalf("Bytes(alternative) error = %v", err)
	}
	assertContains(t, string(raw), "Content-Type: multipart/alternative;")

	msg.Encoding = Encoding("x-bad")
	if _, err := msg.Bytes(); !errors.Is(err, ErrInvalidHeader) {
		t.Fatalf("Bytes(unsupported encoding) error = %v, want %v", err, ErrInvalidHeader)
	}

	badBoundary, err := NewMessage(
		WithFrom("from@example.com"),
		WithTo("to@example.com"),
		WithText("plain"),
		WithHTML("<p>html</p>"),
		WithBoundaryGenerator(func() (string, error) { return "bad\r\nboundary", nil }),
	)
	if err != nil {
		t.Fatalf("NewMessage(bad boundary) error = %v", err)
	}
	if _, err := badBoundary.Bytes(); err == nil {
		t.Fatal("Bytes(bad boundary) error = nil, want error")
	}
}

func TestMessageWriteToPropagatesWriterErrors(t *testing.T) {
	msg, err := NewMessage(
		WithFrom("from@example.com"),
		WithTo("to@example.com"),
		WithText("plain"),
	)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	errWriter := failAfterWriter{limit: 8, err: errors.New("write failed")}
	if _, err := msg.WriteTo(&errWriter); !errors.Is(err, errWriter.err) {
		t.Fatalf("WriteTo() error = %v, want write cause", err)
	}
}

func TestBase64EncoderCloseErrorIsReturned(t *testing.T) {
	errWriter := failAfterWriter{limit: 0, err: errors.New("close padding failed")}
	encoder := base64.NewEncoder(base64.StdEncoding, newBase64LineWriter(&errWriter))
	if _, err := encoder.Write([]byte{0xff}); err != nil {
		t.Fatalf("encoder.Write() error = %v", err)
	}
	if err := encoder.Close(); !errors.Is(err, errWriter.err) {
		t.Fatalf("encoder.Close() error = %v, want close cause", err)
	}
}

type failAfterWriter struct {
	limit int
	err   error
	n     int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.n+len(p) > w.limit {
		return 0, w.err
	}
	w.n += len(p)
	return len(p), nil
}
