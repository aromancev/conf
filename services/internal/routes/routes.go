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

func (r *Routes) LoginViaEmail(token string) string {
	return fmt.Sprintf("%s/login?token=%s", r.base, token)
}

func (r *Routes) Confa(confaHandle string) string {
	return fmt.Sprintf("%s/%s", r.base, confaHandle)
}

func (r *Routes) Talk(confaHandle, talkHandle string) string {
	return fmt.Sprintf("%s/%s/%s", r.base, confaHandle, talkHandle)
}
