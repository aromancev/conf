package email

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/aromancev/confa/internal/platform/email"
)

//go:embed *.html
var templates embed.FS

var login, createPassword, resetPassword, talkRecordingReady *template.Template

func init() {
	login = template.Must(template.ParseFS(templates, "login.html"))
	createPassword = template.Must(template.ParseFS(templates, "create-password.html"))
	resetPassword = template.Must(template.ParseFS(templates, "reset-password.html"))
	talkRecordingReady = template.Must(template.ParseFS(templates, "talk-recording-ready.html"))
}

func newLogin(from email.Address, to []email.Address, secretURL string) (email.Email, error) {
	var html bytes.Buffer
	err := login.Execute(&html, map[string]string{
		"secretURL": secretURL,
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

func newCreatePassword(from email.Address, to []email.Address, secretURL string) (email.Email, error) {
	var html bytes.Buffer
	err := createPassword.Execute(&html, map[string]string{
		"secretURL": secretURL,
	})
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{
		From:    from,
		Subject: "Create Password",
		To:      to,
		HTML:    html.String(),
	}, nil
}

func newResetPassword(from email.Address, to []email.Address, secretURL string) (email.Email, error) {
	var html bytes.Buffer
	err := resetPassword.Execute(&html, map[string]string{
		"secretURL": secretURL,
	})
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{
		From:    from,
		Subject: "Reset Password",
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
