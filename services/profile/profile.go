package profile

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid profile")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicateEntry   = errors.New("profile already exists")
)

type Profile struct {
	ID              uuid.UUID `bson:"_id"`
	Owner           uuid.UUID `bson:"ownerId"`
	Handle          string    `bson:"handle"`
	DisplayName     string    `bson:"displayName,omitempty"`
	AvatarThumbnail Image     `bson:"avatarThumbnail,omitempty"`
	CreatedAt       time.Time `bson:"createdAt"`
}

type Image struct {
	Format string `bson:"format"`
	Data   []byte `bson:"data"`
}

func (i Image) Validate() error {
	if len(i.Data) > 0 {
		switch i.Format {
		case "jpeg", "png", "webp":
		default:
			return errors.New("unsupported image format")
		}
	} else if i.Format != "" {
		return errors.New("must specify format if not empty")
	}
	return nil
}

func (i Image) IsEmpty() bool {
	return len(i.Data) == 0
}

var validHandle = regexp.MustCompile("^[a-z0-9-]{4,64}$")
var validDisplayName = regexp.MustCompile("^[a-zA-Z- ]{0,64}$")

func (p Profile) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if p.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if p.Handle != "" && !validHandle.MatchString(p.Handle) {
		return errors.New("invalid handle")
	}
	if !validDisplayName.MatchString(p.DisplayName) {
		return errors.New("invalid display name")
	}
	if err := p.AvatarThumbnail.Validate(); err != nil {
		return fmt.Errorf("invalid avatar thumbnail: %w", err)
	}
	if len(p.AvatarThumbnail.Data) > 20*1000 {
		return errors.New("avatar thumbnail size should not exceed 20KB")
	}
	return nil
}

type Lookup struct {
	ID     uuid.UUID
	Owners []uuid.UUID
	Handle string
	Limit  int64
	From   Cursor
}

type Cursor struct {
	ID uuid.UUID
}

func (l Lookup) Validate() error {
	if len(l.Owners) > batchLimit {
		return errors.New("too many owners")
	}
	return nil
}

type AvatarSourceStorage struct {
	Bucket string
	Path   string
}

func (s AvatarSourceStorage) Validate() error {
	if s.Bucket == "" {
		return errors.New("bucket should not be empty")
	}
	if s.Path == "" {
		return errors.New("path should not be empty")
	}
	return nil
}

type AvatarSourcePublicURL struct {
	URL string
}

func (s AvatarSourcePublicURL) Validate() error {
	if _, err := url.ParseRequestURI(s.URL); err != nil {
		return errors.New("invalid url")
	}
	return nil
}

type AvatarSource struct {
	Storage   *AvatarSourceStorage
	PublicURL *AvatarSourcePublicURL
}

func (s AvatarSource) Validate() error {
	isEmpty := true
	if s.Storage != nil {
		if err := s.Storage.Validate(); err != nil {
			return fmt.Errorf("invalid storage source: %w", err)
		}
		isEmpty = false
	}
	if s.PublicURL != nil {
		if err := s.PublicURL.Validate(); err != nil {
			return fmt.Errorf("invalid public url source: %w", err)
		}
		isEmpty = false
	}
	if isEmpty {
		return errors.New("should have at least one source")
	}
	return nil
}
