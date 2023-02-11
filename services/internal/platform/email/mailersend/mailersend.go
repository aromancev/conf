package mailersend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aromancev/confa/internal/platform/email"
)

const (
	sendTimeout = 30 * time.Second
)

type Sender struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewSender(client *http.Client, baseURL, token string) *Sender {
	return &Sender{
		baseURL: baseURL,
		token:   token,
		client:  client,
	}
}

func (s *Sender) Send(ctx context.Context, messages ...email.Email) error {
	if len(messages) == 0 {
		return errors.New("must send at leas one message")
	}

	if len(messages) == 1 {
		return s.sendOne(ctx, messages[0])
	}
	return s.sendMany(ctx, messages)
}

func (s *Sender) sendOne(ctx context.Context, message email.Email) error {
	buf, err := json.Marshal(newRequest(message))
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	resp, err := s.do(ctx, http.MethodPost, "/v1/email", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("received unexpected status: %s", resp.Status)
	}
	return nil
}

func (s *Sender) sendMany(ctx context.Context, messages []email.Email) error {
	requests := make([]request, len(messages))
	for i, m := range messages {
		requests[i] = newRequest(m)
	}
	buf, err := json.Marshal(requests)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	resp, err := s.do(ctx, http.MethodPost, "/v1/bulk-email", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("received unexpected status: %s", resp.Status)
	}
	return nil
}

func (s *Sender) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	fullPath, err := url.JoinPath(s.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, fullPath, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}

func newRequest(m email.Email) request {
	to := make([]address, len(m.To))
	for i, addr := range m.To {
		to[i] = address{
			Email: addr.Email,
			Name:  addr.Name,
		}
	}
	return request{
		From: address{
			Email: m.From.Email,
			Name:  m.From.Name,
		},
		To:      to,
		Subject: m.Subject,
		HTML:    m.HTML,
		Text:    m.Text,
	}
}

type request struct {
	From    address   `json:"from"`
	To      []address `json:"to"`
	Subject string    `json:"subject"`
	Text    string    `json:"text,omitempty"`
	HTML    string    `json:"html,omitempty"`
}

type address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}
