package mkv

import (
	"os"
	"os/exec"
)

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

func Convert(in, out string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", in,
		"-y",
		"-hide_banner",
		"-c:v", "libvpx-vp9",
		"-row-mt", "1",
		"-keyint_min", "150",
		"-g", "150",
		"-tile-columns", "4",
		"-frame-parallel", "1",
		"-movflags", "faststart",
		"-f", "webm",
		"-dash", "1",
		"-speed", "3",
		"-threads", "4",
		"-an",
		"-vf",
		"scale=426:240",
		"-b:v", "400k",
		"-r", "30",
		"-dash", "1",
		out,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CreateManifest(in, out string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-f", "webm_dash_manifest",
		"-i", in,
		"-y",
		"-c", "copy",
		"-map", "0",
		"-f", "webm_dash_manifest",
		"-adaptation_sets", "id=0,streams=0",
		out,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
