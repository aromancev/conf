package handler

import (
	"context"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/aromancev/confa/internal/media/video"
	"github.com/aromancev/confa/proto/queue"
)

func processVideo(converter *video.Converter) JobHandle {
	return func(ctx context.Context, j *beanstalk.Job) error {
		var job queue.VideoJob
		err := proto.Unmarshal(j.Body, &job)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal email job")
			if err := j.Delete(ctx); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to delete job")
			}
			return nil
		}

		return converter.Convert(job.MediaId)
	}
}
