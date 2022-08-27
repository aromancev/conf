package webm

import (
	"io"
	"time"

	"github.com/at-wat/ebml-go/webm"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

type VideoRTPWriter struct {
	writer      io.WriteCloser
	blockWriter webm.BlockWriteCloser
	builder     *samplebuilder.SampleBuilder
	timestamp   time.Duration
}

func NewVideoRTPWriter(writer io.WriteCloser, maxLate uint16) (*VideoRTPWriter, error) {
	return &VideoRTPWriter{
		writer:  writer,
		builder: samplebuilder.New(maxLate, &codecs.VP8Packet{}, 90000),
	}, nil
}

func (w *VideoRTPWriter) Duration() time.Duration {
	return w.timestamp
}

// WriteRTP is UNSAFE to call concurrently.
func (w *VideoRTPWriter) WriteRTP(packet *rtp.Packet) error {
	w.builder.Push(packet)

	for {
		sample := w.builder.Pop()
		if sample == nil {
			return nil
		}

		// Read VP8 header.
		isKeyframe := (sample.Data[0]&0x1 == 0)
		if isKeyframe && w.blockWriter == nil {
			// Keyframe has frame information.
			raw := uint(sample.Data[6]) | uint(sample.Data[7])<<8 | uint(sample.Data[8])<<16 | uint(sample.Data[9])<<24
			width := int(raw & 0x3FFF)
			height := int((raw >> 16) & 0x3FFF)
			// Initialize block writer using received frame size.
			bw, err := webm.NewSimpleBlockWriter(
				w.writer,
				[]webm.TrackEntry{
					{
						Name:        "Video",
						TrackNumber: 1,
						CodecID:     "V_VP8",
						TrackType:   1,
						Video: &webm.Video{
							PixelWidth:  uint64(width),
							PixelHeight: uint64(height),
						},
					},
				},
			)
			if err != nil {
				return err
			}
			w.blockWriter = bw[0]
		}
		if w.blockWriter != nil {
			w.timestamp += sample.Duration
			_, err := w.blockWriter.Write(isKeyframe, int64(w.timestamp/time.Millisecond), sample.Data)
			if err != nil {
				return err
			}
		}
	}
}

func (w *VideoRTPWriter) Close() error {
	if w.blockWriter != nil {
		_ = w.blockWriter.Close()
	}
	return w.writer.Close()
}

type AudioRTPWriter struct {
	writer      io.WriteCloser
	blockWriter webm.BlockWriteCloser
	builder     *samplebuilder.SampleBuilder
	timestamp   time.Duration
}

func NewAudioRTPWriter(writer io.WriteCloser) (*AudioRTPWriter, error) {
	blockWriter, err := webm.NewSimpleBlockWriter(
		writer,
		[]webm.TrackEntry{
			{
				Name:        "Audio",
				TrackNumber: 1,
				CodecID:     "A_OPUS",
				TrackType:   2,
				Audio: &webm.Audio{
					SamplingFrequency: 48000.0,
					Channels:          2,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &AudioRTPWriter{
		writer:      writer,
		builder:     samplebuilder.New(10, &codecs.OpusPacket{}, 48000),
		blockWriter: blockWriter[0],
	}, nil
}

func (w *AudioRTPWriter) Duration() time.Duration {
	return w.timestamp
}

// WriteRTP is UNSAFE to call concurrently.
func (w *AudioRTPWriter) WriteRTP(packet *rtp.Packet) error {
	w.builder.Push(packet)

	for {
		sample := w.builder.Pop()
		if sample == nil {
			return nil
		}

		w.timestamp += sample.Duration
		if _, err := w.blockWriter.Write(true, int64(w.timestamp/time.Millisecond), sample.Data); err != nil {
			return err
		}
	}
}

func (w *AudioRTPWriter) Close() error {
	_ = w.blockWriter.Close()
	return w.writer.Close()
}
