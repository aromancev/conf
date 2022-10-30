package email

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/aromancev/confa/internal/platform/email"
)

//go:embed *.html
var templates embed.FS

var loginViaEmail *template.Template

func init() {
	loginViaEmail = template.Must(template.ParseFS(templates, "login_via_email.html"))
}

func newLoginViaEmail(to, secretLoginURL string) (email.Email, error) {
	var html bytes.Buffer
	err := loginViaEmail.Execute(&html, map[string]string{
		"secretLoginURL": secretLoginURL,
	})
	if err != nil {
		return email.Email{}, err
	}
	return email.Email{
		FromName:  "Confa",
		Subject:   "Login",
		ToAddress: to,
		HTML:      html.String(),
	}, nil
}
