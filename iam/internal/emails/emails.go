package emails

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/aromancev/iam/internal/platform/email"
)

//go:embed *.html
var templates embed.FS

var login *template.Template

func init() {
	login = template.Must(template.ParseFS(templates, "login.html"))
}

func Login(baseURL, to, token string) (email.Email, error) {
	var html bytes.Buffer
	err := login.Execute(&html, map[string]string{
		"base_url": baseURL,
		"token":    token,
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
