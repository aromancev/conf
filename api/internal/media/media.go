package media

import (
	"os"
	"os/exec"
)

const (
	media  = "../../../.artifacts/media"
	input  = media + "/vp9.webm"
	output = media + "/testfest.mpd"
)

func Manifest() error {
	cmd := exec.Command("ffmpeg", "-f", "webm_dash_manifest", "-i", "../../../.artifacts/media/vp9.webm", "-c", "copy", "-map", "0", "-f", "webm_dash_manifest", "-adaptation_sets", "id=0,streams=0", "../../../.artifacts/media/testfest.mpd")
	cmd.Stderr = os.Stdout
	return cmd.Run()
}
