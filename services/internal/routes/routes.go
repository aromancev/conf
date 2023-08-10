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
	host   string
	scheme string
}

func NewPages(scheme, host string) *Pages {
	return &Pages{
		scheme: scheme,
		host:   host,
	}
}

func (r *Pages) Login(action LoginAction, token string) string {
	path, _ := url.JoinPath("acc", "login")
	vals := url.Values{}
	vals.Set("action", string(action))
	vals.Set("token", token)
	u := url.URL{
		Host:     r.host,
		Scheme:   r.scheme,
		Path:     path,
		RawQuery: vals.Encode(),
	}
	return u.String()
}

func (r *Pages) Confa(confaHandle string) string {
	u := url.URL{
		Host:   r.host,
		Scheme: r.scheme,
		Path:   confaHandle,
	}
	return u.String()
}

func (r *Pages) Talk(confaHandle, talkHandle string) string {
	path, _ := url.JoinPath(confaHandle, talkHandle)
	u := url.URL{
		Host:   r.host,
		Scheme: r.scheme,
		Path:   path,
	}
	return u.String()
}

type Buckets struct {
	UserPublic string
}

type Storage struct {
	scheme  string
	host    string
	buckets Buckets
}

func NewStorage(scheme, host string, buckets Buckets) *Storage {
	return &Storage{
		scheme:  scheme,
		host:    host,
		buckets: buckets,
	}
}

func (s *Storage) Bucket(name string) string {
	u := url.URL{
		Host:   s.host,
		Scheme: s.scheme,
		Path:   name,
	}
	return u.String()
}

func (s *Storage) ProfileAvatar(userID, avatarID uuid.UUID) string {
	path, _ := url.JoinPath(s.buckets.UserPublic, userID.String(), avatarID.String())
	u := url.URL{
		Host:   s.host,
		Scheme: s.scheme,
		Path:   path,
	}
	return u.String()
}
