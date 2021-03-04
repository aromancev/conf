package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"regexp"
	"time"
)

const (
	sendTimeout = time.Minute
)

type Email struct {
	From      string `json:"from"`
	ToAddress string `json:"toAddress"`
	Subject   string `json:"subject"`
	HTML      string `json:"html"`
}

func (e Email) Validate() error {
	if !emailPattern.MatchString(e.ToAddress) {
		return errors.New("invalid ToAddress")
	}
	return nil
}

func ValidateEmail(email string) error {
	if !emailPattern.MatchString(email) {
		return errors.New("invalid email")
	}
	return nil
}

type Sender struct {
	server      string
	port        string
	fromAddress string
	password    string
}

func NewSender(server, port, fromAddress, password string) *Sender {
	return &Sender{server: server, port: port, fromAddress: fromAddress, password: password}
}

func (s *Sender) Send(ctx context.Context, emails ...Email) (error, []error) {
	dialer := net.Dialer{
		Timeout: sendTimeout,
	}
	tcpConn, err := dialer.DialContext(ctx, "tcp", s.server+":"+s.port)
	if err != nil {
		return err, nil
	}
	defer tcpConn.Close()
	conn := tls.Client(tcpConn, &tls.Config{
		ServerName: s.server,
	})
	err = conn.Handshake()
	if err != nil {
		return err, nil
	}

	client, err := smtp.NewClient(conn, s.server)
	if err != nil {
		return err, nil
	}

	auth := smtp.PlainAuth("", s.fromAddress, s.password, s.server)
	if err = client.Auth(auth); err != nil {
		return err, nil
	}

	var errs []error
	for _, m := range emails {
		if err = s.send(client, m); err != nil {
			_ = client.Reset()
			errs = append(errs, err)
		} else {
			errs = append(errs, nil)
		}
	}

	return client.Quit(), errs
}

var emailPattern = regexp.MustCompile(`^([!#-'*+/-9=?A-Z^-~-]+(\.[!#-'*+/-9=?A-Z^-~-]+)*|"([]!#-[^-~ \t]|(\\[\t -~]))+")@([!#-'*+/-9=?A-Z^-~-]+(\.[!#-'*+/-9=?A-Z^-~-]+)*|\[[\t -Z^-~]*])$`) // nolint: gocritic

func (e Email) write(w io.Writer, fromAddr string) error {
	writeHeader := func(w io.Writer, key, value string) {
		_, _ = fmt.Fprintf(w, "%s: %s\r\n", key, value)
	}

	from := mail.Address{Name: e.From, Address: fromAddr}
	to := mail.Address{Name: "", Address: e.ToAddress}

	b := &bytes.Buffer{}
	writeHeader(b, "From", from.String())
	writeHeader(b, "To", to.String())
	writeHeader(b, "Subject", e.Subject)
	writeHeader(b, "MIME-Version", "1.0")
	writeHeader(b, "Content-Type", "text/html; charset=UTF-8")
	_, _ = fmt.Fprint(b, "\r\n")
	_, _ = fmt.Fprint(b, e.HTML+"\r\n")

	_, err := w.Write(b.Bytes())
	return err
}

func (s *Sender) send(client *smtp.Client, email Email) error {
	if err := email.Validate(); err != nil {
		return err
	}
	if err := client.Mail(s.fromAddress); err != nil {
		return err
	}

	if err := client.Rcpt(email.ToAddress); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	return email.write(w, s.fromAddress)
}
