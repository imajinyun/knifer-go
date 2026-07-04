package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"sync"
	"time"
)

// TLSPolicy controls SMTP transport security.
type TLSPolicy int

const (
	// TLSPolicyUnknown uses the package default, currently mandatory STARTTLS.
	TLSPolicyUnknown TLSPolicy = iota
	// TLSMandatoryStartTLS requires STARTTLS before SMTP AUTH or DATA.
	TLSMandatoryStartTLS
	// TLSImplicit uses implicit TLS from the initial connection.
	TLSImplicit
	// TLSOpportunisticStartTLS upgrades when STARTTLS is advertised.
	TLSOpportunisticStartTLS
	// TLSNone disables TLS. AUTH remains disabled unless AllowPlainAuth is set.
	TLSNone
)

// DialContextFunc dials an SMTP server.
type DialContextFunc func(context.Context, string, string) (net.Conn, error)

// ClientOption customizes Client construction.
type ClientOption func(*Client) error

// Sender is implemented by SMTP send backends.
type Sender interface {
	Send(ctx context.Context, message *Message) error
}

// SendCloser sends multiple messages through a reusable SMTP connection.
type SendCloser interface {
	Sender
	Close() error
}

// SenderFunc adapts a function into Sender.
type SenderFunc func(context.Context, *Message) error

// Send sends message.
func (f SenderFunc) Send(ctx context.Context, message *Message) error { return f(ctx, message) }

// SenderProvider creates a sender for a client configuration.
type SenderProvider func(Config) (Sender, error)

type senderDialer interface {
	Dial(ctx context.Context) (SendCloser, error)
}

// Config configures SMTP delivery.
type Config struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Auth           smtp.Auth
	LocalName      string
	TLSConfig      *tls.Config
	TLSPolicy      TLSPolicy
	AllowPlainAuth bool
	Timeout        time.Duration
	DialContext    DialContextFunc
}

// Client sends messages through SMTP.
type Client struct {
	config         Config
	senderProvider SenderProvider
}

// NewClient creates an SMTP client.
func NewClient(host string, port int, opts ...ClientOption) (*Client, error) {
	c := &Client{
		config: Config{
			Host:        host,
			Port:        port,
			LocalName:   "localhost",
			TLSPolicy:   TLSMandatoryStartTLS,
			Timeout:     10 * time.Second,
			DialContext: (&net.Dialer{}).DialContext,
		},
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.config.Host == "" || c.config.Port <= 0 {
		return nil, fmt.Errorf("%w: invalid smtp address", ErrInvalidAddress)
	}
	if c.config.TLSPolicy == TLSPolicyUnknown {
		c.config.TLSPolicy = TLSMandatoryStartTLS
	}
	if c.config.DialContext == nil {
		c.config.DialContext = (&net.Dialer{}).DialContext
	}
	if c.senderProvider == nil {
		c.senderProvider = func(config Config) (Sender, error) { return smtpSender{config: config}, nil }
	}
	return c, nil
}

// Send sends message through an SMTP server created from host, port, and options.
func Send(ctx context.Context, host string, port int, message *Message, opts ...ClientOption) error {
	client, err := NewClient(host, port, opts...)
	if err != nil {
		return err
	}
	return client.Send(ctx, message)
}

// SendText creates and sends a plain text message.
func SendText(ctx context.Context, host string, port int, from string, to []string, subject, text string, opts ...ClientOption) error {
	msgOpts := make([]MessageOption, 0, 4)
	msgOpts = append(msgOpts, WithFrom(from), WithSubject(subject), WithText(text), WithTo(to...))
	message, err := NewMessage(msgOpts...)
	if err != nil {
		return err
	}
	return Send(ctx, host, port, message, opts...)
}

// SendHTML creates and sends an HTML message.
func SendHTML(ctx context.Context, host string, port int, from string, to []string, subject, html string, opts ...ClientOption) error {
	msgOpts := make([]MessageOption, 0, 4)
	msgOpts = append(msgOpts, WithFrom(from), WithSubject(subject), WithHTML(html), WithTo(to...))
	message, err := NewMessage(msgOpts...)
	if err != nil {
		return err
	}
	return Send(ctx, host, port, message, opts...)
}

// Send sends message with context cancellation and configured SMTP security.
func (c *Client) Send(ctx context.Context, message *Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if message == nil {
		return ErrMissingBody
	}
	if err := message.Validate(); err != nil {
		return err
	}
	sender, err := c.senderProvider(c.config)
	if err != nil {
		return wrapProviderError("mail: create sender failed", err)
	}
	if err := sender.Send(ctx, message); err != nil {
		return wrapProviderError("mail: send message failed", err)
	}
	return nil
}

// Dial opens a reusable SMTP connection for sending multiple messages.
func (c *Client) Dial(ctx context.Context) (SendCloser, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	sender, err := c.senderProvider(c.config)
	if err != nil {
		return nil, wrapProviderError("mail: create sender failed", err)
	}
	if sendCloser, ok := sender.(SendCloser); ok {
		return sendCloser, nil
	}
	dialer, ok := sender.(senderDialer)
	if !ok {
		return nil, wrapProviderError("mail: sender does not support dial", errors.New("mail: sender does not support Dial"))
	}
	sendCloser, err := dialer.Dial(ctx)
	if err != nil {
		return nil, wrapProviderError("mail: dial sender failed", err)
	}
	return sendCloser, nil
}

// WithAuth sets SMTP username and password.
func WithAuth(username, password string) ClientOption {
	return func(c *Client) error {
		c.config.Username = username
		c.config.Password = password
		return nil
	}
}

// WithSMTPAuth sets a custom SMTP authentication mechanism.
func WithSMTPAuth(auth smtp.Auth) ClientOption {
	return func(c *Client) error {
		c.config.Auth = auth
		return nil
	}
}

// WithTLSConfig sets the TLS configuration. The value is cloned.
func WithTLSConfig(config *tls.Config) ClientOption {
	return func(c *Client) error {
		if config == nil {
			c.config.TLSConfig = nil
			return nil
		}
		c.config.TLSConfig = config.Clone()
		return nil
	}
}

// WithTLSPolicy sets SMTP TLS behavior.
func WithTLSPolicy(policy TLSPolicy) ClientOption {
	return func(c *Client) error {
		c.config.TLSPolicy = policy
		return nil
	}
}

// WithAllowPlainAuth permits SMTP AUTH without TLS. Prefer TLS instead.
func WithAllowPlainAuth(allow bool) ClientOption {
	return func(c *Client) error {
		c.config.AllowPlainAuth = allow
		return nil
	}
}

// WithTimeout sets a client-wide operation timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		c.config.Timeout = timeout
		return nil
	}
}

// WithLocalName sets the HELO/EHLO local name.
func WithLocalName(name string) ClientOption {
	return func(c *Client) error {
		if hasCRLF(name) || name == "" {
			return ErrInvalidHeader
		}
		c.config.LocalName = name
		return nil
	}
}

// WithDialContext sets the network dialer.
func WithDialContext(dial DialContextFunc) ClientOption {
	return func(c *Client) error {
		if dial == nil {
			return errors.New("mail: nil dialer")
		}
		c.config.DialContext = dial
		return nil
	}
}

// WithSenderProvider sets a custom sender provider, primarily for deterministic tests.
func WithSenderProvider(provider SenderProvider) ClientOption {
	return func(c *Client) error {
		if provider == nil {
			return errors.New("mail: nil sender provider")
		}
		c.senderProvider = provider
		return nil
	}
}

type smtpSender struct{ config Config }

func (s smtpSender) Send(ctx context.Context, message *Message) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := withClientTimeout(ctx, s.config.Timeout)
	defer cancel()
	sendCloser, err := s.Dial(ctx)
	if err != nil {
		return wrapProviderError("mail: open smtp connection failed", err)
	}
	defer func() {
		if closeErr := sendCloser.Close(); err == nil && closeErr != nil {
			err = wrapProviderError("mail: close smtp connection failed", closeErr)
		}
	}()
	if err := sendCloser.Send(ctx, message); err != nil {
		return wrapProviderError("mail: send smtp message failed", err)
	}
	return nil
}

func (s smtpSender) Dial(ctx context.Context) (sendCloser SendCloser, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := withClientTimeout(ctx, s.config.Timeout)
	defer cancel()
	addr := net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))
	tlsConfig := s.tlsConfig()
	conn, err := s.dial(ctx, addr, tlsConfig)
	if err != nil {
		return nil, err
	}
	defer closeOnError(&err, conn)
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return nil, smtpContextError(ctx, "create smtp client", err)
	}
	defer closeClientOnError(&err, client)
	if s.config.LocalName != "" {
		if err := client.Hello(s.config.LocalName); err != nil {
			return nil, smtpContextError(ctx, "smtp hello", err)
		}
	}
	isTLS := s.config.TLSPolicy == TLSImplicit
	if s.config.TLSPolicy == TLSMandatoryStartTLS || s.config.TLSPolicy == TLSOpportunisticStartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return nil, smtpContextError(ctx, "smtp starttls", err)
			}
			isTLS = true
		} else if s.config.TLSPolicy == TLSMandatoryStartTLS {
			return nil, ErrTLSRequired
		}
	}
	if auth := s.auth(); auth != nil {
		if !isTLS && !s.config.AllowPlainAuth {
			return nil, ErrPlainAuth
		}
		if err := client.Auth(auth); err != nil {
			return nil, smtpContextError(ctx, "smtp auth", err)
		}
	}
	if !stop() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	return &smtpSendCloser{config: s.config, conn: conn, client: client}, nil
}

type smtpSendCloser struct {
	config Config
	conn   net.Conn
	client *smtp.Client

	mu     sync.Mutex
	closed bool
	sent   bool
}

func (s *smtpSendCloser) Send(ctx context.Context, message *Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if message == nil {
		return ErrMissingBody
	}
	if err := message.Validate(); err != nil {
		return err
	}
	ctx, cancel := withClientTimeout(ctx, s.config.Timeout)
	defer cancel()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return wrapProviderError("mail: smtp sender is closed", errors.New("mail: smtp sender is closed"))
	}
	stop := context.AfterFunc(ctx, func() { _ = s.conn.Close() })
	defer stop()
	defer func() {
		if !s.closed {
			_ = s.conn.SetDeadline(time.Time{})
		}
	}()
	if deadline, ok := ctx.Deadline(); ok {
		_ = s.conn.SetDeadline(deadline)
	}
	if s.sent {
		if err := s.client.Reset(); err != nil {
			return s.smtpError(ctx, "smtp reset", err)
		}
	}
	if err := s.client.Mail(message.Sender()); err != nil {
		return s.smtpError(ctx, "smtp mail from", err)
	}
	for _, recipient := range message.Recipients() {
		if err := s.client.Rcpt(recipient); err != nil {
			return s.smtpError(ctx, fmt.Sprintf("smtp rcpt to %q", recipient), err)
		}
	}
	w, err := s.client.Data()
	if err != nil {
		return s.smtpError(ctx, "smtp data", err)
	}
	if _, err := message.WriteTo(w); err != nil {
		_ = w.Close()
		return s.smtpError(ctx, "smtp write data", err)
	}
	if err := w.Close(); err != nil {
		return s.smtpError(ctx, "smtp close data", err)
	}
	if !stop() {
		if err := ctx.Err(); err != nil {
			s.closed = true
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		s.closed = true
		return err
	}
	s.sent = true
	return nil
}

func (s *smtpSendCloser) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if err := s.client.Quit(); err != nil {
		_ = s.client.Close()
		return wrapProviderError("smtp quit", err)
	}
	return nil
}

func (s *smtpSendCloser) smtpError(ctx context.Context, operation string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		s.closed = true
		return ctxErr
	}
	return wrapProviderError(operation, err)
}

func (s smtpSender) auth() smtp.Auth {
	if s.config.Auth != nil {
		return s.config.Auth
	}
	if s.config.Username == "" {
		return nil
	}
	return smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
}

func (s smtpSender) dial(ctx context.Context, addr string, tlsConfig *tls.Config) (net.Conn, error) {
	conn, err := s.config.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, wrapProviderError("smtp dial", err)
	}
	if s.config.TLSPolicy == TLSImplicit {
		tlsConn := tls.Client(conn, tlsConfig)
		if deadline, ok := ctx.Deadline(); ok {
			_ = tlsConn.SetDeadline(deadline)
		}
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return nil, wrapProviderError("smtp tls handshake", err)
		}
		return tlsConn, nil
	}
	return conn, nil
}

func (s smtpSender) tlsConfig() *tls.Config {
	if s.config.TLSConfig != nil {
		config := s.config.TLSConfig.Clone()
		if config.ServerName == "" {
			config.ServerName = s.config.Host
		}
		return config
	}
	return &tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12}
}

func closeOnError(err *error, conn net.Conn) {
	if *err != nil {
		_ = conn.Close()
	}
}

func closeClientOnError(err *error, client *smtp.Client) {
	if *err != nil {
		_ = client.Close()
	}
}

func withClientTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func smtpContextError(ctx context.Context, operation string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return wrapProviderError(operation, err)
}
