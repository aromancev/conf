package email

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/aromancev/confa/internal/platform/email"
)

//go:embed *.html
var templates embed.FS

var loginViaEmail, talkRecordingReady *template.Template

func init() {
	loginViaEmail = template.Must(template.ParseFS(templates, "login_via_email.html"))
	talkRecordingReady = template.Must(template.ParseFS(templates, "talk_recording_ready.html"))
}

func newLoginViaEmail(from email.Address, to []email.Address, secretLoginURL string) (email.Email, error) {
	var html bytes.Buffer
	err := loginViaEmail.Execute(&html, map[string]string{
		"secretLoginURL": secretLoginURL,
	})
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{
		From:    from,
		Subject: "Login",
		To:      to,
		HTML:    html.String(),
	}, nil
}

func newTalkRecordingReady(from email.Address, to []email.Address, confaURL, confaTitle, talkURL, talkTitle string) (email.Email, error) {
	var html bytes.Buffer
	err := talkRecordingReady.Execute(&html, map[string]string{
		"confaURL":   confaURL,
		"confaTitle": confaTitle,
		"talkURL":    talkURL,
		"talkTitle":  talkTitle,
	})
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{
		From:    from,
		Subject: "Talk Recording is Ready",
		To:      to,
		HTML:    html.String(),
	}, nil
}
