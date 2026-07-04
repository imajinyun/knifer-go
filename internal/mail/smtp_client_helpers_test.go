package mail

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeSMTPServer struct {
	listener      net.Listener
	done          chan error
	dataStarted   chan struct{}
	mu            sync.Mutex
	data          string
	mailFrom      string
	rcptTo        []string
	rsetCount     int
	cert          *testCertificate
	startTLS      bool
	implicitTLS   bool
	tlsActive     bool
	authMechanism string
	authInitial   string
	authOK        bool
	authenticated bool
	hangOnData    bool
	quitResponse  string
	once          sync.Once
}

type fakeSMTPOption func(*fakeSMTPServer)

func newFakeSMTPServer(t *testing.T, opts ...fakeSMTPOption) (*fakeSMTPServer, error) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	server := &fakeSMTPServer{
		listener:     listener,
		done:         make(chan error, 1),
		dataStarted:  make(chan struct{}),
		authOK:       true,
		quitResponse: "221 Bye",
	}
	for _, opt := range opts {
		opt(server)
	}
	go server.serve()
	return server, nil
}

func withFakeSMTPStartTLS(cert *testCertificate) fakeSMTPOption {
	return func(s *fakeSMTPServer) {
		s.cert = cert
		s.startTLS = true
	}
}

func withFakeSMTPImplicitTLS(cert *testCertificate) fakeSMTPOption {
	return func(s *fakeSMTPServer) {
		s.cert = cert
		s.implicitTLS = true
	}
}

func withFakeSMTPAuth(mechanism, initial string, ok bool) fakeSMTPOption {
	return func(s *fakeSMTPServer) {
		s.authMechanism = mechanism
		s.authInitial = initial
		s.authOK = ok
	}
}

func withFakeSMTPHangOnData() fakeSMTPOption {
	return func(s *fakeSMTPServer) { s.hangOnData = true }
}

func withFakeSMTPQuitResponse(response string) fakeSMTPOption {
	return func(s *fakeSMTPServer) { s.quitResponse = response }
}

func (s *fakeSMTPServer) Host() string {
	host, _, _ := net.SplitHostPort(s.listener.Addr().String())
	return host
}

func (s *fakeSMTPServer) Port() int {
	_, port, _ := net.SplitHostPort(s.listener.Addr().String())
	n, _ := strconvAtoi(port)
	return n
}

func (s *fakeSMTPServer) Close() {
	_ = s.listener.Close()
}

func (s *fakeSMTPServer) Data() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (s *fakeSMTPServer) MailFrom() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mailFrom
}

func (s *fakeSMTPServer) RcptTo() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.rcptTo...)
}

func (s *fakeSMTPServer) RSETCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rsetCount
}

func (s *fakeSMTPServer) TLSActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tlsActive
}

func (s *fakeSMTPServer) Authenticated() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.authenticated
}

func (s *fakeSMTPServer) WaitForDataCommand(t *testing.T) {
	t.Helper()
	select {
	case <-s.dataStarted:
	case <-time.After(time.Second):
		t.Fatal("fake smtp server did not receive DATA command")
	}
}

func (s *fakeSMTPServer) Wait() error {
	select {
	case err := <-s.done:
		return err
	case <-time.After(2 * time.Second):
		return errors.New("fake smtp server timed out")
	}
}

func (s *fakeSMTPServer) serve() {
	conn, err := s.listener.Accept()
	if err != nil {
		s.done <- err
		return
	}
	defer func() { _ = conn.Close() }()
	if s.implicitTLS {
		conn = tls.Server(conn, s.cert.serverConfig())
		if err := conn.(*tls.Conn).Handshake(); err != nil {
			s.done <- err
			return
		}
		s.setTLSActive()
	}
	reader := bufio.NewReader(conn)
	if _, err := io.WriteString(conn, "220 fake.smtp ESMTP\r\n"); err != nil {
		s.done <- err
		return
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.done <- nil
				return
			}
			s.done <- err
			return
		}
		line = strings.TrimRight(line, "\r\n")
		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			if _, err := io.WriteString(conn, s.ehloResponse()); err != nil {
				s.done <- err
				return
			}
		case line == "STARTTLS" && s.startTLS:
			if _, err := io.WriteString(conn, "220 Ready to start TLS\r\n"); err != nil {
				s.done <- err
				return
			}
			conn = tls.Server(conn, s.cert.serverConfig())
			if err := conn.(*tls.Conn).Handshake(); err != nil {
				s.done <- err
				return
			}
			s.setTLSActive()
			reader = bufio.NewReader(conn)
		case strings.HasPrefix(line, "AUTH "):
			if err := s.handleAuth(conn, line); err != nil {
				s.done <- err
				return
			}
		case line == "*":
			s.done <- nil
			return
		case strings.HasPrefix(line, "MAIL FROM:"):
			s.mu.Lock()
			s.mailFrom = strings.TrimPrefix(line, "MAIL FROM:")
			s.mu.Unlock()
			if _, err := io.WriteString(conn, "250 OK\r\n"); err != nil {
				s.done <- err
				return
			}
		case strings.HasPrefix(line, "RCPT TO:"):
			s.mu.Lock()
			s.rcptTo = append(s.rcptTo, strings.TrimPrefix(line, "RCPT TO:"))
			s.mu.Unlock()
			if _, err := io.WriteString(conn, "250 OK\r\n"); err != nil {
				s.done <- err
				return
			}
		case line == "RSET":
			s.mu.Lock()
			s.rsetCount++
			s.mu.Unlock()
			if _, err := io.WriteString(conn, "250 OK\r\n"); err != nil {
				s.done <- err
				return
			}
		case line == "DATA":
			s.once.Do(func() { close(s.dataStarted) })
			if s.hangOnData {
				_, err := reader.ReadString('\n')
				if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
					s.done <- nil
					return
				}
				s.done <- err
				return
			}
			if _, err := io.WriteString(conn, "354 End data with <CR><LF>.<CR><LF>\r\n"); err != nil {
				s.done <- err
				return
			}
			var data strings.Builder
			for {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					s.done <- err
					return
				}
				if strings.TrimRight(dataLine, "\r\n") == "." {
					break
				}
				data.WriteString(dataLine)
			}
			s.mu.Lock()
			s.data = data.String()
			s.mu.Unlock()
			if _, err := io.WriteString(conn, "250 OK queued\r\n"); err != nil {
				s.done <- err
				return
			}
		case line == "QUIT":
			_, err := io.WriteString(conn, s.quitResponse+"\r\n")
			s.done <- err
			return
		default:
			s.done <- fmt.Errorf("unexpected SMTP command %q", line)
			return
		}
	}
}

func (s *fakeSMTPServer) ehloResponse() string {
	var builder strings.Builder
	builder.WriteString("250-fake.smtp\r\n")
	if s.startTLS && !s.TLSActive() {
		builder.WriteString("250-STARTTLS\r\n")
	}
	if s.authMechanism != "" {
		builder.WriteString("250-AUTH " + s.authMechanism + "\r\n")
	}
	builder.WriteString("250 OK\r\n")
	return builder.String()
}

func (s *fakeSMTPServer) handleAuth(conn net.Conn, line string) error {
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[1] != s.authMechanism {
		_, err := io.WriteString(conn, "504 unsupported auth\r\n")
		return err
	}
	if !s.authOK {
		_, err := io.WriteString(conn, "535 auth failed\r\n")
		return err
	}
	initial := ""
	if len(fields) > 2 {
		decoded, err := base64.StdEncoding.DecodeString(fields[2])
		if err != nil {
			return err
		}
		initial = string(decoded)
	}
	if initial != s.authInitial {
		_, err := io.WriteString(conn, "535 auth failed\r\n")
		return err
	}
	s.mu.Lock()
	s.authenticated = true
	s.mu.Unlock()
	_, err := io.WriteString(conn, "235 authenticated\r\n")
	return err
}

func (s *fakeSMTPServer) setTLSActive() {
	s.mu.Lock()
	s.tlsActive = true
	s.mu.Unlock()
}

type testSMTPAuth struct {
	mechanism string
	initial   []byte
}

func (a testSMTPAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return a.mechanism, a.initial, nil
}

func (a testSMTPAuth) Next([]byte, bool) ([]byte, error) { return nil, nil }

type testCertificate struct {
	cert tls.Certificate
	pool *x509.CertPool
}

func (c *testCertificate) serverConfig() *tls.Config {
	return &tls.Config{Certificates: []tls.Certificate{c.cert}, MinVersion: tls.VersionTLS12}
}

func newTestCertificate(t *testing.T) *testCertificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey() error = %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair() error = %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate() error = %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(parsed)
	return &testCertificate{cert: cert, pool: pool}
}

func strconvAtoi(value string) (int, error) {
	var n int
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid digit %q", r)
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
