// FFMEPEG EXAMPLE DASH CONVERSION
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=426:240 -b:v 400k -r 30 -dash 1 dash/426x240-30-400k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=426:240 -b:v 600k -r 30 -dash 1 dash/426x240-30-600k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=640:360 -b:v 700k -r 30 -dash 1 dash/640x360-30-700k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=640:360 -b:v 900k -r 30 -dash 1 dash/640x360-30-900k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=854:480 -b:v 1250k -r 30 -dash 1 dash/854x480-30-1250k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=854:480 -b:v 1600k -r 30 -dash 1 dash/854x480-30-1600k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1280:720 -b:v 2500k -r 30 -dash 1 dash/1280x720-30-2500k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1280:720 -b:v 3200k -r 30 -dash 1 dash/1280x720-30-3200k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1280:720 -b:v 3500k -r 60 -dash 1 dash/1280x720-60-3500k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1280:720 -b:v 4400k -r 60 -dash 1 dash/1280x720-60-4400k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1920:1080 -b:v 4500k -r 30 -dash 1 dash/1920x1080-30-4500k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1920:1080 -b:v 5300k -r 30 -dash 1 dash/1920x1080-30-5300k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1920:1080 -b:v 5800k -r 60 -dash 1 dash/1920x1080-60-5800k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:v libvpx-vp9 -row-mt 1 -keyint_min 150 -g 150 -tile-columns 4 -frame-parallel 1 \
// -movflags faststart -f webm -dash 1 -speed 3 -threads 4 -an -vf scale=1920:1080 -b:v 7400k -r 60 -dash 1 dash/1920x1080-60-7400k.webm && \
// ffmpeg -hide_banner -i vp9.webm -c:a libvorbis -b:a 192k -vn -f webm -dash 1 dash/audio.webm
//
// ffmpeg \
// -f webm_dash_manifest -i dash/426x240-30-400k.webm \
// -f webm_dash_manifest -i dash/426x240-30-600k.webm \
// -f webm_dash_manifest -i dash/640x360-30-700k.webm \
// -f webm_dash_manifest -i dash/640x360-30-900k.webm \
// -f webm_dash_manifest -i dash/854x480-30-1250k.webm \
// -f webm_dash_manifest -i dash/854x480-30-1600k.webm \
// -f webm_dash_manifest -i dash/1280x720-30-2500k.webm \
// -f webm_dash_manifest -i dash/1280x720-30-3200k.webm \
// -f webm_dash_manifest -i dash/1280x720-60-3500k.webm \
// -f webm_dash_manifest -i dash/1280x720-60-4400k.webm \
// -f webm_dash_manifest -i dash/1920x1080-30-4500k.webm \
// -f webm_dash_manifest -i dash/1920x1080-30-5300k.webm \
// -f webm_dash_manifest -i dash/1920x1080-60-5800k.webm \
// -f webm_dash_manifest -i dash/1920x1080-60-7400k.webm \
// -c copy \
// -map 0 -map 1 -map 2 -map 3 -map 4 -map 5 -map 6 -map 7 -map 8 -map 9 -map 10 -map 11 -map 12 -map 13 \
// -f webm_dash_manifest \
// -adaptation_sets "id=0,streams=0,1,2,3,4,5,6,7,8,9,10,11,12,13" \
// dash/manifest.mpd
package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

type SourceVideo struct {
	Path     string
	Duration time.Duration
}

type DestinationVideo struct {
	Path string
	FPS  uint
}

// WriteDashVideo converts video file into a DASH compatible WebM.
// It cannot operate on streams because ffmpeg performs seeks on input and output. There is no simple way to "stream" dash compatible files.
// More info: http://wiki.webmproject.org/adaptive-streaming/instructions-to-playback-adaptive-webm-using-dash
func WriteDashVideo(ctx context.Context, in SourceVideo, out DestinationVideo) error {
	if in.Duration == 0 {
		return errors.New("source duration should not be zero")
	}
	if out.FPS == 0 {
		out.FPS = 30
	}

	// Minimum number of frames for DASH encoding. The ideal value is 150 frames.
	// If the video is shorter, it will output invalid DASH file.
	// This is why we set it to either the total number of frames in the video or 150.
	minFrames := uint(in.Duration.Seconds()) * out.FPS
	if minFrames > 150 {
		minFrames = 150
	}
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"ffmpeg",
		// Yes to everything.
		"-y",
		"-hide_banner",
		"-i", in.Path,
		// VP9 is expensive to convert, but produces files with the least size and the best quality.
		// Consider VP8 to speed up the conversion.
		"-c:v", "libvpx-vp9",
		// Minimum frames per DASH segment.
		// If less than the number of frames in the resulting video, DASH manifest will fail.
		"-keyint_min", fmt.Sprintf("%d", minFrames),
		// Frames per DASH segment.
		// Has to be the same as keyint_min.
		"-g", fmt.Sprintf("%d", minFrames),
		// Making sure the framerate is constant otherwise, the resulting DASH file will have gaps in it.
		"-vsync", "cfr",
		"-row-mt", "1",
		"-tile-columns", "4",
		"-frame-parallel", "1",
		"-movflags", "faststart",
		// Ignore audio.
		"-an",
		"-f", "webm",
		"-dash", "1",
		"-speed", "3",
		"-threads", fmt.Sprintf("%d", runtime.NumCPU()),
		// Scale video down if it's bigger than the desired resoltuon.
		"-vf", "scale='min(1920,iw)':min'(1080,ih)':force_original_aspect_ratio=decrease",
		// Bitrate.
		"-b:v", "7400k",
		"-r", fmt.Sprintf("%d", out.FPS),
		"-dash", "1",
		out.Path,
	)
	return cmd.Run()
}

type SourceAudio struct {
	Path     string
	Duration time.Duration
}

type DestinationAudio struct {
	Path string
}

// WriteDashAudio converts audio file into a DASH compatible WebM.
// It cannot operate on streams because ffmpeg performs seeks on input and output. There is no simple way to "stream" dash compatible files.
// More info: http://wiki.webmproject.org/adaptive-streaming/instructions-to-playback-adaptive-webm-using-dash
//
// If the file is shorter than 5 seconds, it will pad it with silence to 5 second length.
// We have to do this becuase otherwise it will produce an invalid DASH file and the manifest creation will fail.
func WriteDashAudio(ctx context.Context, in SourceAudio, out DestinationAudio) error {
	if in.Duration == 0 {
		return errors.New("source duration should not be zero")
	}
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"ffmpeg",
		// Yes to everything.
		"-y",
		"-i", in.Path,
		// OPUS is the successor of VORBIS. It is the best codec available for audio.
		"-c:a", "libopus",
		// Bitrate.
		"-b:a", "64k",
		// Audio type.
		// ‘voip’ -Favor improved speech intelligibility.
		// ‘audio’ - Favor faithfulness to the input (the default).
		// ‘lowdelay’ - Restrict to only the lowest delay modes.
		"-application", "voip",
		// Pad audio to 5 seconds with silence.
		// We have to do this becuase otherwise it will produce an invalid DASH file and the manifest creation will fail.
		"-af", "apad='whole_dur=5'",
		// Ignore video.
		"-vn",
		"-f", "webm",
		// Sample rate.
		"-ar", "48000",
		// Number of channels.
		"-ac", "2",
		"-dash", "1",
		out.Path,
	)
	return cmd.Run()
}

// WriteDashManifest creates a DASH manifest from a webm file.
// It cannot operate on streams because ffmpeg performs seeks on input and output. There is no simple way to "stream" dash compatible files.
// More info: http://wiki.webmproject.org/adaptive-streaming/instructions-to-playback-adaptive-webm-using-dash
func WriteDashManifest(ctx context.Context, in, out string) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"ffmpeg",
		"-y",
		"-f", "webm_dash_manifest",
		"-i", in,
		"-c", "copy",
		"-map", "0",
		"-f", "webm_dash_manifest",
		"-adaptation_sets", "id=0,streams=0",
		out,
	)
	return cmd.Run()
}
