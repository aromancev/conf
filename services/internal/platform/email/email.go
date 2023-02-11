package email

import (
	"errors"
	"fmt"
	"regexp"
)

func Validate(email string) error {
	if !emailPattern.MatchString(email) {
		return errors.New("invalid email address")
	}
	return nil
}

type Email struct {
	From    Address   `json:"from"`
	To      []Address `json:"to"`
	Subject string    `json:"subject"`
	HTML    string    `json:"html"`
	Text    string    `json:"text"`
}

type Address struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (e Email) Validate() error {
	if err := Validate(e.From.Email); err != nil {
		return fmt.Errorf("invalid from: %w", err)
	}
	for _, a := range e.To {
		if err := Validate(a.Email); err != nil {
			return fmt.Errorf("invalid to: %w", err)
		}
	}
	if len(e.To) == 0 {
		return errors.New("to must not be empty")
	}
	return nil
}

var emailPattern = regexp.MustCompile(`^([!#-'*+/-9=?A-Z^-~-]+(\.[!#-'*+/-9=?A-Z^-~-]+)*|"([]!#-[^-~ \t]|(\\[\t -~]))+")@([!#-'*+/-9=?A-Z^-~-]+(\.[!#-'*+/-9=?A-Z^-~-]+)*|\[[\t -Z^-~]*])$`) // nolint: gocritic
