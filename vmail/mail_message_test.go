package vmail

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewMessageFacade(t *testing.T) {
	message, err := NewMessage(
		WithFrom("from@example.com"),
		WithTo("to@example.com"),
		WithSubject("hello"),
		WithText("body"),
	)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	raw, err := message.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	if !strings.Contains(string(raw), "Subject: hello") {
		t.Fatalf("message did not contain encoded subject: %s", raw)
	}
}

func TestFacadeConstructorsAndMessageOptions(t *testing.T) {
	addr, err := NewAddress("Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("NewAddress() error = %v", err)
	}
	if _, err := ParseAddress(addr.String()); err != nil {
		t.Fatalf("ParseAddress() error = %v", err)
	}
	list, err := ParseAddressList("bob@example.com, carol@example.com")
	if err != nil {
		t.Fatalf("ParseAddressList() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ParseAddressList() len = %d, want 2", len(list))
	}
	attachment, err := NewAttachment("report.txt", []byte("report"), TypeTextPlain)
	if err != nil {
		t.Fatalf("NewAttachment() error = %v", err)
	}
	readerAttachment, err := NewAttachmentReader("reader.txt", 6, TypeTextPlain, func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("reader")), nil
	})
	if err != nil {
		t.Fatalf("NewAttachmentReader() error = %v", err)
	}
	inline, err := NewInline("logo.png", "logo", []byte("inline"), "")
	if err != nil {
		t.Fatalf("NewInline() error = %v", err)
	}
	readerInline, err := NewInlineReader("icon.png", "icon", 4, "image/png", func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("icon")), nil
	})
	if err != nil {
		t.Fatalf("NewInlineReader() error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "extra.txt")
	if err := os.WriteFile(path, []byte("extra"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	fileAttachment, err := NewAttachmentFile(path)
	if err != nil {
		t.Fatalf("NewAttachmentFile() error = %v", err)
	}
	inlinePath := filepath.Join(dir, "inline.png")
	if err := os.WriteFile(inlinePath, []byte("inline-file"), 0o600); err != nil {
		t.Fatalf("WriteFile(inline) error = %v", err)
	}
	fileInline, err := NewInlineFile(inlinePath, "inline-file")
	if err != nil {
		t.Fatalf("NewInlineFile() error = %v", err)
	}
	message, err := NewMessage(
		WithFromAddress(addr),
		WithEnvelopeFrom("bounce@example.com"),
		WithTo("to@example.com"),
		WithCc("cc@example.com"),
		WithBcc("bcc@example.com"),
		WithReplyTo("reply@example.com"),
		WithSubject("facade"),
		WithText("plain"),
		WithHTML("<b>html</b>"),
		WithHeader("X-Facade", "yes"),
		WithAttachment(attachment.Name, []byte("report"), attachment.ContentType),
		WithAttachmentReader(readerAttachment.Name, readerAttachment.Size, readerAttachment.ContentType, readerAttachment.Open),
		WithInline(inline.Name, inline.ContentID, []byte("inline"), inline.ContentType),
		WithInlineReader(readerInline.Name, readerInline.ContentID, readerInline.Size, readerInline.ContentType, readerInline.Open),
		WithAttachmentFile(path),
		WithAttachment(fileAttachment.Name, []byte("extra"), fileAttachment.ContentType),
		WithInlineFile(inlinePath, fileInline.ContentID),
		WithDate(time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)),
		WithMessageID("facade@example.com"),
		WithCharset(CharsetUTF8),
		WithEncoding(EncodingQuotedPrintable),
		WithMaxAttachmentBytes(1024),
		WithBoundaryGenerator(sequenceBoundary("mixed", "related", "alternative")),
	)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	raw, err := message.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	text := string(raw)
	assertContains(t, text, "X-Facade: yes")
	assertContains(t, text, "Message-ID: <facade@example.com>")
	assertContains(t, text, `Content-Type: multipart/mixed; boundary="mixed"`)
	assertContains(t, text, `Content-Disposition: attachment; filename=report.txt`)
	if message.Sender() != "bounce@example.com" {
		t.Fatalf("Sender() = %q, want envelope sender", message.Sender())
	}
}
