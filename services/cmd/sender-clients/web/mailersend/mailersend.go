package mailersend

import (
	"encoding/json"
	"net/http"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/julienschmidt/httprouter"
)

type Mailersend struct {
	emails []email.Email
}

func NewMailersend() *Mailersend {
	return &Mailersend{
		emails: []email.Email{},
	}
}

func (m *Mailersend) AddRoutes(r *httprouter.Router) {
	r.POST("/v1/email", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		m.emails = append([]email.Email{newMessage(req)}, m.emails...)
		w.WriteHeader(http.StatusAccepted)
	})

	r.POST("/v1/bulk-email", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var reqs []request
		err := json.NewDecoder(r.Body).Decode(&reqs)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		messages := make([]email.Email, len(reqs))
		for i, r := range reqs {
			messages[i] = newMessage(r)
		}

		m.emails = append(messages, m.emails...)
		w.WriteHeader(http.StatusAccepted)
	})
}

func (m *Mailersend) Emails() []email.Email {
	return m.emails
}

func newMessage(req request) email.Email {
	message := email.Email{
		From: email.Address{
			Email: req.From.Email,
			Name:  req.From.Name,
		},
		Subject: req.Subject,
		Text:    req.Text,
		HTML:    req.HTML,
	}
	message.To = make([]email.Address, len(req.To))
	for i, to := range req.To {
		message.To[i] = email.Address{
			Email: to.Email,
			Name:  to.Name,
		}
	}
	return message
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
