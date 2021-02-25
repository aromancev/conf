package iam

import (
	"net/http"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID
}

func Authenticate(r *http.Request) (User, error) {
	return User{
		ID: uuid.New(),
	}, nil
}
