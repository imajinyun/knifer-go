# vmail Quickstart

`vmail` builds RFC 5322 email messages, renders MIME text/HTML/inline/attachment bodies, and sends them through context-aware SMTP clients with secure TLS defaults.

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

	"github.com/imajinyun/go-knifer/vmail"
)

func main() {
	msg, err := vmail.NewMessage(
		vmail.WithFrom("sender@example.com"),
		vmail.WithTo("receiver@example.com"),
		vmail.WithSubject("report"),
		vmail.WithHTML(`<p><img src="cid:logo">report attached</p>`),
		vmail.WithInline("logo.png", "logo", []byte("png bytes"), "image/png"),
		vmail.WithAttachment("report.txt", []byte("report body"), vmail.TypeTextPlain),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(msg.Recipients())
}
```

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

## Security defaults

- `vmail.NewClient` requires STARTTLS by default. Use `WithTLSPolicy` only when the server requires a different policy.
- SMTP AUTH over plaintext is rejected unless `WithAllowPlainAuth(true)` is set explicitly.
- Address and header helpers reject CRLF injection.
- Attachments are size-limited by default; tune with `WithMaxAttachmentBytes`.
