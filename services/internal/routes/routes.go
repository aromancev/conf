package routes

import (
	"net/url"

	"github.com/google/uuid"
)

type LoginAction string

const (
	ActionLogin          LoginAction = "login"
	ActionCreatePassword LoginAction = "create-password"
	ActionResetPassword  LoginAction = "reset-password"
)

type Pages struct {
	base string
}

func NewPages(base string) *Pages {
	return &Pages{
		base: base,
	}
}

func (r *Pages) Login(action LoginAction, token string) string {
	path, _ := url.JoinPath(r.base, "acc", "login")
	vals := url.Values{}
	vals.Set("action", string(action))
	vals.Set("token", token)
	u := url.URL{
		Path:     path,
		RawQuery: vals.Encode(),
	}
	return u.String()
}

func (r *Pages) Confa(confaHandle string) string {
	path, _ := url.JoinPath(r.base, confaHandle)
	return path
}

func (r *Pages) Talk(confaHandle, talkHandle string) string {
	path, _ := url.JoinPath(r.base, confaHandle, talkHandle)
	return path
}

type Buckets struct {
	UserPublic string
}

type Storage struct {
	base    string
	buckets Buckets
}

func NewStorage(base string, buckets Buckets) *Storage {
	return &Storage{
		base:    base,
		buckets: buckets,
	}
}

func (s *Storage) ProfileAvatar(userID, avatarID uuid.UUID) string {
	path, _ := url.JoinPath(s.base, s.buckets.UserPublic, userID.String(), avatarID.String())
	return path
}
