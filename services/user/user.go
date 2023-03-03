package user

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrDuplicateEntry   = errors.New("duplicate entry")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrValidation       = errors.New("validation error")
)

type User struct {
	ID           uuid.UUID `bson:"_id"`
	CreatedAt    time.Time `bson:"createdAt"`
	Idents       []Ident   `bson:"idents"`
	PasswordHash []byte    `bson:"passwordHash,omitempty"`
}

func (u User) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if len(u.Idents) == 0 {
		return errors.New("must have at least one identifier")
	}
	if len(u.Idents) > 10 {
		return errors.New("maximum 10 identifiers")
	}
	idents := make(map[string]struct{}, len(u.Idents))
	for _, ident := range u.Idents {
		if err := ident.Validate(); err != nil {
			return fmt.Errorf("identifier is not valid: %w", err)
		}
		idx := string(ident.Platform) + ident.Value
		if _, ok := idents[idx]; ok {
			return fmt.Errorf("identifier is not valid: %w", ErrDuplicateEntry)
		}
		idents[idx] = struct{}{}
	}
	if len(u.PasswordHash) > 256 {
		return errors.New("password hash cannot be longer than 256 bytes")
	}
	return nil
}

type Password string

var validPassword = regexp.MustCompile(`^[^ \t].{6,64}[^ \t]$`)

func (p Password) Validate() error {
	if !validPassword.MatchString(string(p)) {
		return fmt.Errorf("invalid password")
	}
	return nil
}

func (p Password) Hash() ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), 12)
	return hash, err
}

func (p Password) Check(hash []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, []byte(p))
	switch {
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func SimulatePasswordCheck() {
	// Random hashed password with cost of 12.
	const randomHash = "$2a$12$EGuLxP/s4jCu5hUwxVPR../87RQJlmltasRFqREp3ZCLHCihA9LIq"
	_ = bcrypt.CompareHashAndPassword([]byte(randomHash), []byte(randomHash))
}

type Platform string

const (
	PlatformUnknown Platform = ""
	PlatformEmail   Platform = "email"
	PlatformTwitter Platform = "twitter"
	PlatformGithub  Platform = "github"
)

func (p Platform) Validate() error {
	switch p {
	case PlatformEmail, PlatformTwitter, PlatformGithub:
		return nil
	}
	return errors.New("unknown platform")
}

type Ident struct {
	Platform Platform `bson:"platform"`
	Value    string   `bson:"value"`
}

var twitterHandle = regexp.MustCompile(`^@?(\w){1,15}$`)
var githubHandle = regexp.MustCompile(`^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$`)

func (i Ident) Validate() error {
	if err := i.Platform.Validate(); err != nil {
		return err
	}
	if i.Value == "" {
		return errors.New("value should not be empty")
	}
	switch i.Platform {
	case PlatformEmail:
		if err := email.Validate(i.Value); err != nil {
			return errors.New("ivalid email identifier")
		}
	case PlatformTwitter:
		if !twitterHandle.MatchString(i.Value) {
			return errors.New("ivalid Twitter identifier")
		}
	case PlatformGithub:
		if len(i.Value) > 40 || !githubHandle.MatchString(i.Value) {
			return errors.New("ivalid GitHub identifier")
		}
	}
	return nil
}

func (i Ident) Normalized() Ident {
	switch i.Platform {
	case PlatformEmail, PlatformTwitter, PlatformGithub:
		i.Value = strings.ToLower(i.Value)
	}
	return i
}

type Update struct {
	PasswordHash []byte `bson:"passwordHash,omitempty"`
}

func (u Update) Validate() error {
	if len(u.PasswordHash) == 0 {
		return errors.New("no fields provided")
	}
	if len(u.PasswordHash) > 256 {
		return errors.New("password hash cannot be longer than 256 bytes")
	}
	return nil
}

type UpdateResult struct {
	Updated int64
}

type Lookup struct {
	ID              uuid.UUID
	Idents          []Ident
	WithoutPassword bool
	Limit           int64
}

func (l Lookup) Validate() error {
	switch {
	case len(l.Idents) > 0:
		for _, ident := range l.Idents {
			if err := ident.Validate(); err != nil {
				return fmt.Errorf("invalid ident: %w", err)
			}
		}
	case l.ID != uuid.Nil:
	default:
		return errors.New("empty lookup not allowed")
	}
	return nil
}
