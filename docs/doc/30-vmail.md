# vmail Quickstart

`vmail` builds RFC 5322 email messages, renders MIME text/HTML/inline/attachment bodies, and sends them through context-aware SMTP clients with secure TLS defaults. It also provides account-based quick send helpers for applications that keep SMTP defaults in configuration.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Build a message without sending it | `NewMessage` with `WithFrom`, `WithTo`, `WithSubject`, `WithText`, `WithHTML` | Best when the caller needs to inspect bytes, recipients, or render errors before delivery. |
| Parse or construct addresses | `ParseAddress`, `ParseAddressList`, `NewAddress` | Address helpers reject malformed values before they become headers or SMTP envelope data. |
| Add in-memory attachments | `WithAttachment`, `NewAttachment` | Use for small generated files; size limits still apply during message rendering. |
| Add lazily opened content | `WithAttachmentReader`, `WithInlineReader` | Use when content is large, generated on demand, or should be opened only while rendering. |
| Add files from disk | `WithAttachmentFile`, `WithInlineFile` | File helpers stat during construction and open while rendering; validate paths at the application boundary. |
| Send one text or HTML message | `SendText`, `SendHTML` | Convenience wrappers that build and send a simple message through a one-shot client. |
| Send with configured account defaults | `QuickSend`, `SendAccountText`, `SendAccountHTML` | Keeps host, credentials, sender, and TLS policy in an `Account`. |
| Reuse an SMTP connection | `NewClient` + `Dial` + `SendCloser.Send` | Use for batches to avoid reconnecting for each message; close the sender when done. |
| Customize transport for tests | `WithSenderProvider`, `WithDialContext` | Inject fake senders or dialers to keep tests hermetic and offline. |
| Customize SMTP security | `WithTLSPolicy`, `WithTLSConfig`, `WithAllowPlainAuth` | Prefer secure defaults; override only for known server requirements. |

## Mail safety checklist

- Keep `TLSMandatoryStartTLS` or another explicit secure policy unless a trusted SMTP server requires different behavior.
- Do not enable `WithAllowPlainAuth(true)` unless the connection is otherwise protected and the SMTP server is trusted.
- Use `WithEnvelopeFrom` for bounce handling instead of overloading the visible `From` header.
- Validate and normalize recipients before sending; message helpers reject invalid headers but cannot decide business authorization.
- Cap attachments with `WithMaxAttachmentBytes` and avoid loading unbounded user-provided files into memory.
- Inject `WithSenderProvider` or `WithDialContext` in tests so no real SMTP server, credentials, or network calls are used.
- Avoid logging account passwords, SMTP auth values, message bodies, or attachment content on send errors.

## Build a text and HTML message

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmail"
)

func main() {
	msg, err := vmail.NewMessage(
		vmail.WithFrom("Sender <sender@example.com>"),
		vmail.WithTo("Receiver <receiver@example.com>"),
		vmail.WithSubject("hello"),
		vmail.WithText("plain body"),
		vmail.WithHTML("<p>html body</p>"),
	)
	if err != nil {
		panic(err)
	}

	data, err := msg.Bytes()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(data))
}
```

## Add inline files and attachments

```go
package main

import (
	"fmt"
	"os"

	"github.com/imajinyun/go-knifer/vmail"
)

func main() {
	tmp, err := os.CreateTemp("", "report-*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("report body"); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		panic(err)
	}

	msg, err := vmail.NewMessage(
		vmail.WithFrom("sender@example.com"),
		vmail.WithTo("receiver@example.com"),
		vmail.WithSubject("report"),
		vmail.WithHTML(`<p><img src="cid:logo">report attached</p>`),
		vmail.WithInline("logo.png", "logo", []byte("png bytes"), "image/png"),
		vmail.WithAttachment("report.txt", []byte("report body"), vmail.TypeTextPlain),
		vmail.WithAttachmentFile(tmp.Name()),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(msg.Recipients())
}
```

Use `WithAttachmentReader` or `WithInlineReader` when content should be opened lazily by a caller-provided `io.ReadCloser` factory. File helpers stat the target during message construction and open it only while rendering.

## Send with SMTP

```go
package main

import (
	"context"
	"time"

	"github.com/imajinyun/go-knifer/vmail"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := vmail.SendText(
		ctx,
		"smtp.example.com",
		587,
		"sender@example.com",
		[]string{"receiver@example.com"},
		"hello",
		"plain body",
		vmail.WithAuth("sender@example.com", "password"),
	)
	if err != nil {
		panic(err)
	}
}
```

## Send with account defaults

```go
package main

import (
	"context"
	"time"

	"github.com/imajinyun/go-knifer/vmail"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	account := vmail.Account{
		Host:      "smtp.example.com",
		Port:      587,
		Username:  "sender@example.com",
		Password:  "password",
		From:      "sender@example.com",
		FromName:  "Example Sender",
		TLSPolicy: vmail.TLSMandatoryStartTLS,
	}

	err := vmail.SendAccountHTML(
		ctx,
		account,
		[]string{"receiver@example.com"},
		"hello",
		"<p>html body</p>",
		vmail.WithQuickMessageOptions(vmail.WithEnvelopeFrom("bounce@example.com")),
	)
	if err != nil {
		panic(err)
	}
}
```

## Reuse an SMTP connection

```go
package main

import (
	"context"
	"time"

	"github.com/imajinyun/go-knifer/vmail"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := vmail.NewClient(
		"smtp.example.com",
		587,
		vmail.WithAuth("sender@example.com", "password"),
	)
	if err != nil {
		panic(err)
	}

	sender, err := client.Dial(ctx)
	if err != nil {
		panic(err)
	}
	defer sender.Close()

	for _, receiver := range []string{"first@example.com", "second@example.com"} {
		message, err := vmail.NewMessage(
			vmail.WithFrom("sender@example.com"),
			vmail.WithTo(receiver),
			vmail.WithSubject("hello"),
			vmail.WithText("plain body"),
		)
		if err != nil {
			panic(err)
		}
		if err := sender.Send(ctx, message); err != nil {
			panic(err)
		}
	}
}
```

## Security defaults

- `vmail.NewClient` requires STARTTLS by default. Use `WithTLSPolicy` only when the server requires a different policy.
- SMTP AUTH over plaintext is rejected unless `WithAllowPlainAuth(true)` is set explicitly.
- Address and header helpers reject CRLF injection.
- Attachments are size-limited by default; tune with `WithMaxAttachmentBytes`.
- `WithEnvelopeFrom` separates the SMTP MAIL FROM address from the visible `From` header for bounce handling.
- `Client.Dial` returns a `SendCloser` that reuses the SMTP connection and issues `RSET` before each subsequent message.

## When not to use vmail

- Use a provider SDK when delivery depends on vendor APIs, templates, suppression lists, analytics, or webhooks instead of raw SMTP.
- Use a queue or background worker when messages must be retried, rate limited, or sent outside the request path.
- Use a dedicated MIME library when you need nonstandard multipart structures not covered by the facade options.
- Do not use quick-send helpers for tests or libraries that must not perform network I/O; inject a sender provider or build messages only.

## Related packages

- Use `vtpl` when email bodies need reusable HTML or text templates.
- Use `vfile` when attachments or generated message files require path and filesystem policy checks.
- Use `vconf` when SMTP accounts, TLS policy, or sender defaults come from layered configuration.

## Benchmarks and trade-offs

- One-shot `SendText` and `SendHTML` are concise but reconnect for each message. Reusing `Client.Dial` reduces SMTP handshake overhead for batches.
- In-memory attachments are simple but require the full byte slice up front. Reader and file helpers defer opening content until render time.
- Mandatory STARTTLS protects credentials by default but can fail against legacy servers; relaxing TLS policy is a compatibility trade-off that should be visible in configuration.
- MIME rendering validates headers and boundaries each time bytes are produced. Fixed `WithBoundaryGenerator` values improve deterministic tests.
- Account quick helpers reduce call-site configuration but can hide per-message differences; use explicit `NewMessage` options for unusual envelopes.

## FAQ

### Why does SMTP AUTH fail on plaintext connections?

The default client rejects plaintext authentication to avoid sending credentials without transport protection. Use STARTTLS or TLS, and only set `WithAllowPlainAuth(true)` for a trusted, protected environment.

### How do I test sending without SMTP?

Pass `WithSenderProvider` to return a fake `Sender` or `SendCloser`, or pass `WithDialContext` for lower-level dial tests. This keeps unit tests offline and deterministic.

### When should I use `WithEnvelopeFrom`?

Use it when the SMTP `MAIL FROM` address for bounces differs from the visible `From` header, such as a bounce mailbox or VERP-style return path.

### Should attachments be passed as bytes or readers?

Use byte attachments for small generated content. Use reader or file helpers when content is large, expensive to create, or should be opened only while rendering.
