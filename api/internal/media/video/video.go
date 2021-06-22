package video

import (
	"fmt"
	"os"
	"path"

	"github.com/aromancev/confa/internal/platform/mkv"
)

type Converter struct {
	mediaDir string
}

func NewConverter(mediaDir string) *Converter {
	return &Converter{mediaDir: mediaDir}
}

func (p *Converter) Convert(mediaID string) error {
	err := mkv.Convert(
		path.Join(p.mediaDir, mediaID, "raw.ivf"),
		path.Join(p.mediaDir, mediaID, "stream.webm"),
	)
	if err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	err = mkv.CreateManifest(
		path.Join(p.mediaDir, mediaID, "stream.webm"),
		path.Join(p.mediaDir, mediaID, "manifest.mpd"),
	)
	if err != nil {
		return fmt.Errorf("failed create manifest: %w", err)
	}

	_ = os.Remove(path.Join(p.mediaDir, mediaID, "raw.ivf"))
	return nil
}
