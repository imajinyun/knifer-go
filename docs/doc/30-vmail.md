# vmail Quickstart

`vmail` builds RFC 5322 email messages, renders MIME text/HTML/inline/attachment bodies, and sends them through context-aware SMTP clients with secure TLS defaults. It also provides account-based quick send helpers for applications that keep SMTP defaults in configuration.

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
