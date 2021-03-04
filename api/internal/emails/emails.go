package emails

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/aromancev/confa/internal/platform/email"
)

//go:embed login.html
var loginT string
var login *template.Template

func init() {
	login = template.Must(template.New("login").Parse(loginT))
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

	msg := email.Email{
		From:      "Confa",
		Subject:   "Login",
		ToAddress: to,
		HTML:      html.String(),
	}
	return msg, nil
}
