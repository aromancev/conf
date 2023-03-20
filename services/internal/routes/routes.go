package routes

import (
	"fmt"
)

type Routes struct {
	base string
}

func NewRoutes(base string) *Routes {
	return &Routes{
		base: base,
	}
}

func (r *Routes) LoginWithEmail(token string) string {
	return fmt.Sprintf("%s/acc/login?action=login&token=%s", r.base, token)
}

func (r *Routes) CreatePassword(token string) string {
	return fmt.Sprintf("%s/acc/login?action=create-password&token=%s", r.base, token)
}

func (r *Routes) ResetPassword(token string) string {
	return fmt.Sprintf("%s/acc/login?action=reset-password&token=%s", r.base, token)
}

func (r *Routes) Confa(confaHandle string) string {
	return fmt.Sprintf("%s/%s", r.base, confaHandle)
}

func (r *Routes) Talk(confaHandle, talkHandle string) string {
	return fmt.Sprintf("%s/%s/%s", r.base, confaHandle, talkHandle)
}
