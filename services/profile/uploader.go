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

	"github.com/aromancev/confa/internal/routes"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
)

type Repo interface {
	CreateOrUpdate(ctx context.Context, request Profile) (Profile, error)
}

type Emitter interface {
	UpdateProfile(ctx context.Context, userID uuid.UUID, gavenName, familyName *string, thumbnail, avatar FileSource) error
}

type Buckets struct {
	UserUploads string
	UserPublic  string
}

type Updater struct {
	storageRoutes *routes.Storage
	storage       *minio.Client
	buckets       Buckets
	emitter       Emitter
	repo          Repo
	client        *http.Client
}

func NewUpdater(storageRoutes *routes.Storage, buckets Buckets, storage *minio.Client, emitter Emitter, repo Repo, client *http.Client) *Updater {
	return &Updater{
		storageRoutes: storageRoutes,
		storage:       storage,
		buckets:       buckets,
		emitter:       emitter,
		repo:          repo,
		client:        client,
	}
}

func (u *Updater) UpdateAndRequestUpload(ctx context.Context, userID uuid.UUID) (string, map[string]string, error) {
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

	err = u.emitter.UpdateProfile(ctx, userID, nil, nil, FileSource{}, FileSource{
		Storage: &FileSourceStorage{
			Bucket: bucket,
			Path:   objectPath,
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to emit avatar update: %w", err)
	}

	return u.storageRoutes.Bucket(bucket), data, nil
}

func (u *Updater) Update(ctx context.Context, userID uuid.UUID, givenName, familyName *string, thumbnail, avatar FileSource) error {
	const avatarFullSize = 460
	const thumbnailSize = 128
	const quality = 50 // Ranges from 1 to 100 inclusive, higher is better.

	if givenName == nil && familyName == nil && thumbnail.IsZero() && avatar.IsZero() {
		return fmt.Errorf("%w: %s", ErrValidation, "empty update not allowed")
	}
	if err := thumbnail.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err)
	}
	if err := avatar.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err)
	}

	update := Profile{
		ID:         uuid.New(),
		Owner:      userID,
		GivenName:  givenName,
		FamilyName: familyName,
	}
	avatarImg, err := u.fetchImage(ctx, avatar)
	if err != nil {
		return err
	}
	thumbnailImg, err := u.fetchImage(ctx, thumbnail)
	if err != nil {
		return err
	}
	// Generate thumbnail from avatar.
	if thumbnailImg == nil && avatarImg != nil {
		thumbnailImg = avatarImg
	}

	if avatarImg != nil {
		update.AvatarID = uuid.New()
		cropped := imaging.Fill(avatarImg, avatarFullSize, avatarFullSize, imaging.Center, imaging.Box)
		buffer := bytes.NewBuffer(nil)
		err := imaging.Encode(buffer, cropped, imaging.JPEG, imaging.JPEGQuality(quality))
		if err != nil {
			return fmt.Errorf("failed to encode image: %w", err)
		}
		_, err = u.storage.PutObject(ctx, u.buckets.UserPublic, path.Join(userID.String(), update.AvatarID.String()), buffer, int64(buffer.Len()), minio.PutObjectOptions{
			ContentType: "image/jpeg",
		})
		if err != nil {
			return fmt.Errorf("failed to upload converted avatar: %w", err)
		}
	}
	if thumbnailImg != nil {
		cropped := imaging.Fill(thumbnailImg, thumbnailSize, thumbnailSize, imaging.Center, imaging.Box)
		buffer := bytes.NewBuffer(nil)
		err := imaging.Encode(buffer, cropped, imaging.JPEG, imaging.JPEGQuality(quality))
		if err != nil {
			return fmt.Errorf("failed to encode image: %w", err)
		}
		update.AvatarThumbnail = Image{
			Format: "jpeg",
			Data:   buffer.Bytes(),
		}
	}

	_, err = u.repo.CreateOrUpdate(ctx, update)
	if err != nil {
		return fmt.Errorf("failed to upsert profile: %w", err)
	}
	u.cleanupObject(ctx, avatar)
	u.cleanupObject(ctx, thumbnail)
	return nil
}

func (u *Updater) fetchImage(ctx context.Context, source FileSource) (image.Image, error) {
	var buff []byte
	switch {
	case source.Storage != nil:
		object, err := u.storage.GetObject(ctx, source.Storage.Bucket, source.Storage.Path, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get object: %w", err)
		}
		buff, err = io.ReadAll(object)
		switch {
		case minio.ToErrorResponse(err).StatusCode == http.StatusNotFound:
			return nil, ErrNotFound
		case err != nil:
			return nil, fmt.Errorf("failed to read object: %w", err)
		}
	case source.PublicURL != nil:
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.PublicURL.URL, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := u.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected response code")
		}
		buff, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to download file: %w", err)
		}
	default:
		return nil, nil
	}
	img, _, err := image.Decode(bytes.NewReader(buff))
	if err != nil {
		u.cleanupObject(ctx, source)
		return nil, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	return img, nil
}

func (u *Updater) cleanupObject(ctx context.Context, source FileSource) {
	if source.Storage == nil {
		return
	}
	err := u.storage.RemoveObject(ctx, source.Storage.Bucket, source.Storage.Path, minio.RemoveObjectOptions{})
	if err != nil {
		log.Ctx(ctx).Err(err).Str("bucket", source.Storage.Bucket).Str("path", source.Storage.Path).Msg("Failed to remove temporary object.")
	}
}
