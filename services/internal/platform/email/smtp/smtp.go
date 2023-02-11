package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/aromancev/confa/internal/platform/email"
)

const (
	sendTimeout = time.Minute
)

type Sender struct {
	server   string
	port     string
	password string
	secure   bool
}

func NewSender(server, port, password string, secure bool) *Sender {
	return &Sender{server: server, port: port, password: password, secure: secure}
}

func (s *Sender) Send(ctx context.Context, emails ...email.Email) error {
	dialer := net.Dialer{
		Timeout: sendTimeout,
	}
	conn, err := dialer.DialContext(ctx, "tcp", s.server+":"+s.port)
	if err != nil {
		return err
	}
	defer conn.Close()

	var client *smtp.Client
	if s.secure {
		client, err = s.secureClient(conn, emails[0].From.Email)
	} else {
		client, err = s.insecureClient(conn)
	}
	if err != nil {
		return err
	}
	for _, m := range emails {
		if err = s.send(client, m); err != nil {
			_ = client.Quit()
			return err
		}
	}
	return nil
}

func (s *Sender) secureClient(conn net.Conn, fromEmail string) (*smtp.Client, error) {
	c := tls.Client(conn, &tls.Config{
		ServerName: s.server,
		MinVersion: tls.VersionTLS12,
	})
	err := c.Handshake()
	if err != nil {
		return nil, err
	}
	client, err := smtp.NewClient(c, s.server)
	if err != nil {
		return nil, err
	}
	auth := smtp.PlainAuth("", fromEmail, s.password, s.server)
	if err := client.Auth(auth); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Sender) insecureClient(conn net.Conn) (*smtp.Client, error) {
	return smtp.NewClient(conn, s.server)
}

func (s *Sender) send(client *smtp.Client, msg email.Email) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	if err := client.Mail(msg.From.Email); err != nil {
		return err
	}

	for _, to := range msg.To {
		if err := client.Rcpt(to.Email); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	return writeEmail(msg, w)
}

func writeEmail(msg email.Email, w io.Writer) error {
	writeHeader := func(w io.Writer, key, value string) {
		_, _ = fmt.Fprintf(w, "%s: %s\r\n", key, value)
	}

	from := mail.Address{Name: msg.From.Name, Address: msg.From.Email}
	b := &bytes.Buffer{}
	writeHeader(b, "From", from.String())
	for _, to := range msg.To {
		addr := mail.Address{Name: to.Name, Address: to.Email}
		writeHeader(b, "To", addr.String())
	}
	writeHeader(b, "Subject", msg.Subject)
	writeHeader(b, "MIME-Version", "1.0")
	writeHeader(b, "Content-Type", "text/html; charset=UTF-8")
	_, _ = fmt.Fprint(b, "\r\n")
	_, _ = fmt.Fprint(b, msg.HTML+"\r\n")

	_, err := w.Write(b.Bytes())
	return err
}
