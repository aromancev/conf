package profile

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type Repo interface {
	CreateOrUpdate(ctx context.Context, request Profile) (Profile, error)
}

type Emitter interface {
	UpdateProfile(ctx context.Context, userID uuid.UUID, source AvatarSource) error
}

type Buckets struct {
	UserUploads string
	UserPublic  string
}

type Updater struct {
	storage *minio.Client
	buckets Buckets
	baseURL string
	emitter Emitter
	repo    Repo
}

func NewUpdater(baseURL string, buckets Buckets, storage *minio.Client, emitter Emitter, repo Repo) *Updater {
	return &Updater{
		storage: storage,
		buckets: buckets,
		baseURL: baseURL,
		emitter: emitter,
		repo:    repo,
	}
}

func (u *Updater) RequestUpload(ctx context.Context, userID uuid.UUID) (string, map[string]string, error) {
	objectID := "avatar"
	objectPath := path.Join(userID.String(), objectID)
	policy := minio.NewPostPolicy()
	bucket := u.buckets.UserUploads
	if err := policy.SetBucket(bucket); err != nil {
		return "", nil, fmt.Errorf("failed to set policy bucket: %w", err)
	}
	if err := policy.SetKey(objectPath); err != nil {
		return "", nil, fmt.Errorf("failed to set policy key: %w", err)
	}
	if err := policy.SetExpires(time.Now().Add(30 * time.Second)); err != nil {
		return "", nil, fmt.Errorf("failed to set policy expires: %w", err)
	}
	if err := policy.SetContentLengthRange(0, 5*1000*1000); err != nil {
		return "", nil, fmt.Errorf("failed to set policy content length: %w", err)
	}
	_, data, err := u.storage.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign post policy: %w", err)
	}

	err = u.emitter.UpdateProfile(ctx, userID, AvatarSource{
		Storage: &AvatarSourceStorage{
			Bucket: bucket,
			Path:   objectPath,
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to emit avatar update: %w", err)
	}

	return path.Join(u.baseURL, bucket), data, nil
}

func (u *Updater) Update(ctx context.Context, userID uuid.UUID, source AvatarSource) error {
	const avatarFullSize = 460
	const avatarThumbnailSize = 128
	const quality = 50 // Ranges from 1 to 100 inclusive, higher is better.

	if err := source.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err)
	}

	cleanup := func() error {
		// Removing the temporary object if the source is storage.
		if source.Storage != nil {
			err := u.storage.RemoveObject(ctx, source.Storage.Bucket, source.Storage.Path, minio.RemoveObjectOptions{})
			if err != nil {
				return fmt.Errorf("failed to remove temporary object: %w", err)
			}
		}
		return nil
	}

	// Fetching the original image.
	var sourceImage image.Image
	switch {
	case source.Storage != nil:
		object, err := u.storage.GetObject(ctx, source.Storage.Bucket, source.Storage.Path, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to get object: %w", err)
		}
		buff, err := io.ReadAll(object)
		switch {
		case minio.ToErrorResponse(err).StatusCode == http.StatusNotFound:
			return ErrNotFound
		case err != nil:
			return fmt.Errorf("failed to read object: %w", err)
		}
		sourceImage, _, err = image.Decode(bytes.NewReader(buff))
		if err != nil {
			if err := cleanup(); err != nil {
				return err
			}
			return fmt.Errorf("%w: %s", ErrValidation, err)
		}
	default:
		return fmt.Errorf("%w: no image source", ErrValidation)
	}

	// Cropping thumbnail and full size.
	fullSize := imaging.Fill(sourceImage, avatarFullSize, avatarFullSize, imaging.Center, imaging.Box)
	thumbnail := imaging.Fill(fullSize, avatarThumbnailSize, avatarThumbnailSize, imaging.Center, imaging.Box)

	// Updating the thumbnail in the profile.
	buffer := bytes.NewBuffer(nil)
	err := imaging.Encode(buffer, thumbnail, imaging.JPEG, imaging.JPEGQuality(quality))
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	prof, err := u.repo.CreateOrUpdate(ctx, Profile{
		ID:    uuid.New(),
		Owner: userID,
		AvatarThumbnail: Image{
			Format: "jpeg",
			Data:   buffer.Bytes(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert profile: %w", err)
	}

	// Uploading fullsize avatar to the storage.
	buffer.Reset()
	err = imaging.Encode(buffer, fullSize, imaging.JPEG, imaging.JPEGQuality(quality))
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	_, err = u.storage.PutObject(ctx, u.buckets.UserPublic, path.Join(userID.String(), prof.ID.String()), buffer, int64(buffer.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return fmt.Errorf("failed to upload converted avatar: %w", err)
	}
	return cleanup()
}
