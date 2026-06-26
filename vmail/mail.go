package vmail

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/smtp"
	"time"

	"github.com/imajinyun/knifer-go/internal/mail"
)

// Address is an RFC 5322 mailbox address.
type Address = mail.Address

// Account stores SMTP server, authentication, and default sender settings.
type Account = mail.Account

// Attachment is a MIME file part.
type Attachment = mail.Attachment

// BoundaryGenerator creates MIME multipart boundaries.
type BoundaryGenerator = mail.BoundaryGenerator

// Charset identifies a MIME character set.
type Charset = mail.Charset

// Client sends messages through SMTP.
type Client = mail.Client

// ClientOption customizes Client construction.
type ClientOption = mail.ClientOption

// Config configures SMTP delivery.
type Config = mail.Config

// ContentType identifies a MIME media type.
type ContentType = mail.ContentType

// DialContextFunc dials an SMTP server.
type DialContextFunc = mail.DialContextFunc

// Encoding identifies a Content-Transfer-Encoding value.
type Encoding = mail.Encoding

// Header stores message headers in insertion order.
type Header = mail.Header

// Message is an email message with text, HTML, inline files, and attachments.
type Message = mail.Message

// MessageOption customizes message construction.
type MessageOption = mail.MessageOption

// QuickOption customizes account-based quick send helpers.
type QuickOption = mail.QuickOption

// Sender is implemented by SMTP send backends.
type Sender = mail.Sender

// SendCloser sends multiple messages through a reusable SMTP connection.
type SendCloser = mail.SendCloser

// SenderFunc adapts a function into Sender.
type SenderFunc = mail.SenderFunc

// SenderProvider creates a sender for a client configuration.
type SenderProvider = mail.SenderProvider

// TLSPolicy controls SMTP transport security.
type TLSPolicy = mail.TLSPolicy

const (
	CharsetUTF8  Charset = mail.CharsetUTF8
	CharsetASCII Charset = mail.CharsetASCII

	EncodingBase64          Encoding = mail.EncodingBase64
	EncodingQuotedPrintable Encoding = mail.EncodingQuotedPrintable
	Encoding7Bit            Encoding = mail.Encoding7Bit
	Encoding8Bit            Encoding = mail.Encoding8Bit

	TypeTextPlain              ContentType = mail.TypeTextPlain
	TypeTextHTML               ContentType = mail.TypeTextHTML
	TypeApplicationOctetStream ContentType = mail.TypeApplicationOctetStream

	TLSPolicyUnknown         TLSPolicy = mail.TLSPolicyUnknown
	TLSMandatoryStartTLS     TLSPolicy = mail.TLSMandatoryStartTLS
	TLSImplicit              TLSPolicy = mail.TLSImplicit
	TLSOpportunisticStartTLS TLSPolicy = mail.TLSOpportunisticStartTLS
	TLSNone                  TLSPolicy = mail.TLSNone
)

var (
	ErrInvalidAddress     = mail.ErrInvalidAddress
	ErrInvalidHeader      = mail.ErrInvalidHeader
	ErrMissingFrom        = mail.ErrMissingFrom
	ErrMissingRecipient   = mail.ErrMissingRecipient
	ErrMissingBody        = mail.ErrMissingBody
	ErrTLSRequired        = mail.ErrTLSRequired
	ErrPlainAuth          = mail.ErrPlainAuth
	ErrAttachmentTooLarge = mail.ErrAttachmentTooLarge
)

// NewAddress validates and returns a mailbox address.
func NewAddress(name, email string) (*Address, error) { return mail.NewAddress(name, email) }

// ParseAddress parses a single mailbox address.
func ParseAddress(value string) (*Address, error) { return mail.ParseAddress(value) }

// ParseAddressList parses a comma-separated mailbox address list.
func ParseAddressList(value string) ([]*Address, error) { return mail.ParseAddressList(value) }

// NewAttachment creates an attachment from bytes.
func NewAttachment(name string, content []byte, contentType ContentType) (Attachment, error) {
	return mail.NewAttachment(name, content, contentType)
}

// NewAttachmentReader creates an attachment from a reader opener.
func NewAttachmentReader(name string, size int64, contentType ContentType, open func() (io.ReadCloser, error)) (Attachment, error) {
	return mail.NewAttachmentReader(name, size, contentType, open)
}

// NewAttachmentFile creates an attachment loaded lazily from path.
func NewAttachmentFile(path string) (Attachment, error) { return mail.NewAttachmentFile(path) }

// NewInline creates an inline attachment from bytes with a Content-ID.
func NewInline(name, contentID string, content []byte, contentType ContentType) (Attachment, error) {
	return mail.NewInline(name, contentID, content, contentType)
}

// NewInlineReader creates an inline attachment from a reader opener with a Content-ID.
func NewInlineReader(
	name string,
	contentID string,
	size int64,
	contentType ContentType,
	open func() (io.ReadCloser, error),
) (Attachment, error) {
	return mail.NewInlineReader(name, contentID, size, contentType, open)
}

// NewInlineFile creates an inline attachment loaded lazily from path with a Content-ID.
func NewInlineFile(path, contentID string) (Attachment, error) {
	return mail.NewInlineFile(path, contentID)
}

// NewMessage creates and validates an email message.
func NewMessage(opts ...MessageOption) (*Message, error) { return mail.NewMessage(opts...) }

// NewClient creates an SMTP client.
func NewClient(host string, port int, opts ...ClientOption) (*Client, error) {
	return mail.NewClient(host, port, opts...)
}

// Send sends message through an SMTP server created from host, port, and options.
func Send(ctx context.Context, host string, port int, message *Message, opts ...ClientOption) error {
	return mail.Send(ctx, host, port, message, opts...)
}

// SendText creates and sends a plain text message.
func SendText(ctx context.Context, host string, port int, from string, to []string, subject, text string, opts ...ClientOption) error {
	return mail.SendText(ctx, host, port, from, to, subject, text, opts...)
}

// SendHTML creates and sends an HTML message.
func SendHTML(ctx context.Context, host string, port int, from string, to []string, subject, html string, opts ...ClientOption) error {
	return mail.SendHTML(ctx, host, port, from, to, subject, html, opts...)
}

// QuickSend creates and sends a message using account defaults plus quick options.
func QuickSend(ctx context.Context, account Account, opts ...QuickOption) error {
	return mail.QuickSend(ctx, account, opts...)
}

// SendAccountText creates and sends a plain text message using account defaults.
func SendAccountText(
	ctx context.Context,
	account Account,
	to []string,
	subject string,
	text string,
	opts ...QuickOption,
) error {
	return mail.SendAccountText(ctx, account, to, subject, text, opts...)
}

// SendAccountHTML creates and sends an HTML message using account defaults.
func SendAccountHTML(
	ctx context.Context,
	account Account,
	to []string,
	subject string,
	html string,
	opts ...QuickOption,
) error {
	return mail.SendAccountHTML(ctx, account, to, subject, html, opts...)
}

// WithFrom sets the From address.
func WithFrom(address string) MessageOption { return mail.WithFrom(address) }

// WithFromAddress sets the From address.
func WithFromAddress(addr *Address) MessageOption { return mail.WithFromAddress(addr) }

// WithEnvelopeFrom sets the SMTP envelope sender for MAIL FROM.
func WithEnvelopeFrom(address string) MessageOption { return mail.WithEnvelopeFrom(address) }

// WithTo appends To recipients.
func WithTo(addresses ...string) MessageOption { return mail.WithTo(addresses...) }

// WithCc appends Cc recipients.
func WithCc(addresses ...string) MessageOption { return mail.WithCc(addresses...) }

// WithBcc appends Bcc recipients.
func WithBcc(addresses ...string) MessageOption { return mail.WithBcc(addresses...) }

// WithReplyTo appends Reply-To addresses.
func WithReplyTo(addresses ...string) MessageOption { return mail.WithReplyTo(addresses...) }

// WithSubject sets the Subject header.
func WithSubject(subject string) MessageOption { return mail.WithSubject(subject) }

// WithText sets the text/plain body.
func WithText(text string) MessageOption { return mail.WithText(text) }

// WithHTML sets the text/html body.
func WithHTML(html string) MessageOption { return mail.WithHTML(html) }

// WithHeader appends an additional header.
func WithHeader(name string, values ...string) MessageOption { return mail.WithHeader(name, values...) }

// WithAttachment appends an attachment from bytes.
func WithAttachment(name string, content []byte, contentType ContentType) MessageOption {
	return mail.WithAttachment(name, content, contentType)
}

// WithAttachmentReader appends an attachment from a reader opener.
func WithAttachmentReader(name string, size int64, contentType ContentType, open func() (io.ReadCloser, error)) MessageOption {
	return mail.WithAttachmentReader(name, size, contentType, open)
}

// WithInline appends an inline file from bytes.
func WithInline(name, contentID string, content []byte, contentType ContentType) MessageOption {
	return mail.WithInline(name, contentID, content, contentType)
}

// WithInlineReader appends an inline file from a reader opener with a Content-ID.
func WithInlineReader(
	name string,
	contentID string,
	size int64,
	contentType ContentType,
	open func() (io.ReadCloser, error),
) MessageOption {
	return mail.WithInlineReader(name, contentID, size, contentType, open)
}

// WithAttachmentFile appends an attachment loaded lazily from path.
func WithAttachmentFile(path string) MessageOption { return mail.WithAttachmentFile(path) }

// WithInlineFile appends an inline attachment loaded lazily from path with a Content-ID.
func WithInlineFile(path, contentID string) MessageOption {
	return mail.WithInlineFile(path, contentID)
}

// WithDate sets the Date header value.
func WithDate(t time.Time) MessageOption { return mail.WithDate(t) }

// WithMessageID sets the Message-ID header without angle brackets.
func WithMessageID(id string) MessageOption { return mail.WithMessageID(id) }

// WithCharset sets the charset used for message text parts and encoded headers.
func WithCharset(charset Charset) MessageOption { return mail.WithCharset(charset) }

// WithEncoding sets the transfer encoding used for text parts.
func WithEncoding(encoding Encoding) MessageOption { return mail.WithEncoding(encoding) }

// WithMaxAttachmentBytes sets the per-attachment size limit. A non-positive value disables the limit.
func WithMaxAttachmentBytes(maxBytes int64) MessageOption {
	return mail.WithMaxAttachmentBytes(maxBytes)
}

// WithBoundaryGenerator injects the MIME boundary generator.
func WithBoundaryGenerator(generator BoundaryGenerator) MessageOption {
	return mail.WithBoundaryGenerator(generator)
}

// WithQuickMessageOptions appends message options used by QuickSend and account helpers.
func WithQuickMessageOptions(opts ...MessageOption) QuickOption {
	return mail.WithQuickMessageOptions(opts...)
}

// WithQuickClientOptions appends client options used by QuickSend and account helpers.
func WithQuickClientOptions(opts ...ClientOption) QuickOption {
	return mail.WithQuickClientOptions(opts...)
}

// WithAuth sets SMTP username and password.
func WithAuth(username, password string) ClientOption { return mail.WithAuth(username, password) }

// WithSMTPAuth sets a custom SMTP authentication mechanism.
func WithSMTPAuth(auth smtp.Auth) ClientOption { return mail.WithSMTPAuth(auth) }

// WithTLSConfig sets the TLS configuration. The value is cloned.
func WithTLSConfig(config *tls.Config) ClientOption { return mail.WithTLSConfig(config) }

// WithTLSPolicy sets SMTP TLS behavior.
func WithTLSPolicy(policy TLSPolicy) ClientOption { return mail.WithTLSPolicy(policy) }

// WithAllowPlainAuth permits SMTP AUTH without TLS. Prefer TLS instead.
func WithAllowPlainAuth(allow bool) ClientOption { return mail.WithAllowPlainAuth(allow) }

// WithTimeout sets a client-wide operation timeout.
func WithTimeout(timeout time.Duration) ClientOption { return mail.WithTimeout(timeout) }

// WithLocalName sets the HELO/EHLO local name.
func WithLocalName(name string) ClientOption { return mail.WithLocalName(name) }

// WithDialContext sets the network dialer.
func WithDialContext(dial func(context.Context, string, string) (net.Conn, error)) ClientOption {
	return mail.WithDialContext(dial)
}

// WithSenderProvider sets a custom sender provider, primarily for deterministic tests.
func WithSenderProvider(provider SenderProvider) ClientOption {
	return mail.WithSenderProvider(provider)
}
