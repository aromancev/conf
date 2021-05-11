package media

import (
	"os"
	"os/exec"
)

const (
	media  = "../../../.artifacts/media"
	input  = media + "/a74f4ca2-be5a-4cfe-ae2f-08a46cd4f95c.webm"
	output = media + "/testfest.mpd"
)

func Manifest() error {
	cmd := exec.Command("ffmpeg", "-f", "webm_dash_manifest", "-i", input, "-c", "copy", "-map", "0", "-f", "webm_dash_manifest", "-adaptation_sets", "id=0,streams=0", output)
	cmd.Stderr = os.Stdout
	return cmd.Run()
}
