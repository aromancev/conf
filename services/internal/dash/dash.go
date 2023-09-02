package dash

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/aromancev/confa/internal/platform/ffmpeg"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Record struct {
	ID         uuid.UUID
	BucketName string
	ObjectName string
	Duration   time.Duration
}

type Converter struct {
	s3client *s3.Client
	bucket   string
}

func NewConverter(s3client *s3.Client, destBucket string) *Converter {
	return &Converter{
		s3client: s3client,
		bucket:   destBucket,
	}
}

func (c *Converter) ConvertVideo(ctx context.Context, roomID uuid.UUID, record Record) error {
	tmpPath := path.Join("/tmp", "avp", roomID.String(), record.ID.String())
	recordPath := path.Join(tmpPath, "record")
	videoPath := path.Join(tmpPath, "video")
	manifestPath := path.Join(tmpPath, "manifest")

	// Downloading record from storage into a tmp directory.
	err := c.downloadToFile(ctx, recordPath, record.BucketName, record.ObjectName)
	if err != nil {
		return err
	}

	// Write DASH compatible video and manifest files.
	err = ffmpeg.WriteDashVideo(
		ctx,
		ffmpeg.SourceVideo{
			Path:     recordPath,
			Duration: record.Duration,
		},
		ffmpeg.DestinationVideo{
			Path: videoPath,
			FPS:  30,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to write video file: %w", err)
	}
	err = ffmpeg.WriteDashManifest(ctx, videoPath, manifestPath)
	if err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	// Upload DASH files to storage.
	err = c.uploadFromFile(ctx, videoPath, c.bucket, path.Join(roomID.String(), record.ID.String(), "video"), "video/webm")
	if err != nil {
		return err
	}
	err = c.uploadFromFile(ctx, manifestPath, c.bucket, path.Join(roomID.String(), record.ID.String(), "manifest"), "application/dash+xml")
	if err != nil {
		return err
	}

	// Remove all tmp files and the record object.
	if err := os.RemoveAll(tmpPath); err != nil {
		return fmt.Errorf("failed to drop tmp folder: %w", err)
	}
	_, err = c.s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &record.BucketName,
		Key:    &record.ObjectName,
	})
	if err != nil {
		return fmt.Errorf("failed to remove record object: %w", err)
	}
	return nil
}

func (c *Converter) ConvertAudio(ctx context.Context, roomID uuid.UUID, record Record) error {
	tmpPath := path.Join("/tmp", "avp", roomID.String(), record.ID.String())
	recordPath := path.Join(tmpPath, "record")
	audioPath := path.Join(tmpPath, "audio")
	manifestPath := path.Join(tmpPath, "manifest")

	// Downloading record from storage into a tmp directory.
	err := c.downloadToFile(ctx, recordPath, record.BucketName, record.ObjectName)
	if err != nil {
		return err
	}

	// Write DASH compatible webm and manifest files.
	err = ffmpeg.WriteDashAudio(
		ctx,
		ffmpeg.SourceAudio{
			Path:     recordPath,
			Duration: record.Duration,
		},
		ffmpeg.DestinationAudio{
			Path: audioPath,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to write audio file: %w", err)
	}
	err = ffmpeg.WriteDashManifest(ctx, audioPath, manifestPath)
	if err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	// Upload DASH files to storage.
	err = c.uploadFromFile(ctx, audioPath, c.bucket, path.Join(roomID.String(), record.ID.String(), "audio"), "audio/webm")
	if err != nil {
		return err
	}
	err = c.uploadFromFile(ctx, manifestPath, c.bucket, path.Join(roomID.String(), record.ID.String(), "manifest"), "application/xhtml+xml")
	if err != nil {
		return err
	}

	// Remove all tmp files and the record object.
	if err := os.RemoveAll(tmpPath); err != nil {
		return fmt.Errorf("failed to drop tmp folder: %w", err)
	}
	_, err = c.s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &record.BucketName,
		Key:    &record.ObjectName,
	})
	if err != nil {
		return fmt.Errorf("failed to remove record object: %w", err)
	}
	return nil
}

func (c *Converter) downloadToFile(ctx context.Context, fileName, bucketName, objectName string) error {
	if err := os.MkdirAll(path.Dir(fileName), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create a tmp directory: %w", err)
	}
	record, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create a record file: %w", err)
	}
	defer record.Close()
	object, err := c.s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectName,
	})
	if err != nil {
		return fmt.Errorf("failed to access storage: %w", err)
	}
	defer object.Body.Close()
	_, err = io.Copy(record, object.Body)
	if err != nil {
		return fmt.Errorf("failed to download record from storage: %w", err)
	}
	return nil
}

func (c *Converter) uploadFromFile(ctx context.Context, fileName, bucketName, objectName, contentType string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open webm file: %w", err)
	}
	defer file.Close()
	uploader := manager.NewUploader(c.s3client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 1             // default is 5
	})
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &objectName,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload webm file to storage: %w", err)
	}
	return nil
}
